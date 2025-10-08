package service

import (
	"bytes"
	"cloud-storage-file-service/internal/model"

	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
)

type StorageService struct {
	client   *minio.Client
	bucket   string
	fileDAO  model.FileDAO
	partSize int64 // 分片大小（以字节为单位）
}

// NewStorageService 创建一个新的 StorageService 实例
//
// 参数：
//   - client: *minio.Client 类型，指向一个 MinIO 客户端实例
//   - bucket: string 类型，存储桶的名称
//   - dao: model.FileDAO 类型，文件数据访问对象
//
// 返回值：
//   - *StorageService 类型，指向新创建的 StorageService 实例
//
// 备注：
//   - 默认分片大小为5MB，符合MinIO要求
func NewStorageService(client *minio.Client, bucket string, dao model.FileDAO) *StorageService {
	return &StorageService{
		client:  client,
		bucket:  bucket,
		fileDAO: dao,
		// 默认分片大小为5MB，符合MinIO要求
		partSize: 5 * 1024 * 1024,
	}
}

// SetPartSize 设置分片大小
func (s *StorageService) SetPartSize(size int64) {
	if size <= 5*1024*1024 {
		return
	}
	s.partSize = size
}

// GetPartSize 获取分片大小
func (s *StorageService) GetPartSize() int64 {
	return s.partSize
}

// UploadFileDirectly 直接上传文件
func (s *StorageService) UploadFileDirectly(ctx context.Context, fileName string, data []byte, clientMD5 string) (*model.File, error) {
	hash := md5.Sum(data)
	md5Str := hex.EncodeToString(hash[:])
	if md5Str != clientMD5 {
		return nil, fmt.Errorf("数据校验失败")
	}
	objectName := fmt.Sprintf("uploads/%s", fileName)

	info, err := s.client.PutObject(ctx, s.bucket, objectName, bytes.NewReader(data), -1, minio.PutObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %v", err)
	}

	file := &model.File{
		FileName:   fileName,
		Bucket:     s.bucket,
		ObjectName: objectName,
		Size:       info.Size,
		Md5:        info.ETag,
		Status:     1, // 已完成
		CreatedAt:  time.Now(),
	}

	if err := s.fileDAO.CreateFile(file); err != nil {
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}
	return file, nil
}

// InitUpload 初始化上传
func (s *StorageService) InitUpload(ctx context.Context, fileName string, size int64, md5 string, userID int64) (*model.File, error) {
	existedFile, err := s.fileDAO.GetFileByMD5(md5)
	if err == nil && existedFile != nil {
		return existedFile, nil
	}
	objectName := fmt.Sprintf("uploads/%s", fileName)

	file := &model.File{
		FileName:   fileName,
		Bucket:     s.bucket,
		ObjectName: objectName,
		Size:       size,
		Md5:        md5,
		Status:     0, // 上传中
		CreatedAt:  time.Now(),
		UserID:     userID,
	}

	if err := s.fileDAO.CreateFile(file); err != nil {
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}
	return file, nil
}

// UploadPart 上传分片
func (s *StorageService) UploadPart(ctx context.Context, fileID int64, partNumber int, data []byte, clientMD5 string) error {
	hash := md5.Sum(data)
	md5Str := hex.EncodeToString(hash[:])
	if md5Str != clientMD5 {
		return fmt.Errorf("数据校验失败")
	}
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return fmt.Errorf("找不到文件记录: %v", err)
	}

	partObject := fmt.Sprintf("%s.part.%d", file.ObjectName, partNumber)

	info, err := s.client.PutObject(ctx, s.bucket, partObject, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("上传分片失败: %v", err)
	}

	part := &model.FilePart{
		FileID:     file.ID,
		PartNumber: partNumber,
		ETag:       info.ETag,
		Size:       int64(len(data)),
		UploadedAt: time.Now(),
	}

	return s.fileDAO.SavePart(part)
}

// 上传分片并发
func (s *StorageService) UploadPartsConcurrently(ctx context.Context, fileID int64, parts map[int][]byte) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(parts))

	for partNumber, data := range parts {
		wg.Add(1)
		go func(pn int, d []byte) {
			defer wg.Done()
			if err := s.uploadPart(ctx, fileID, pn, d); err != nil {
				errCh <- err
			}
		}(partNumber, data)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

// 重载UploadPartsConcurrently方法，增加MD5校验支持
func (s *StorageService) UploadPartsConcurrentlyWithMD5(ctx context.Context, fileID int64, parts map[int][]byte, md5s map[int]string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(parts))

	for partNumber, data := range parts {
		wg.Add(1)
		go func(pn int, d []byte) {
			defer wg.Done()
			// 校验MD5
			hash := md5.Sum(d)
			md5Str := hex.EncodeToString(hash[:])
			if md5Str != md5s[pn] {
				errCh <- fmt.Errorf("分片%d数据校验失败", pn)
				return
			}

			if err := s.uploadPart(ctx, fileID, pn, d); err != nil {
				errCh <- err
			}
		}(partNumber, data)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

// 获取未上传的分片
func (s *StorageService) GetIncompleteParts(ctx context.Context, fileID int64, totalParts int) ([]int, error) {
	uploadedParts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return nil, err
	}

	exists := make(map[int]bool)
	for _, p := range uploadedParts {
		exists[p.PartNumber] = true
	}

	var missing []int
	for i := 1; i <= totalParts; i++ {
		if !exists[i] {
			missing = append(missing, i)
		}
	}
	return missing, nil
}

// 完成分片上传
func (s *StorageService) UploadComplete(ctx context.Context, fileID int64) error {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return err
	}

	parts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return err
	}

	var srcs []minio.CopySrcOptions
	for _, p := range parts {
		src := minio.CopySrcOptions{
			Bucket: s.bucket,
			Object: fmt.Sprintf("%s.part.%d", file.ObjectName, p.PartNumber),
		}
		srcs = append(srcs, src)
	}

	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: file.ObjectName,
	}

	// 使用正确的 ComposeObject 调用格式：上下文、目标、多个源对象
	if _, err := s.client.ComposeObject(ctx, dst, srcs...); err != nil {
		return fmt.Errorf("合并分片失败: %v", err)
	}

	return s.fileDAO.UpdateFileStatus(fileID, 1)
}

// 下载文件到Writer
func (s *StorageService) DownloadFileToWriter(ctx context.Context, fileID int64, w io.Writer) error {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return err
	}
	parts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return err
	}
	for _, p := range parts {
		partObjectName := fmt.Sprintf("%s.part.%d", file.ObjectName, p.PartNumber)
		object, err := s.client.GetObject(ctx, s.bucket, partObjectName, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("下载分片失败: %v", err)
		}
		_, err = io.Copy(w, object)
		object.Close()
		if err != nil {
			return fmt.Errorf("写入分片失败: %v", err)
		}
	}
	return nil
}

// 下载文件的指定分片
func (s *StorageService) DownloadFileWithOffset(ctx context.Context, fileID int64, startPart int, startOffset int64, w io.Writer) error {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return fmt.Errorf("获取文件失败: %v", err)
	}

	parts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return fmt.Errorf("获取分片列表失败: %v", err)
	}

	for _, p := range parts {
		if p.PartNumber < startPart {
			continue
		}

		objName := fmt.Sprintf("%s.part.%d", file.ObjectName, p.PartNumber)
		opts := minio.GetObjectOptions{}
		if p.PartNumber == startPart && startOffset > 0 {
			opts.SetRange(startOffset, 0)
		}

		object, err := s.client.GetObject(ctx, s.bucket, objName, opts)
		if err != nil {
			return fmt.Errorf("下载分片失败: %v", err)
		}

		_, err = io.Copy(w, object)
		object.Close()
		if err != nil {
			return fmt.Errorf("写入分片失败: %v", err)
		}
	}

	return nil
}

// 下载分片数据
func (s *StorageService) DownloadChunk(ctx context.Context, fileID int64, chunkIndex int, startOffset int64) ([]byte, string, error) {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return nil, "", fmt.Errorf("获取文件信息失败: %v", err)
	}
	part, err := s.fileDAO.GetPart(fileID, chunkIndex)
	if err != nil {
		return nil, "", fmt.Errorf("获取分片信息失败: %v", err)
	}

	objName := fmt.Sprintf("%s.part.%d", file.ObjectName, chunkIndex)
	opts := minio.GetObjectOptions{}
	if startOffset > 0 {
		opts.SetRange(startOffset, 0)
	}

	object, err := s.client.GetObject(ctx, s.bucket, objName, opts)
	if err != nil {
		return nil, "", fmt.Errorf("下载分片失败: %v", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, "", fmt.Errorf("读取分片失败: %v", err)
	}

	return data, part.ETag, nil
}

// 获取上传进度
func (s *StorageService) GetUploadProgress(fileID int64) (uploadedSize int64, totalSize int64, err error) {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return 0, 0, err
	}

	parts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return 0, 0, err
	}

	var uploaded int64
	for _, p := range parts {
		uploaded += p.Size
	}

	return uploaded, file.Size, nil
}

// 删除分片
func (s *StorageService) DeleteFile(ctx context.Context, fileID int64) error {
	parts, err := s.fileDAO.ListParts(fileID)
	if err != nil {
		return err
	}
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return err
	}
	// 删除分片
	for _, p := range parts {
		objName := fmt.Sprintf("%s.part.%d", file.ObjectName, p.PartNumber)
		err := s.client.RemoveObject(ctx, s.bucket, objName, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("删除分片 %s 失败: %v", objName, err)
		}
	}
	objName := file.ObjectName
	err = s.client.RemoveObject(ctx, s.bucket, objName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件 %s 失败: %v", objName, err)
	}
	// 删除数据库记录
	return s.fileDAO.DeleteParts(fileID)
}

// 上传分片
func (s *StorageService) uploadPart(ctx context.Context, fileID int64, partNumber int, data []byte) error {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return fmt.Errorf("找不到文件记录: %v", err)
	}

	partObject := fmt.Sprintf("%s.part.%d", file.ObjectName, partNumber)

	info, err := s.client.PutObject(ctx, s.bucket, partObject, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("上传分片失败: %v", err)
	}

	part := &model.FilePart{
		FileID:     file.ID,
		PartNumber: partNumber,
		ETag:       info.ETag,
		Size:       int64(len(data)),
		UploadedAt: time.Now(),
	}

	return s.fileDAO.SavePart(part)
}

// GetFileInfo 获取文件信息
func (s *StorageService) GetFileInfo(ctx context.Context, fileID int64) (*model.File, error) {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 检查文件状态
	if file.Status != 1 {
		return nil, fmt.Errorf("文件未完成上传或不可用")
	}

	return file, nil
}

func (s *StorageService) FileDAO() model.FileDAO {
	return s.fileDAO
}
func (s *StorageService) GeneratePresignedURL(ctx context.Context, fileID int64, expireSeconds int32) (string, int64, error) {
	file, err := s.fileDAO.GetFileByID(fileID)
	if err != nil {
		return "", 0, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 检查文件状态是否已完成
	if file.Status != 1 {
		return "", 0, fmt.Errorf("文件尚未完成上传")
	}

	// 设置默认过期时间为1小时
	if expireSeconds <= 0 {
		expireSeconds = 3600
	}

	// 计算过期时间
	expireTime := time.Now().Add(time.Duration(expireSeconds) * time.Second)

	// 生成预签名URL
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, file.ObjectName, time.Duration(expireSeconds)*time.Second, nil)
	if err != nil {
		return "", 0, fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return presignedURL.String(), expireTime.Unix(), nil
}

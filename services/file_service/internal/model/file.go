package model

import (
	"time"

	"gorm.io/gorm"
)

// -------------------- 数据表结构 --------------------

// File 文件元数据
type File struct {
	ID         int64  `gorm:"primaryKey"`
	FileName   string `gorm:"size:255"`
	Bucket     string `gorm:"size:255"`
	ObjectName string `gorm:"size:512"` // 最终合并对象名
	Size       int64
	Md5        string
	Status     int // 0 = uploading, 1 = completed
	CreatedAt  time.Time
}

// FilePart 分片信息
type FilePart struct {
	ID         int64 `gorm:"primaryKey"`
	FileID     int64 `gorm:"index"`
	PartNumber int
	ETag       string `gorm:"size:255"` // MinIO 返回的 ETag
	Size       int64
	UploadedAt time.Time
}

// -------------------- DAO 接口 --------------------

type FileDAO interface {
	CreateFile(file *File) error
	GetFileByID(id int64) (*File, error)
	UpdateFileStatus(id int64, status int) error
	SavePart(part *FilePart) error
	ListParts(fileID int64) ([]FilePart, error)
	DeleteParts(fileID int64) error
	GetPart(fileID int64, partNumber int) (*FilePart, error)
	GetFileByMD5(md5 string) (*File, error)
}

// -------------------- DAO 实现 --------------------

type fileDAOImpl struct {
	db *gorm.DB
}

// NewFileDAO 创建 DAO 实例
func NewFileDAO(db *gorm.DB) *fileDAOImpl {
	// 自动迁移表
	db.AutoMigrate(&File{}, &FilePart{})
	return &fileDAOImpl{db: db}
}

// CreateFile 创建文件记录
func (dao *fileDAOImpl) CreateFile(file *File) error {
	return dao.db.Create(file).Error
}

// GetFileByID 获取文件信息
func (dao *fileDAOImpl) GetFileByID(id int64) (*File, error) {
	var file File
	err := dao.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, err
}

// UpdateFileStatus 更新文件状态
func (dao *fileDAOImpl) UpdateFileStatus(id int64, status int) error {
	return dao.db.Model(&File{}).Where("id = ?", id).Update("status", status).Error
}

// SavePart 保存分片信息
func (dao *fileDAOImpl) SavePart(part *FilePart) error {
	// 如果分片已存在，先删除再保存
	var existing FilePart
	if err := dao.db.Where("file_id = ? AND part_number = ?", part.FileID, part.PartNumber).First(&existing).Error; err == nil {
		dao.db.Delete(&existing)
	}
	return dao.db.Create(part).Error
}

// GetPart 获取分片信息
func (dao *fileDAOImpl) GetPart(fileID int64, partNumber int) (*FilePart, error) {
	var part FilePart
	err := dao.db.Where("file_id = ? AND part_number = ?", fileID, partNumber).First(&part).Error
	if err != nil {
		return nil, err
	}
	return &part, err
}

// ListParts 查询所有已上传分片
func (dao *fileDAOImpl) ListParts(fileID int64) ([]FilePart, error) {
	var parts []FilePart
	err := dao.db.Where("file_id = ?", fileID).Order("part_number asc").Find(&parts).Error
	if err != nil {
		return nil, err
	}
	return parts, err
}

// DeleteParts 删除文件所有分片（取消上传或上传失败时使用）
func (dao *fileDAOImpl) DeleteParts(fileID int64) error {
	return dao.db.Where("file_id = ?", fileID).Delete(&FilePart{}).Error
}
func (dao *fileDAOImpl) DeleteFile(fileID int64) error {
	return dao.db.Where("file_id = ?", fileID).Delete(&FilePart{}).Error
}
func (dao *fileDAOImpl) GetFileByMD5(MD5 string) (*File, error) {
	var file File
	err := dao.db.Where("md5 = ?", MD5).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, err
}

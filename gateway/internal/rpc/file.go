package rpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/keepalive"

	filepb "github.com/waitform/micro-cloud-storage/protos/file/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FileServiceClient 封装文件服务客户端
type FileServiceClient struct {
	*ServiceClient
	grpcClient filepb.FileServiceClient
	conn       *grpc.ClientConn
}

// NewFileServiceClient 创建文件服务客户端
func NewFileServiceClient(serviceClient *ServiceClient) (*FileServiceClient, error) {
	// 使用etcd解析器创建连接
	target := fmt.Sprintf("etcd:///%s", "file-service")
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second, // 每20秒发送一次ping
			Timeout:             5 * time.Second,  // ping超时时间
			PermitWithoutStream: true,             // 允许在没有活跃流时发送ping
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial file-service: %w", err)
	}

	// 创建gRPC客户端
	grpcClient := filepb.NewFileServiceClient(conn)

	return &FileServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
		conn:          conn,
	}, nil
}

// Close 关闭gRPC连接
func (f *FileServiceClient) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}

// InitUpload 初始化上传
func (f *FileServiceClient) InitUpload(ctx context.Context, req *filepb.InitUploadRequest) (*filepb.InitUploadResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.InitUpload(ctx, req)
}

// UploadPart 上传分片
func (f *FileServiceClient) UploadPart(ctx context.Context, req *filepb.UploadPartRequest) (*emptypb.Empty, error) {
	// 为上传分片设置更长的超时时间，因为可能涉及大文件传输
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// 创建流
	stream, err := f.grpcClient.UploadPart(ctx)
	if err != nil {
		return nil, err
	}

	// 发送请求
	err = stream.Send(req)
	if err != nil {
		return nil, err
	}

	// 关闭流并接收响应
	return stream.CloseAndRecv()
}

func (f *FileServiceClient) UploadPartStream(
	ctx context.Context,
	fileID int64,
	partNumber int32,
	partSize int64,
	reader io.Reader,
	partMD5 string,
) (*emptypb.Empty, error) {
	// === 1️⃣ 超时控制 ===
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	// === 2️⃣ 建立 gRPC 流 ===
	stream, err := f.grpcClient.UploadPart(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建上传流失败: %v", err)
	}

	// === 3️⃣ 发送元数据 ===
	metaMsg := &filepb.UploadPartRequest{
		PartData: &filepb.UploadPartRequest_PartMetadata{
			PartMetadata: &filepb.PartMetadata{
				FileId:     fileID,
				PartNumber: int64(partNumber),
				Size:       partSize,
				Md5:        partMD5,
			},
		},
	}

	if err := stream.Send(metaMsg); err != nil {
		return nil, fmt.Errorf("发送元数据失败: %v", err)
	}

	// === 4️⃣ 分块读取文件并发送数据 ===
	buf := make([]byte, 512*1024) // 512KB 缓冲区
	var total int64

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			total += int64(n)
			dataMsg := &filepb.UploadPartRequest{
				PartData: &filepb.UploadPartRequest_PartContent{
					PartContent: &filepb.PartContent{
						Data: buf[:n],
					},
				},
			}
			if err := stream.Send(dataMsg); err != nil {
				return nil, fmt.Errorf("发送数据块失败: %v", err)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取数据出错: %v", err)
		}
	}

	// === 5️⃣ 关闭发送并等待响应 ===
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("上传完成但等待响应失败: %v", err)
	}

	fmt.Printf("[UploadPartStream] 分片上传完成: fileID=%d, part=%d, size=%.2fMB\n",
		fileID, partNumber, float64(total)/1024/1024)

	return resp, nil
}

// CompleteUpload 完成上传
func (f *FileServiceClient) CompleteUpload(ctx context.Context, req *filepb.CompleteUploadRequest) (*filepb.CompleteUploadResponse, error) {
	// 设置默认超时时间，完成上传可能需要合并大量分片，所以设置较长超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	return f.grpcClient.CompleteUpload(ctx, req)
}

// GetFileInfo 获取文件信息
func (f *FileServiceClient) GetFileInfo(ctx context.Context, req *filepb.GetFileInfoRequest) (*filepb.GetFileInfoResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.GetFileInfo(ctx, req)
}

// GeneratePresignedURL 生成预签名URL
func (f *FileServiceClient) GeneratePresignedURL(ctx context.Context, req *filepb.GeneratePresignedURLRequest) (*filepb.GeneratePresignedURLResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.GeneratePresignedURL(ctx, req)
}

// GetUploadProgress 获取上传进度
func (f *FileServiceClient) GetUploadProgress(ctx context.Context, req *filepb.GetUploadProgressRequest) (*filepb.GetUploadProgressResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.GetUploadProgress(ctx, req)
}

// GetIncompleteParts 获取未完成分片
func (f *FileServiceClient) GetIncompleteParts(ctx context.Context, req *filepb.GetIncompletePartsRequest) (*filepb.GetIncompletePartsResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.GetIncompleteParts(ctx, req)
}

// CancelUpload 取消上传
func (f *FileServiceClient) CancelUpload(ctx context.Context, req *filepb.CancelUploadRequest) (*emptypb.Empty, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.CancelUpload(ctx, req)
}

// DeleteFile 删除文件
func (f *FileServiceClient) DeleteFile(ctx context.Context, req *filepb.DeleteRequest) (*emptypb.Empty, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return f.grpcClient.DeleteFile(ctx, req)
}

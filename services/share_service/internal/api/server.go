package api

import (
	"cloud-storage-share-service/internal/model"
	pb "cloud-storage-share-service/proto"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShareServer struct {
	pb.UnimplementedShareServiceServer
	dao *model.ShareDAO
}

func NewShareHandler(dao *model.ShareDAO) *ShareServer {
	return &ShareServer{dao: dao}
}
func (s *ShareServer) CreateShare(ctx context.Context, req *pb.CreateShareRequest) (*pb.CreateShareResponse, error) {
	shareID, err := uuid.NewUUID()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成分享ID失败")
	}
	share := &model.Share{
		FileID:   req.GetFileId(),
		OwnerID:  req.GetOwnerId(),
		Password: string(hash),
		ShareID:  shareID.String(),
		ExpireAt: time.Now().Add(time.Duration(req.ExpireIn) * time.Second),
	}

	err = s.dao.Create(share)
	if err != nil {
		return nil, err
	}
	return &pb.CreateShareResponse{
		ShareId: share.ShareID,
	}, nil
}
func (s *ShareServer) GetShareInfo(ctx context.Context, req *pb.GetShareInfoRequest) (*pb.GetShareInfoResponse, error) {
	id := req.ShareId
	// 这里假设 share_id 可直接对应数据库 ID（如果你有单独字段，请修改查询逻辑）
	share, err := s.dao.GetByShareID(req.ShareId)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, errors.New("share not found")
	}

	info := &pb.ShareInfo{
		ShareId:   id,
		FileId:    share.FileID,
		OwnerId:   share.OwnerID,
		Password:  share.Password,
		ExpireAt:  share.ExpireAt.Unix(),
		CreatedAt: share.CreatedAt.Format(time.RFC3339),
	}
	return &pb.GetShareInfoResponse{Info: info}, nil
}
func (s *ShareServer) ValidateAccess(ctx context.Context, req *pb.ValidateAccessRequest) (*pb.ValidateAccessResponse, error) {
	share, err := s.dao.GetByShareID(req.ShareId)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, errors.New("share not found")
	}
	if time.Now().After(share.ExpireAt) {
		return nil, errors.New("share expired")
	}

	valid := checkPasswordHash(req.Password, share.Password)
	resp := &pb.ValidateAccessResponse{
		Valid:   valid,
		FileId:  share.FileID,
		OwnerId: share.OwnerID,
	}
	return resp, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

package types

// 用户注册
type CreateUserRequest struct {
	Username string
	Password string
	Nickname string
}

type CreateUserResponse struct {
	ID int64
}

// 用户信息
type UserInfo struct {
	ID         int64
	Username   string
	Avatar     string
	UsedSpace  int64
	TotalSpace int64
}

type RegisterRequest struct {
	Username string
	Password string
	Email    string
}

type RegisterResponse struct {
	Success bool
	Message string
	User    *UserInfo
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Success bool
	Message string
	UserID  int64
	Token   string
}

type GetUserInfoRequest struct {
	ID int64
}

type GetUserInfoResponse struct {
	User *UserInfo
}

// 更新用户信息
type UpdateUserRequest struct {
	ID int64

	Avatar string
}

// 修改容量
type UpdateCapacityRequest struct {
	UserID   int64
	NewTotal int64
}

// 检查容量是否足够
type CheckCapacityRequest struct {
	UserID   int64
	FileSize int64
}

type CheckCapacityResponse struct {
	IsEnough bool
}
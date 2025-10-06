package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret = []byte("cloud-storage-secret-key") // 实际项目中应该从配置文件读取
)

// Claims 定义JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// SetSecret 设置JWT密钥（从配置文件读取）
func SetSecret(secret []byte) {
	jwtSecret = secret
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string) (string, error) {
	// 设置token过期时间 (7*24小时)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-cloud-storage",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateTokenWithDuration 生成指定过期时间的JWT token
func GenerateTokenWithDuration(userID uint, username string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-cloud-storage",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并验证JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新JWT token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查token是否快要过期（例如还剩1小时以内）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not expired or close to expiration")
	}

	// 生成新的token
	return GenerateToken(claims.UserID, claims.Username)
}

// GetUserIDFromToken 从token中获取用户ID
func GetUserIDFromToken(tokenString string) (uint, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetUsernameFromToken 从token中获取用户名
func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}

// ShareClaims 定义分享JWT声明结构
type ShareClaims struct {
	ShareID   uint   `json:"share_id"`
	Token     string `json:"token"`
	jwt.RegisteredClaims
}

// GenerateShareToken 生成分享JWT token
func GenerateShareToken(shareID uint, token string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)

	claims := &ShareClaims{
		ShareID: shareID,
		Token:   token,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-cloud-storage-share",
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(jwtSecret)
}

// ParseShareToken 解析并验证分享JWT token
func ParseShareToken(tokenString string) (*ShareClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ShareClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ShareClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid share token")
}

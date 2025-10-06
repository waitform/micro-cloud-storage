package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// 32 字节的 AES-256 Key（建议放到配置文件或环境变量，而不是写死）
var encryptionKey = []byte("cloud-storage-secret-key-1234567890abcd")

// EncryptTokenToUuid 使用 AES-GCM 加密 token，并转成 UUID-like 格式
func EncryptTokenToUuid(token string) (string, error) {
	// 1. 创建 AES 加密器
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 2. 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 3. 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 4. 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)

	// 5. 转 hex 编码
	encoded := hex.EncodeToString(ciphertext)

	// 6. 格式化成 UUID-like
	if len(encoded) < 32 {
		return "", fmt.Errorf("encoded too short")
	}
	uuid := fmt.Sprintf("%s-%s-%s-%s-%s",
		encoded[0:8],
		encoded[8:12],
		encoded[12:16],
		encoded[16:20],
		encoded[20:32])

	return uuid, nil
}

// DecryptUuidToToken 解密 UUID-like token，恢复原始字符串
func DecryptUuidToToken(uuid string) (string, error) {
	// 1. 移除 "-"
	cleanUuid := ""
	for _, ch := range uuid {
		if ch != '-' {
			cleanUuid += string(ch)
		}
	}

	// 2. 还原 hex 数据
	ciphertext, err := hex.DecodeString(cleanUuid)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %w", err)
	}

	// 3. 创建 AES 解密器
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 4. 拆分 nonce 和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 5. 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

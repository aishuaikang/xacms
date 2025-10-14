package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sync"
	"uav_defender/internal/pkg/config"

	"golang.org/x/crypto/bcrypt"
)

type SimplePasswordCrypto struct {
	serverSalt string // 服务端盐值
}

var (
	instance *SimplePasswordCrypto
	once     sync.Once
)

// NewSimplePasswordCrypto 创建简单密码加密实例（单例模式）
func NewSimplePasswordCrypto() *SimplePasswordCrypto {
	once.Do(func() {
		instance = &SimplePasswordCrypto{
			serverSalt: config.AppConfig.Configuration.PasswordSalt,
		}
	})
	return instance
}

// GetInstance 获取单例实例
func GetInstance() *SimplePasswordCrypto {
	return NewSimplePasswordCrypto()
}

// ValidateClientHashedPassword 验证前端传来的哈希密码格式（SHA256+Base64）
func (spc *SimplePasswordCrypto) ValidateClientHashedPassword(hashedPassword string) error {
	// 验证Base64格式并解码
	decoded, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return errors.New("密码格式错误：非有效的Base64编码")
	}

	// SHA256哈希应该是32字节
	if len(decoded) != 32 {
		return errors.New("密码格式错误：哈希长度不正确")
	}

	return nil
}

// ProcessClientPassword 处理前端传来的哈希密码
func (spc *SimplePasswordCrypto) ProcessClientPassword(clientHashedPassword string) (string, error) {
	// 1. 验证格式
	if err := spc.ValidateClientHashedPassword(clientHashedPassword); err != nil {
		return "", err
	}

	// 2. 解码Base64获取原始哈希字节
	hashBytes, err := base64.StdEncoding.DecodeString(clientHashedPassword)
	if err != nil {
		return "", err
	}

	// 3. 转换为十六进制字符串
	clientHashHex := hex.EncodeToString(hashBytes)

	// 4. 服务端二次哈希（加服务端盐）
	serverHash := spc.hashWithServerSalt(clientHashHex)

	// 5. 使用 bcrypt 最终哈希
	finalHash, err := bcrypt.GenerateFromPassword([]byte(serverHash), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(finalHash), nil
}

// VerifyPassword 验证密码
func (spc *SimplePasswordCrypto) VerifyPassword(clientHashedPassword, storedHash string) bool {
	// 1. 验证客户端哈希格式
	if err := spc.ValidateClientHashedPassword(clientHashedPassword); err != nil {
		return false
	}

	// 2. 解码Base64获取原始哈希字节
	hashBytes, err := base64.StdEncoding.DecodeString(clientHashedPassword)
	if err != nil {
		return false
	}

	// 3. 转换为十六进制字符串
	clientHashHex := hex.EncodeToString(hashBytes)

	// 4. 重新计算服务端哈希
	serverHash := spc.hashWithServerSalt(clientHashHex)

	// 5. 使用 bcrypt 验证
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(serverHash))
	return err == nil
}

// hashWithServerSalt 使用服务端盐值进行二次哈希
func (spc *SimplePasswordCrypto) hashWithServerSalt(clientHash string) string {
	// 客户端哈希 + 服务端盐值 -> 服务端哈希
	combined := clientHash + spc.serverSalt
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

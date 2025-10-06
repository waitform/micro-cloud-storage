package global

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Timeout   int      `yaml:"timeout"` // 单位秒
}

// MinioConfig MinIO配置
type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	UseSSL    bool   `yaml:"useSSL"`
}

// GlobalConfig 全局配置结构
type GlobalConfig struct {
	Etcd  EtcdConfig  `yaml:"etcd"`
	Minio MinioConfig `yaml:"minio"`
}

// LoadConfig 加载全局配置文件
func LoadConfig(configPath string) (*GlobalConfig, error) {
	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("global config file does not exist: %s", absPath)
	}

	// 读取配置文件内容
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config file: %v", err)
	}

	// 解析YAML配置
	var config GlobalConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse global config file: %v", err)
	}

	return &config, nil
}

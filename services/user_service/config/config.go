package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	JWT      JWTConfig      `yaml:"jwt"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

// AppConfig 全局配置变量
var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig() {
	// 读取配置文件
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	AppConfig = &Config{}
	err = yaml.Unmarshal(data, AppConfig)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
}
package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Environment string
}

type JWTConfig struct {
	Secret string
	AccessTokenDuration string
	RefreshTokenDuration string
}

func Load() *Config {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// 尝试从当前目录的父目录加载
		if err := godotenv.Load(filepath.Join("..", ".env")); err != nil {
			// 如果找不到 .env 文件，继续使用环境变量
			fmt.Printf("Warning: Could not load .env file: %v\n", err)
		}
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "ltx"),
			Password: getEnv("DB_PASSWORD", ""),  // 不提供默认值，强制使用环境变量
			DBName:   getEnv("DB_NAME", "jobView_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8010"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""), // 不提供默认值，强制使用环境变量
			AccessTokenDuration:  getEnv("JWT_ACCESS_DURATION", "24h"),
			RefreshTokenDuration: getEnv("JWT_REFRESH_DURATION", "720h"), // 30天
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment 检查是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction 检查是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	// 生产环境必须设置数据库密码
	if c.IsProduction() && c.Database.Password == "" {
		return fmt.Errorf("production environment requires DB_PASSWORD to be set")
	}
	
	// 生产环境必须设置JWT密钥
	if c.IsProduction() && c.JWT.Secret == "" {
		return fmt.Errorf("production environment requires JWT_SECRET to be set")
	}
	
	// JWT密钥长度检查
	if len(c.JWT.Secret) > 0 && len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	
	return nil
}
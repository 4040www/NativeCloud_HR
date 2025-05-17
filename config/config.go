package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var JWTSecret []byte

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	}
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	}
	Redis struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	MessageQueue struct {
		Type    string   `yaml:"type"`
		Brokers []string `yaml:"brokers"`
	}
}

func LoadConfig() (*Config, error) {
	// 1. 載入 .env
	_ = godotenv.Load()

	// 2. 讀取 YAML 設定檔
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// 3. 使用 .env 覆蓋 YAML 的 database 欄位（如果有提供）
	if val := os.Getenv("DB_HOST"); val != "" {
		cfg.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		// optional: parse string to int if needed
	}
	if val := os.Getenv("DB_USER"); val != "" {
		cfg.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		cfg.Database.Password = val
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		cfg.Database.Name = val
	}

	JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(JWTSecret) == 0 {
		return nil, fmt.Errorf("JWT_SECRET 未設定")
	}

	return &cfg, nil
}

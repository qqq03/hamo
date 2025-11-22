package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
)

// DBCredentials DB 인증 정보
type DBCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config 애플리케이션 설정
type Config struct {
	DBUser     string
	DBPass     string
	DBHost     string
	DBPort     string
	DBName     string
	ServerPort string
}

// LoadConfig 설정 로드
func LoadConfig() (*Config, error) {
	// .env 파일 로드 (로컬 개발용)
	_ = godotenv.Load(".env")

	cfg := &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPass:     os.Getenv("DB_PASS"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}

	// AWS Secrets Manager 사용
	if os.Getenv("USE_SECRETS_MANAGER") == "true" {
		creds, err := getSecretFromAWS(
			os.Getenv("SECRET_NAME"),
			os.Getenv("AWS_REGION"),
		)
		if err == nil {
			cfg.DBUser = creds.Username
			cfg.DBPass = creds.Password
			log.Println("AWS Secrets Manager credentials loaded.")
		} else {
			log.Printf("AWS Secrets Manager load failed: %v", err)
		}
	}

	// 기본값 설정
	if cfg.DBPort == "" {
		cfg.DBPort = "3306"
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
		cfg.DBName = "museumdb"
		cfg.DBUser = "root"
	}

	return cfg, nil
}

// getSecretFromAWS AWS Secrets Manager에서 비밀 정보 가져오기
func getSecretFromAWS(secretName string, region string) (*DBCredentials, error) {
	if region == "" {
		region = "ap-northeast-2"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("AWS config error: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return nil, err
	}

	var creds DBCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

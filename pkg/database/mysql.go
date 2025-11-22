package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLConnection MySQL 연결 생성
func NewMySQLConnection(user, pass, host, port, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB open error: %w", err)
	}

	// 연결 풀 설정
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// DB 연결 테스트
	if os.Getenv("SKIP_DB_CHECK") != "true" {
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("DB ping error: %w", err)
		}
		log.Println("DB 연결 성공")
	}

	return db, nil
}

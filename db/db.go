package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func Init() error {
	_ = godotenv.Load() // .env を読み込む（存在しない場合は無視）

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	if user == "" || pass == "" || host == "" || port == "" || name == "" {
		return fmt.Errorf("DB接続情報が不足しています")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("sql.Openエラー: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("DB.Ping失敗: %w", err)
	}

	log.Println("✅ DB接続に成功しました")
	return nil
}

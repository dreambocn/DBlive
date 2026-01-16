// 数据库连接与初始化
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	// 启用外键约束并构造 DSN
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	// 连接探活确保路径可用
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

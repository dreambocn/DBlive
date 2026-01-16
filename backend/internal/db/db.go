package db

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
    dsn := fmt.Sprintf("file:%s?_foreign_keys=on", path)
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}

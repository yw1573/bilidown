package store

import (
	"database/sql"
)

// CreateLog 创建日志记录
func CreateLog(db *sql.DB, content string) error {
	SqliteLock.Lock()
	_, err := db.Exec(`INSERT INTO "log" ("content") VALUES (?)`, content)
	SqliteLock.Unlock()
	return err
}

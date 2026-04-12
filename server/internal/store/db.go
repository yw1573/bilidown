package store

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var SqliteLock sync.Mutex

// MustGetDB 获取数据库连接，数据库路径默认为二进制文件所在目录下的 database/data.db
func MustGetDB(path ...string) *sql.DB {
	var pathStr string
	if len(path) == 0 {
		// 获取二进制文件所在目录
		execPath, err := os.Executable()
		if err != nil {
			log.Fatalln("os.Executable:", err)
		}
		execDir := filepath.Dir(execPath)
		dbDir := filepath.Join(execDir, "database")
		// 确保 database 目录存在
		if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
			log.Fatalln("创建 database 目录失败:", err)
		}
		pathStr = filepath.Join(dbDir, "data.db")
	} else {
		pathStr = path[0]
	}

	db, err := sql.Open("sqlite", pathStr)
	if err != nil {
		log.Fatalln("sql.Open:", err)
	}
	return db
}

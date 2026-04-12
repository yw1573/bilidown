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

// MustGetDB 获取数据库连接，数据库路径默认为二进制文件所在目录下的 data/data.db
func MustGetDB(path ...string) *sql.DB {
	var pathStr string
	if len(path) == 0 {
		// 获取二进制文件所在目录
		execPath, err := os.Executable()
		if err != nil {
			log.Fatalln("os.Executable:", err)
		}
		execDir := filepath.Dir(execPath)
		dataDir := filepath.Join(execDir, "data")
		// 确保 data 目录存在
		if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
			log.Fatalln("创建 data 目录失败:", err)
		}
		pathStr = filepath.Join(dataDir, "data.db")
	} else {
		pathStr = path[0]
	}

	db, err := sql.Open("sqlite", pathStr)
	if err != nil {
		log.Fatalln("sql.Open:", err)
	}
	return db
}

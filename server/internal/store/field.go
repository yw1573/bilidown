package store

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"bilidown/internal/util"
)

type FieldUtil struct{}

func (f FieldUtil) AllowSelect() []string {
	return []string{
		"download_folder",
	}
}

func (f FieldUtil) AllowUpdate() []string {
	return []string{
		"download_folder",
	}
}

func (f FieldUtil) IsAllow(allFields []string, names ...string) bool {
	allowedFields := make(map[string]struct{})
	for _, field := range allFields {
		allowedFields[field] = struct{}{}
	}
	for _, name := range names {
		if _, exists := allowedFields[name]; !exists {
			return false
		}
	}
	return true
}

func (f FieldUtil) IsAllowSelect(names ...string) bool {
	return f.IsAllow(f.AllowSelect(), names...)
}

func (f FieldUtil) IsAllowUpdate(names ...string) bool {
	return f.IsAllow(f.AllowUpdate(), names...)
}

// GetFields 获取字段值
func GetFields(db *sql.DB, names ...string) (map[string]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(names))
	for i := 0; i < len(names); i++ {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf(`SELECT "name", "value" FROM "field" WHERE "name" IN (%s)`, strings.Join(placeholders, ","))

	values := make([]interface{}, len(names))
	for i := 0; i < len(names); i++ {
		values[i] = names[i]
	}
	SqliteLock.Lock()
	row, err := db.Query(query, values...)
	SqliteLock.Unlock()
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var name, value string
	fields := make(map[string]string)
	for row.Next() {
		if err := row.Scan(&name, &value); err != nil {
			return nil, err
		}
		fields[name] = value
	}
	return fields, nil
}

// SaveFields 保存字段值
func SaveFields(db *sql.DB, data [][2]string) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO "field" ("name", "value") VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, d := range data {
		SqliteLock.Lock()
		_, err = stmt.Exec(d[0], d[1])
		SqliteLock.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCurrentFolder 获取数据库中的下载保存路径，如果不存在则将默认路径保存到数据库
func GetCurrentFolder(db *sql.DB) (string, error) {
	var folder string
	SqliteLock.Lock()
	err := db.QueryRow(`SELECT "value" FROM "field" WHERE "name" = 'download_folder'`).Scan(&folder)
	SqliteLock.Unlock()
	if err != nil && err == sql.ErrNoRows {
		folder, err = util.GetDefaultDownloadFolder()
		if err != nil {
			return "", err
		}
		err = os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return "", err
		}
		err = SaveDownloadFolder(db, folder)
		if err != nil {
			return "", err
		}
		return folder, nil
	}
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return "", err
	}
	return folder, nil
}

// SaveDownloadFolder 保存下载路径，不存在则自动创建
func SaveDownloadFolder(db *sql.DB, downloadFolder string) error {
	_, err := os.Stat(downloadFolder)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(downloadFolder, os.ModePerm)
			if err != nil {
				return err
			}
		}
		return err
	}
	SqliteLock.Lock()
	_, err = db.Exec(`INSERT OR REPLACE INTO "field" ("name", "value") VALUES ('download_folder', ?)`, downloadFolder)
	SqliteLock.Unlock()
	return err
}

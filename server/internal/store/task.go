package store

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bilidown/internal/common"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusDone    TaskStatus = "done"
	TaskStatusWaiting TaskStatus = "waiting"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusError   TaskStatus = "error"
)

// TaskInitOption 创建任务时需要从 POST 请求获取的参数
type TaskInitOption struct {
	Bvid         string             `json:"bvid"`
	Cid          int                `json:"cid"`
	Format       common.MediaFormat `json:"format"`
	Title        string             `json:"title"`
	Owner        string             `json:"owner"`
	Cover        string             `json:"cover"`
	Status       TaskStatus         `json:"status"`
	Folder       string             `json:"folder"`
	Audio        string             `json:"audio"`
	Video        string             `json:"video"`
	Duration     int                `json:"duration"`
	DownloadType string             `json:"downloadType"`
}

// TaskInDB 任务数据库中的数据
type TaskInDB struct {
	TaskInitOption
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"createAt"`
}

// FilePath 生成任务文件路径
func (task *TaskInDB) FilePath() string {
	ext := ".mp4"
	if task.DownloadType == "audio" {
		ext = ".m4a"
	}
	return filepath.Join(task.Folder,
		fmt.Sprintf("%s %s%s", task.Title,
			strings.Replace(base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(task.ID, 10))), "=", "", -1),
			ext,
		),
	)
}

// CreateTask 创建任务记录
func CreateTask(db *sql.DB, task *TaskInitOption) (int64, error) {
	SqliteLock.Lock()
	result, err := db.Exec(`INSERT INTO "task" ("bvid", "cid", "format", "title", "owner", "cover", "status", "folder", "duration", "download_type")
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.Bvid,
		task.Cid,
		task.Format,
		task.Title,
		task.Owner,
		task.Cover,
		task.Status,
		task.Folder,
		task.Duration,
		task.DownloadType,
	)
	SqliteLock.Unlock()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(db *sql.DB, taskID int64, status TaskStatus) error {
	SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "task" SET "status" = ? WHERE "id" = ?`, status, taskID)
	SqliteLock.Unlock()
	return err
}

// GetTaskList 获取任务列表
func GetTaskList(db *sql.DB, page int, pageSize int) ([]TaskInDB, error) {
	tasks := []TaskInDB{}
	SqliteLock.Lock()
	rows, err := db.Query(`SELECT
		"id", "bvid", "cid", "format", "title",
		"owner", "cover", "status", "folder", "duration", "download_type", "create_at"
	FROM "task" ORDER BY "id" DESC LIMIT ?, ?`,
		page*pageSize, pageSize,
	)
	SqliteLock.Unlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	createAt := ""

	for rows.Next() {
		task := TaskInDB{}
		err = rows.Scan(
			&task.ID,
			&task.Bvid,
			&task.Cid,
			&task.Format,
			&task.Title,
			&task.Owner,
			&task.Cover,
			&task.Status,
			&task.Folder,
			&task.Duration,
			&task.DownloadType,
			&createAt,
		)
		if err != nil {
			return nil, err
		}
		task.CreateAt, err = time.Parse("2006-01-02 15:04:05", createAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// DeleteTask 删除任务记录
func DeleteTask(db *sql.DB, taskID int) error {
	SqliteLock.Lock()
	_, err := db.Exec(`DELETE FROM "task" WHERE "id" = ?`, taskID)
	SqliteLock.Unlock()
	return err
}

// GetTask 获取单个任务
func GetTask(db *sql.DB, taskID int) (*TaskInDB, error) {
	task := TaskInDB{}
	createAt := ""
	SqliteLock.Lock()
	err := db.QueryRow(`SELECT
		"id", "bvid", "cid", "format", "title",
		"owner", "cover", "status", "folder", "duration", "download_type", "create_at"
	FROM "task" WHERE "id" = ?`,
		taskID,
	).Scan(
		&task.ID,
		&task.Bvid,
		&task.Cid,
		&task.Format,
		&task.Title,
		&task.Owner,
		&task.Cover,
		&task.Status,
		&task.Folder,
		&task.Duration,
		&task.DownloadType,
		&createAt,
	)
	SqliteLock.Unlock()
	if err != nil {
		return nil, err
	}

	task.CreateAt, err = time.Parse("2006-01-02 15:04:05", createAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
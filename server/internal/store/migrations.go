package store

import (
	"database/sql"
	"log"
)

// InitTables 初始化数据表
func InitTables(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "field" (
		"name" TEXT PRIMARY KEY NOT NULL,
		"value" TEXT
	)`); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "log" (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"content" TEXT NOT NULL,
		"create_at" text NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "task" (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"bvid" text NOT NULL,
		"cid" integer NOT NULL,
		"format" integer NOT NULL,
		"title" text NOT NULL,
		"owner" text NOT NULL,
		"cover" text NOT NULL,
		"status" text NOT NULL,
		"folder" text NOT NULL,
		"duration" integer NOT NULL,
		"download_type" text NOT NULL DEFAULT 'merge',
		"create_at" text NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}

	if _, err := GetCurrentFolder(db); err != nil {
		return err
	}

	if err := initHistoryTask(db); err != nil {
		return err
	}

	// 添加可能缺失的列（用于数据库迁移）
	if err := addMissingColumns(db); err != nil {
		return err
	}

	return nil
}

// addMissingColumns 添加可能缺失的列（用于数据库迁移）
func addMissingColumns(db *sql.DB) error {
	SqliteLock.Lock()
	_, _ = db.Exec(`ALTER TABLE "task" ADD COLUMN "download_type" TEXT DEFAULT 'merge'`)
	_, _ = db.Exec(`UPDATE "task" SET "download_type" = 'merge' WHERE "download_type" IS NULL`)
	SqliteLock.Unlock()
	return nil
}

// initHistoryTask 将上一次程序运行时未完成的任务进度全部变为 error
func initHistoryTask(db *sql.DB) error {
	SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "task" SET "status" = 'error' WHERE "status" IN ('waiting', 'running')`)
	SqliteLock.Unlock()
	return err
}

// MustInitTables 初始化数据表，失败则 panic
func MustInitTables(db *sql.DB) {
	if err := InitTables(db); err != nil {
		log.Fatalln("InitTables:", err)
	}
}
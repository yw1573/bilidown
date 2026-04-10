package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"bilidown/internal/service"
	"bilidown/internal/store"
	"bilidown/internal/util"
)

// CreateTask 创建任务
func CreateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}
	var body []store.TaskInitOption
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	// 异步处理任务创建
	go func() {
		db := store.MustGetDB()
		defer db.Close()
		for _, item := range body {
			if !util.CheckBvidFormat(item.Bvid) {
				log.Printf("bvid 格式错误: %s", item.Bvid)
				continue
			}
			if item.Cover == "" || item.Title == "" || item.Owner == "" {
				log.Printf("参数错误: bvid=%s", item.Bvid)
				continue
			}

			if !util.IsValidURL(item.Cover) {
				log.Printf("封面链接格式错误: bvid=%s", item.Bvid)
				continue
			}
			if !util.IsValidURL(item.Audio) {
				log.Printf("音频链接格式错误: bvid=%s", item.Bvid)
				continue
			}
			if !util.IsValidURL(item.Video) {
				log.Printf("视频链接格式错误: bvid=%s", item.Bvid)
				continue
			}
			if !util.IsValidFormatCode(item.Format) {
				log.Printf("清晰度代码错误: bvid=%s", item.Bvid)
				continue
			}
			baseFolder, err := store.GetCurrentFolder(db)
			if err != nil {
				log.Printf("store.GetCurrentFolder: %v", err)
				continue
			}
			// 如果有子目录，创建并拼接路径
			if item.Subfolder != "" {
				item.Subfolder = util.FilterFileName(item.Subfolder)
				item.Folder = filepath.Join(baseFolder, item.Subfolder)
				if err := os.MkdirAll(item.Folder, os.ModePerm); err != nil {
					log.Printf("创建子目录失败: %v", err)
					continue
				}
			} else {
				item.Folder = baseFolder
			}
			item.Status = store.TaskStatusWaiting
			_task := service.NewTask(&item)
			_task.Title = util.FilterFileName(_task.Title)
			err = _task.Create(db)
			if err != nil {
				log.Printf("_task.Create: %v", err)
				continue
			}
			go _task.Start()
		}
	}()

	util.Res{Success: true, Message: "任务已提交，正在后台创建"}.Write(w)
}

// GetActiveTask 获取活动任务
func GetActiveTask(w http.ResponseWriter, r *http.Request) {
	util.Res{Success: true, Data: service.GlobalTaskList}.Write(w)
}

// GetTaskList 获取任务列表
func GetTaskList(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	db := store.MustGetDB()
	defer db.Close()
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 0
	}
	pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
	if err != nil {
		pageSize = 360
	}
	tasks, err := store.GetTaskList(db, page, pageSize)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: tasks}.Write(w)
}

// DeleteTask 删除任务
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.FormValue("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	db := store.MustGetDB()
	defer db.Close()

	_task, err := store.GetTask(db, taskID)
	if err == sql.ErrNoRows {
		util.Res{Success: true, Message: "数据库中没有该条记录，所以本次操作被忽略，可以算作成功。"}.Write(w)
		return
	}
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("store.GetTask: %v", err)}.Write(w)
		return
	}
	filePath := _task.FilePath()
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		util.Res{Success: false, Message: fmt.Sprintf("文件删除失败 os.Remove: %v", err)}.Write(w)
		return
	}

	err = store.DeleteTask(db, taskID)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("store.DeleteTask: %v", err)}.Write(w)
		return
	}
	util.Res{Success: true, Message: "删除成功"}.Write(w)
}

// CancelTask 取消任务（批量）
func CancelTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}
	var taskIDs []int64
	err := json.NewDecoder(r.Body).Decode(&taskIDs)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	for _, id := range taskIDs {
		service.CancelTask(id)
	}
	util.Res{Success: true, Message: "任务已取消"}.Write(w)
}
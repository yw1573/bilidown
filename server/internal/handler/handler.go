package handler

import (
	"net/http"

	"bilidown/internal/handler/login"
	"bilidown/internal/handler/setting"
	"bilidown/internal/handler/task"
	"bilidown/internal/handler/video"
)

// API 返回路由器
func API() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/getVideoInfo", video.GetVideoInfo)
	router.HandleFunc("/getSeasonInfo", video.GetSeasonInfo)
	router.HandleFunc("/getQRInfo", login.GetQRInfo)
	router.HandleFunc("/getQRStatus", login.GetQRStatus)
	router.HandleFunc("/checkLogin", login.CheckLogin)
	router.HandleFunc("/getPlayInfo", video.GetPlayInfo)
	router.HandleFunc("/createTask", task.CreateTask)
	router.HandleFunc("/getActiveTask", task.GetActiveTask)
	router.HandleFunc("/getTaskList", task.GetTaskList)
	router.HandleFunc("/getFields", setting.GetFields)
	router.HandleFunc("/saveFields", setting.SaveFields)
	router.HandleFunc("/logout", login.Logout)
	router.HandleFunc("/quit", setting.Quit)
	router.HandleFunc("/getPopularVideos", video.GetPopularVideos)
	router.HandleFunc("/deleteTask", task.DeleteTask)
	router.HandleFunc("/getRedirectedLocation", video.GetRedirectedLocation)
	router.HandleFunc("/downloadVideo", video.DownloadVideo)
	router.HandleFunc("/getSeasonsArchivesListFirstBvid", video.GetSeasonsArchivesListFirstBvid)
	router.HandleFunc("/getFavList", video.GetFavList)
	router.HandleFunc("/checkFFmpeg", setting.CheckFFmpeg)
	return router
}
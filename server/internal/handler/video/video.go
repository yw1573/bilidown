package video

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"bilidown/internal/bilibili"
	"bilidown/internal/store"
	"bilidown/internal/util"
	"bilidown/internal/util/res_error"
)

// GetVideoInfo 获取视频信息
func GetVideoInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	bvid := r.FormValue("bvid")
	if !util.CheckBvidFormat(bvid) {
		res_error.Send(w, res_error.BvidFormatError)
		return
	}
	db := store.MustGetDB()
	defer db.Close()

	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	videoInfo, err := client.GetVideoInfo(bvid)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: videoInfo}.Write(w)
}

// GetSeasonInfo 获取剧集信息
func GetSeasonInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	var epid int
	epid, err := strconv.Atoi(r.FormValue("epid"))
	if r.FormValue("epid") != "" && err != nil {
		util.Res{Success: false, Message: "epid 格式错误"}.Write(w)
		return
	}
	var ssid int
	if epid == 0 {
		ssid, err = strconv.Atoi(r.FormValue("ssid"))
		if r.FormValue("ssid") != "" && err != nil {
			util.Res{Success: false, Message: "ssid 格式错误"}.Write(w)
			return
		}
	}
	db := store.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}

	client := bilibili.BiliClient{SESSDATA: sessdata}
	seasonInfo, err := client.GetSeasonInfo(epid, ssid)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: seasonInfo}.Write(w)
}

// GetPlayInfo 获取播放信息
func GetPlayInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	bvid := r.FormValue("bvid")
	if !util.CheckBvidFormat(bvid) {
		util.Res{Success: false, Message: "bvid 格式错误"}.Write(w)
		return
	}
	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		util.Res{Success: false, Message: "cid 格式错误"}.Write(w)
		return
	}
	db := store.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	playInfo, err := client.GetPlayInfo(bvid, cid)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("client.GetPlayInfo: %v", err)}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: playInfo}.Write(w)
}

// GetPopularVideos 获取热门视频
func GetPopularVideos(w http.ResponseWriter, r *http.Request) {
	db := store.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}

	client := bilibili.BiliClient{SESSDATA: sessdata}
	videos, err := client.GetPopularVideos()
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	bvidList := make([]string, 0)
	for _, v := range videos {
		bvidList = append(bvidList, v.Bvid)
	}
	util.Res{Success: true, Message: "获取成功", Data: bvidList}.Write(w)
}

// DownloadVideo 下载视频文件
var DownloadVideo = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	safePath := filepath.Clean(path)
	safePath = strings.ReplaceAll(safePath, "\\", "/")
	http.ServeFile(w, r, safePath)
})

// GetSeasonsArchivesListFirstBvid 获取合集第一个视频的BV号
var GetSeasonsArchivesListFirstBvid = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var mid int
	var seasonId int
	var err error
	if mid, err = strconv.Atoi(r.URL.Query().Get("mid")); err != nil {
		res_error.Send(w, res_error.MidFormatError)
		return
	}
	if seasonId, err = strconv.Atoi(r.URL.Query().Get("seasonId")); err != nil {
		res_error.Send(w, res_error.SeasonIdFormatError)
		return
	}
	client := bilibili.BiliClient{}
	bvid, err := client.GetSeasonsArchivesListFirstBvid(mid, seasonId)
	if err != nil {
		res_error.Send(w, fmt.Sprintf("client.GetSeasonsArchivesList: %v", err))
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: bvid}.Write(w)
})

// GetFavList 获取收藏夹列表
var GetFavList = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mediaId, err := strconv.Atoi(r.URL.Query().Get("mediaId"))
	if err != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	db := store.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	favList, err := client.GetFavlist(mediaId)
	if err != nil {
		res_error.Send(w, err.Error())
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: favList}.Write(w)
})

// GetRedirectedLocation 获取重定向地址
func GetRedirectedLocation(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	url := r.FormValue("url")
	if !util.IsValidURL(url) {
		res_error.Send(w, res_error.URLFormatError)
		return
	}
	if location, err := util.GetRedirectedLocation(url); err != nil {
		res_error.Send(w, res_error.NoLocationError)
		return
	} else {
		util.Res{Success: true, Message: "获取成功", Data: location}.Write(w)
		return
	}
}
package service

import (
	"bilidown/internal/bilibili"
	"bilidown/internal/store"
)

// GetVideoInfo 获取视频信息
func GetVideoInfo(client *bilibili.BiliClient, bvid string) (*bilibili.VideoInfo, error) {
	return client.GetVideoInfo(bvid)
}

// GetSeasonInfo 获取剧集信息
func GetSeasonInfo(client *bilibili.BiliClient, epid int, ssid int) (*bilibili.SeasonInfo, error) {
	return client.GetSeasonInfo(epid, ssid)
}

// GetPlayInfo 获取播放信息
func GetPlayInfo(client *bilibili.BiliClient, bvid string, cid int) (*bilibili.PlayInfo, error) {
	return client.GetPlayInfo(bvid, cid)
}

// GetPopularVideos 获取热门视频
func GetPopularVideos(client *bilibili.BiliClient) ([]bilibili.VideoInfo, error) {
	return client.GetPopularVideos()
}

// GetSeasonsArchivesListFirstBvid 获取合集第一个视频的BV号
func GetSeasonsArchivesListFirstBvid(client *bilibili.BiliClient, mid int, seasonId int) (string, error) {
	return client.GetSeasonsArchivesListFirstBvid(mid, seasonId)
}

// GetFavlist 获取收藏夹列表
func GetFavlist(client *bilibili.BiliClient, mediaId int) (*bilibili.FavList, error) {
	return client.GetFavlist(mediaId)
}

// NewTask 创建新任务
func NewTask(opt *store.TaskInitOption) *Task {
	return &Task{
		TaskInDB: store.TaskInDB{
			TaskInitOption: *opt,
		},
	}
}
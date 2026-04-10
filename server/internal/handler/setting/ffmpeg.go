package setting

import (
	"net/http"

	"bilidown/internal/util"
)

// CheckFFmpeg 检查 ffmpeg 是否可用
func CheckFFmpeg(w http.ResponseWriter, r *http.Request) {
	available := true
	version := ""

	if _, err := util.GetFFmpegPath(); err != nil {
		available = false
	} else {
		version = util.GetFFmpegVersion()
	}

	util.Res{
		Success: true,
		Data: map[string]interface{}{
			"available": available,
			"version":   version,
		},
	}.Write(w)
}
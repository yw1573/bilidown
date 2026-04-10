package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"bilidown/internal/handler"
	"bilidown/internal/static"
	"bilidown/internal/store"
	"bilidown/internal/util"
)

const (
	HTTP_PORT = 8098      // HTTP 服务器端口
	HTTP_HOST = ""        // HTTP 服务器主机
	VERSION   = "v2.1.0" // 软件版本号
)

var urlLocal = fmt.Sprintf("http://127.0.0.1:%d", HTTP_PORT)

// App 应用结构体
type App struct {
	Version string
	Port    int
}

// New 创建应用实例
func New() *App {
	return &App{
		Version: VERSION,
		Port:    HTTP_PORT,
	}
}

// Start 启动应用
func (a *App) Start() {
	// 检查 FFmpeg
	a.checkFFmpeg()

	// 初始化数据表
	db := store.MustGetDB()
	defer db.Close()
	store.MustInitTables(db)

	// 启动 HTTP 服务器
	a.startServer()

	// 打开浏览器
	OpenBrowser(urlLocal)

	// 保持运行
	select {}
}

// checkFFmpeg 检测 ffmpeg 的安装情况
func (a *App) checkFFmpeg() {
	if _, err := util.GetFFmpegPath(); err != nil {
		fmt.Println("🚨 FFmpeg is missing. Install it from https://www.ffmpeg.org/download.html or place it in ./bin, then restart the application.")
		select {}
	}
}

// startServer 启动 HTTP 服务器
func (a *App) startServer() {
	// 前端静态文件（嵌入二进制）
	staticFS, err := fs.Sub(static.Files, "ui")
	if err != nil {
		log.Fatal("static.Files:", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))
	// 后端接口服务
	http.Handle("/api/", http.StripPrefix("/api", handler.API()))
	// 启动 HTTP 服务器
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", HTTP_HOST, a.Port), nil)
		if err != nil {
			log.Fatal("http.ListenAndServe:", err)
		}
	}()
}

// OpenBrowser 调用系统默认浏览器打开指定 URL
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("OpenBrowser: %v.", errors.New("unsupported operating system"))
	}
	if err := cmd.Start(); err != nil {
		log.Printf("OpenBrowser: %v.", err)
	}
	fmt.Printf("Opened in default browser: %s.\n", url)
}

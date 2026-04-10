package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

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
	log.Printf("========================================")
	log.Printf("Bilidown %s 启动中...", a.Version)
	log.Printf("========================================")

	// 检查 FFmpeg
	a.checkFFmpeg()

	// 初始化数据表
	log.Printf("初始化数据库...")
	db := store.MustGetDB()
	defer db.Close()
	store.MustInitTables(db)
	log.Printf("数据库初始化完成")

	// 启动 HTTP 服务器
	a.startServer()

	// 打开浏览器
	log.Printf("打开浏览器...")
	OpenBrowser(urlLocal)

	log.Printf("========================================")
	log.Printf("Bilidown %s 已启动", a.Version)
	log.Printf("访问地址: %s", urlLocal)
	log.Printf("按 Ctrl+C 关闭软件")
	log.Printf("========================================")

	// 保持运行
	select {}
}

// checkFFmpeg 检测 ffmpeg 的安装情况
func (a *App) checkFFmpeg() {
	log.Printf("检查 FFmpeg...")
	path, err := util.GetFFmpegPath()
	if err != nil {
		log.Printf("❌ FFmpeg 未安装，请从 https://www.ffmpeg.org/download.html 下载安装")
		log.Printf("安装后重新启动软件")
		select {}
	}
	version := util.GetFFmpegVersion()
	log.Printf("✓ FFmpeg 已安装: %s (版本: %s)", path, version)
}

// startServer 启动 HTTP 服务器
func (a *App) startServer() {
	log.Printf("启动 HTTP 服务器 (端口: %d)...", a.Port)

	// 前端静态文件（嵌入二进制）
	staticFS, err := fs.Sub(static.Files, "ui")
	if err != nil {
		log.Fatal("static.Files:", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))
	// 后端接口服务
	http.Handle("/api/", http.StripPrefix("/api", handler.API()))

	// 添加请求日志中间件
	loggedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		http.DefaultServeMux.ServeHTTP(w, r)
		log.Printf("[%s] %s %s (%v)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})

	// 启动 HTTP 服务器
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", HTTP_HOST, a.Port), loggedHandler)
		if err != nil {
			log.Fatal("http.ListenAndServe:", err)
		}
	}()

	log.Printf("✓ HTTP 服务器已启动")
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
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("打开浏览器失败: %v", err)
	} else {
		log.Printf("✓ 浏览器已打开: %s", url)
	}
}

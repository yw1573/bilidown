package main

import (
	"log"
	"os"

	"bilidown/internal/app"
)

func init() {
	// 设置日志格式：日期 时间 文件:行号
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func main() {
	application := app.New()
	application.Start()
}

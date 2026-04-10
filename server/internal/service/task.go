package service

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"bilidown/internal/bilibili"
	"bilidown/internal/common"
	"bilidown/internal/store"
	"bilidown/internal/util"
)

// Task 任务结构体
type Task struct {
	store.TaskInDB
	AudioProgress float64 `json:"audioProgress"`
	VideoProgress float64 `json:"videoProgress"`
	MergeProgress float64 `json:"mergeProgress"`
}

// GlobalTaskList 全局任务列表
var GlobalTaskList = []*Task{}

// GlobalTaskMux 全局任务锁
var GlobalTaskMux = &sync.Mutex{}

// GlobalDownloadSem 下载并发控制
var GlobalDownloadSem = util.NewSemaphore(3)

// GlobalMergeSem 合并并发控制
var GlobalMergeSem = util.NewSemaphore(3)

// Create 创建任务记录
func (task *Task) Create(db *sql.DB) error {
	id, err := store.CreateTask(db, &task.TaskInitOption)
	if err != nil {
		return err
	}
	task.ID = id
	task.CreateAt = time.Now()
	return nil
}

// Start 启动任务
func (task *Task) Start() {
	if task.DownloadType == "" {
		task.DownloadType = "merge"
	}
	GlobalTaskMux.Lock()
	GlobalTaskList = append(GlobalTaskList, task)
	GlobalTaskMux.Unlock()
	db := store.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil {
		task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("bilibili.GetSessdata: %v", err))
		return
	}
	client := &bilibili.BiliClient{SESSDATA: sessdata}

	GlobalDownloadSem.Acquire()
	task.UpdateStatus(db, store.TaskStatusRunning)

	if task.DownloadType == "audio" {
		// 仅音频模式：只下载音频，重命名音频文件为输出文件
		err = DownloadMedia(client, task.Audio, task, "audio")
		if err != nil {
			GlobalDownloadSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("DownloadMedia: %v", err))
			return
		}
		GlobalDownloadSem.Release()
		outputPath := task.TaskInDB.FilePath()
		audioPath := filepath.Join(task.Folder, strconv.FormatInt(task.ID, 10)+".audio")
		err = os.Rename(audioPath, outputPath)
		if err != nil {
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("os.Rename: %v", err))
			return
		}
		task.UpdateStatus(db, store.TaskStatusDone)
		return
	} else if task.DownloadType == "video" {
		// 仅视频模式：只下载视频，重命名视频文件为输出文件
		err = DownloadMedia(client, task.Video, task, "video")
		if err != nil {
			GlobalDownloadSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("DownloadMedia: %v", err))
			return
		}
		GlobalDownloadSem.Release()
		outputPath := task.TaskInDB.FilePath()
		videoPath := filepath.Join(task.Folder, strconv.FormatInt(task.ID, 10)+".video")
		err = os.Rename(videoPath, outputPath)
		if err != nil {
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("os.Rename: %v", err))
			return
		}
		task.UpdateStatus(db, store.TaskStatusDone)
		return
	} else {
		// 合并模式：下载音频和视频，然后合并
		err = DownloadMedia(client, task.Audio, task, "audio")
		if err != nil {
			GlobalDownloadSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("DownloadMedia: %v", err))
			return
		}
		err = DownloadMedia(client, task.Video, task, "video")
		if err != nil {
			GlobalDownloadSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("DownloadMedia: %v", err))
			return
		}
		GlobalDownloadSem.Release()

		outputPath := task.TaskInDB.FilePath()
		videoPath := filepath.Join(task.Folder, strconv.FormatInt(task.ID, 10)+".video")
		audioPath := filepath.Join(task.Folder, strconv.FormatInt(task.ID, 10)+".audio")
		GlobalMergeSem.Acquire()
		err = task.MergeMedia(outputPath, videoPath, audioPath)
		if err != nil {
			GlobalMergeSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("task.MergeMedia: %v", err))
			return
		}
		err = os.Remove(videoPath)
		if err != nil {
			GlobalMergeSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("os.Remove: %v", err))
			return
		}
		err = os.Remove(audioPath)
		if err != nil {
			GlobalMergeSem.Release()
			task.UpdateStatus(db, store.TaskStatusError, fmt.Errorf("os.Remove: %v", err))
			return
		}
		GlobalMergeSem.Release()
		task.UpdateStatus(db, store.TaskStatusDone)
	}
}

// UpdateStatus 更新任务状态
func (task *Task) UpdateStatus(db *sql.DB, status store.TaskStatus, errs ...error) error {
	if err := store.UpdateTaskStatus(db, task.ID, status); err != nil {
		return err
	}
	for _, err := range errs {
		if err != nil {
			if logErr := store.CreateLog(db, fmt.Sprintf("Task-%d-Error: %v", task.ID, err)); logErr != nil {
				log.Fatalln("CreateLog:", logErr)
			}
		}
	}
	task.Status = status
	return nil
}

// MergeMedia 合并音视频
func (task *Task) MergeMedia(outputPath string, inputPaths ...string) error {
	inputs := []string{}
	for _, path := range inputPaths {
		inputs = append(inputs, "-i", path)
	}

	ffmpegPath, err := util.GetFFmpegPath()
	if err != nil {
		return err
	}

	cmd := exec.Command(ffmpegPath, append(inputs, "-c:v", "copy", "-c:a", "copy", "-progress", "pipe:1", "-strict", "-2", outputPath)...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)

	progress := newProgressBar(int64(task.Duration))
	outTimeRegex := regexp.MustCompile(`out_time_ms=(\d+)`) // 毫秒

	for scanner.Scan() {
		line := scanner.Text()
		match := outTimeRegex.FindStringSubmatch(line)
		if len(match) == 2 {
			outTime, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				return err
			}
			progress.current = outTime / 1000000
			task.MergeProgress = progress.percent()
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	task.MergeProgress = 1
	return nil
}

// progressBar 进度条
type progressBar struct {
	total   int64
	current int64
}

func (p *progressBar) add(n int) {
	p.current += int64(n)
}

func (p *progressBar) percent() float64 {
	return float64(p.current) / float64(p.total)
}

func newProgressBar(total int64) *progressBar {
	return &progressBar{
		total: total,
	}
}

// DownloadMedia 下载媒体文件
func DownloadMedia(client *bilibili.BiliClient, _url string, task *Task, mediaType string) error {
	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ {
		resp, err = client.SimpleGET(_url, nil)
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	filename := strconv.FormatInt(task.ID, 10) + "." + mediaType
	filepath := filepath.Join(task.Folder, filename)

	progress := newProgressBar(resp.ContentLength)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := io.TeeReader(resp.Body, file)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		progress.add(n)
		GlobalTaskMux.Lock()
		if mediaType == "video" {
			task.VideoProgress = progress.percent()
		} else {
			task.AudioProgress = progress.percent()
		}
		GlobalTaskMux.Unlock()
	}
	return nil
}

// GetVideoURL 获取视频URL
func GetVideoURL(medias []bilibili.Media, format common.MediaFormat) (string, error) {
	for _, code := range []int{12, 7, 13} {
		for _, item := range medias {
			if item.ID == format && item.Codecid == code {
				return item.BaseURL, nil
			}
		}
	}
	return "", errors.New("未找到对应视频分辨率格式")
}

// GetAudioURL 获取音频URL
func GetAudioURL(dash *bilibili.Dash) string {
	if dash.Flac != nil {
		return dash.Flac.Audio.BaseURL
	}
	var maxAudioID common.MediaFormat
	var audioURL string
	for _, item := range dash.Audio {
		if item.ID > maxAudioID {
			maxAudioID = item.ID
			audioURL = item.BaseURL
		}
	}
	return audioURL
}

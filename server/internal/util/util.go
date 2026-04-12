package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"bilidown/internal/common"
)

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	return fmt.Sprintf("%x", randomBytes)[:length]
}

func CheckBvidFormat(bvid string) bool {
	return regexp.MustCompile("^BV1[a-zA-Z0-9]+").MatchString(bvid)
}

// GetDefaultDownloadFolder 获取默认下载路径
func GetDefaultDownloadFolder() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)
	downloadDir := filepath.Join(execDir, "download")
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		return "", err
	}
	return downloadDir, nil
}

func IsNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

// IsValidURL 判断字符串是否为合法的URL
func IsValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	return err == nil
}

// IsValidFormatCode 判断格式码是否合法
func IsValidFormatCode(format common.MediaFormat) bool {
	allowed := []common.MediaFormat{6, 16, 32, 64, 74, 80, 112, 116, 120, 125, 126, 127}
	for _, v := range allowed {
		if v == format {
			return true
		}
	}
	return false
}

// FilterFileName 过滤字符串中的特殊字符，使其允许作为文件名。
func FilterFileName(fileName string) string {
	return regexp.MustCompile(`[\\/:*?"<>|\n]`).ReplaceAllString(fileName, "")
}

// GetFFmpegPath 获取可用的 FFmpeg 执行路径。
func GetFFmpegPath() (string, error) {
	if err := exec.Command("ffmpeg", "-version").Run(); err == nil {
		return "ffmpeg", nil
	}
	return "", errors.New("ffmpeg not found")
}

// GetFFmpegVersion 获取 FFmpeg 版本信息
func GetFFmpegVersion() string {
	cmd := exec.Command("ffmpeg", "-version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	// 解析版本号，格式如: ffmpeg version 6.0 ...
	re := regexp.MustCompile(`ffmpeg version (\S+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetRedirectedLocation 获取响应头中的 Location，不会自动跟随重定向。
func GetRedirectedLocation(url string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if locationURL, err := response.Location(); err != nil {
		return "", err
	} else {
		return locationURL.String(), nil
	}
}

func MD5Hash(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	return hashString
}

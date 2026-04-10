package service

import (
	"database/sql"

	"bilidown/internal/bilibili"
)

// CheckLogin 检查登录状态
func CheckLogin(client *bilibili.BiliClient) (bool, error) {
	return client.CheckLogin()
}

// NewQRInfo 获取二维码信息
func NewQRInfo(client *bilibili.BiliClient) (*bilibili.QRInfo, error) {
	return client.NewQRInfo()
}

// GetQRStatus 获取二维码状态
func GetQRStatus(client *bilibili.BiliClient, qrKey string) (*bilibili.QRStatus, string, error) {
	return client.GetQRStatus(qrKey)
}

// SaveSessdata 保存 SESSDATA
func SaveSessdata(db *sql.DB, sessdata string) error {
	return bilibili.SaveSessdata(db, sessdata)
}

// GetSessdata 获取 SESSDATA
func GetSessdata(db *sql.DB) (string, error) {
	return bilibili.GetSessdata(db)
}
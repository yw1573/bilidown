package setting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bilidown/internal/store"
	"bilidown/internal/util"
)

// GetFields 获取字段
func GetFields(w http.ResponseWriter, r *http.Request) {
	db := store.MustGetDB()
	defer db.Close()

	fields, err := store.GetFields(db, store.FieldUtil{}.AllowSelect()...)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Data: fields}.Write(w)
}

// SaveFields 保存字段
func SaveFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}
	defer r.Body.Close()
	var body [][2]string

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	db := store.MustGetDB()
	defer db.Close()

	fu := store.FieldUtil{}

	for _, d := range body {
		if !fu.IsAllowUpdate(d[0]) {
			util.Res{Success: false, Message: fmt.Sprintf("字段 %s 不允许修改", d[0])}.Write(w)
			return
		}

		if d[0] == "download_folder" {
			if _, err := os.Stat(d[1]); os.IsNotExist(err) {
				if err := os.MkdirAll(d[1], os.ModePerm); err != nil {
					util.Res{Success: false, Message: fmt.Sprintf("目录创建失败：%s", d[1])}.Write(w)
					return
				}
			} else if err != nil {
				util.Res{Success: false, Message: fmt.Sprintf("路径设置失败：%v", err)}.Write(w)
				return
			}
		}
	}

	err = store.SaveFields(db, body)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "保存成功"}.Write(w)
}

// Quit 退出应用
func Quit(w http.ResponseWriter, r *http.Request) {
	util.Res{Success: true, Message: "退出成功"}.Write(w)
	go func() {
		os.Exit(0)
	}()
}
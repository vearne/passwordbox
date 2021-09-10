package sc

import (
	"github.com/vearne/passwordbox/resource"
	slog "github.com/vearne/simplelog"
	"path/filepath"
)

func CompareAndUpload(fileName, fullPath string) {
	if resource.GlobalOSS == nil {
		return
	}
	key := filepath.Join(resource.GlobalOSS.GetDirPath(), fileName)
	newer, err := resource.GlobalOSS.Compare(key, fullPath)
	if err != nil {
		slog.Error("CompareAndDownload:%v", err)
		return
	}
	if newer >= 0 {
		slog.Debug("no need to upload")
	} else {
		slog.Info("upload, key:%v", key)
		resource.GlobalOSS.UploadFile(key, fullPath)
		err := resource.GlobalOSS.AdjustMTime(key, fullPath)
		if err != nil {
			slog.Error("GlobalOSS.AdjustMTime:%v", err)
		}
	}
}

func CompareAndDownloadAll() {
	if resource.GlobalOSS == nil {
		return
	}
	keys, err := resource.GlobalOSS.ListKeys(resource.GlobalOSS.GetDirPath())
	if err != nil {
		slog.Error("GlobalOSS.ListKeys, error:%v", err)
		return
	}
	for _, key := range keys {
		_, filename := filepath.Split(key)
		fullpath := filepath.Join(resource.DataPath, filename)
		newer, err := resource.GlobalOSS.Compare(key, fullpath)
		if err != nil {
			slog.Error("CompareAndDownload:%v", err)
			return
		}
		if newer > 0 {
			slog.Info("download, key:%v", key)
			resource.GlobalOSS.DownloadFile(key, fullpath)
			err := resource.GlobalOSS.AdjustMTime(key, fullpath)
			if err != nil {
				slog.Error("GlobalOSS.AdjustMTime:%v", err)
			}
		}
	}

}

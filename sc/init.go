package sc

import (
	"github.com/vearne/passwordbox/resource"
	slog "github.com/vearne/simplelog"
	"path/filepath"
)

var (
	GlobalOSS ObjectStorage
)

type ObjectStorage interface {
	Init() error
	GetDirPath() string
	UploadFile(key string, filepath string) bool
	DownloadFile(key string, filepath string) bool
	// If file in oss is newer than localfile?
	Compare(key string, localFilePath string) (bool, error)
	ListKeys(prefix string) ([]string, error)
}

func CompareAndUpload(fileName, fullPath string) {
	if GlobalOSS == nil {
		return
	}
	key := filepath.Join(GlobalOSS.GetDirPath(), fileName)
	newer, err := GlobalOSS.Compare(key, fullPath)
	if err != nil {
		slog.Error("CompareAndDownload:%v", err)
		return
	}
	if newer {
		slog.Debug("no need to upload")
		return
	} else {
		slog.Info("upload, key:%v", key)
		GlobalOSS.UploadFile(key, fullPath)
	}
}

func CompareAndDownloadAll() {
	if GlobalOSS == nil {
		return
	}
	keys, err := GlobalOSS.ListKeys(GlobalOSS.GetDirPath())
	if err != nil {
		slog.Error("GlobalOSS.ListKeys, error:%v", err)
		return
	}
	for _, key := range keys {
		_, filename := filepath.Split(key)
		fullpath := filepath.Join(resource.DataPath, filename)
		newer, err := GlobalOSS.Compare(key, fullpath)
		if err != nil {
			slog.Error("CompareAndDownload:%v", err)
			return
		}
		if newer {
			slog.Info("download, key:%v", key)
			GlobalOSS.DownloadFile(key, fullpath)
		}
	}

}

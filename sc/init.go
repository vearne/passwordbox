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
	UploadFile(key string, localFilePath string) bool
	DownloadFile(key string, localFilePath string) bool
	// If file in oss is newer than localfile?
	// if newer > 0, file in oss is newer than localfile
	// if newer == 0 file in oss is new as localfile
	// if newer < 0 localfile is newer than file in oss
	Compare(key string, localFilePath string) (newer int, err error)
	// modify Mtime of local file to consistent with file in oss
	AdjustMTime(key string, localFilePath string) error
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
	if newer >= 0 {
		slog.Debug("no need to upload")
	} else {
		slog.Info("upload, key:%v", key)
		GlobalOSS.UploadFile(key, fullPath)
		err := GlobalOSS.AdjustMTime(key, fullPath)
		if err != nil {
			slog.Error("GlobalOSS.AdjustMTime:%v", err)
		}
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
		if newer > 0 {
			slog.Info("download, key:%v", key)
			GlobalOSS.DownloadFile(key, fullpath)
			err := GlobalOSS.AdjustMTime(key, fullpath)
			if err != nil {
				slog.Error("GlobalOSS.AdjustMTime:%v", err)
			}
		}
	}

}

package model

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

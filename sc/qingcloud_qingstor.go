package sc

import (
	slog "github.com/vearne/simplelog"
	"github.com/yunify/qingstor-sdk-go/config"
	qs "github.com/yunify/qingstor-sdk-go/service"
	"io"
	"net/http"
	"os"
	"strings"
)

type QingStor struct {
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	BucketName string `mapstructure:"bucket_name"`
	Zone       string `mapstructure:"zone"`
	DirPath    string `mapstructure:"dir_path"`
	Bucket     *qs.Bucket
}

func (s *QingStor) Init() error {
	var err error
	configuration, err := config.New(s.AccessKey, s.SecretKey)
	if err != nil {
		return err
	}
	qsService, err := qs.Init(configuration)
	if err != nil {
		return err
	}
	s.Bucket, err = qsService.Bucket(s.BucketName, s.Zone)
	if err != nil {
		return err
	}
	return nil
}

func (s *QingStor) GetDirPath() string {
	return s.DirPath
}

func (s *QingStor) UploadFile(key string, filepath string) bool {
	// Open file
	var file *os.File
	file, err := os.Open(filepath)
	if err != nil {
		slog.Error("QingStor.UploadFile--open file error,filepath:%v", filepath)
		return false
	}
	defer file.Close()

	// Put object
	oOutput, err := s.Bucket.PutObject(key, &qs.PutObjectInput{Body: file})

	if qs.IntValue(oOutput.StatusCode) == http.StatusCreated {
		// Print the HTTP status code.
		// Example: 201
		return true
	} else if err != nil {
		// Example: QingStor Error: StatusCode 403, Code "permission_denied"...
		slog.Error("QingStor.UploadFile--error,filepath:%v", filepath)
		return false
	}
	return false
}

func (s *QingStor) ListKeys(prefix string) ([]string, error) {
	bOutput, err := s.Bucket.ListObjects(&qs.ListObjectsInput{Prefix: &s.DirPath})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, item := range bOutput.Keys {
		result = append(result, *item.Key)
	}
	return result, nil
}

func (s *QingStor) DownloadFile(key string, logFilePath string) bool {
	getOutput, err := s.Bucket.GetObject(key,
		&qs.GetObjectInput{},
	)
	if err != nil {
		slog.Error("DownloadFile error, %v", err)
		return false
	}
	defer getOutput.Close()
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		slog.Error("DownloadFile-open file error, %v", err)
		return false
	}
	defer f.Close()
	_, err = io.Copy(f, getOutput.Body)

	if err != nil {
		slog.Error("DownloadFile-copy error, %v", err)
		return false
	}
	return true
}

func (s *QingStor) Compare(key string, localFilePath string) (bool, error) {
	remote, err := s.Bucket.HeadObject(key, nil)
	if err != nil && strings.Index(err.Error(), "404") != -1 {
		return false, nil
	} else if err != nil {
		slog.Error("Bucket.HeadObject error, %v", err)
		return false, err
	}

	info, err := os.Stat(localFilePath)
	if err != nil {
		//slog.Error("os.Stat error, %v", err)
		return true, nil
	}
	localSize := info.Size()
	localLastModified := info.ModTime()
	if localSize != *remote.ContentLength && (*remote.LastModified).After(localLastModified) {
		return true, nil
	}
	return false, nil
}

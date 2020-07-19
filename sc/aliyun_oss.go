package sc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	slog "github.com/vearne/simplelog"
	"os"
)

type AliOSS struct {
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
	Endpoint        string `mapstructure:"endpoint"`
	DirPath         string `mapstructure:"dir_path"`
	Bucket          *oss.Bucket
}

func (s *AliOSS) Init() error {
	client, err := oss.New(s.Endpoint, s.AccessKeyId, s.AccessKeySecret)
	if err != nil {
		return err
	}
	s.Bucket, err = client.Bucket(s.BucketName)
	if err != nil {
		return err
	}
	return nil
}

func (s *AliOSS) GetDirPath() string {
	return s.DirPath
}

func (s *AliOSS) UploadFile(key string, localFilePath string) bool {
	err := s.Bucket.PutObjectFromFile(key, localFilePath)
	if err != nil {
		slog.Error("AliOSS.UploadFile, error:%v", err)
		return false
	}
	return true
}

func (s *AliOSS) ListKeys(prefix string) ([]string, error) {
	lsRes, err := s.Bucket.ListObjects(oss.Prefix(prefix))
	if err != nil {
		slog.Error("AliOSS.ListKeys, error:%v", err)
		return nil, err
	}

	result := make([]string, 0)
	for _, object := range lsRes.Objects {
		result = append(result, object.Key)
	}
	return result, nil
}

func (s *AliOSS) DownloadFile(key string, localFilePath string) bool {
	err := s.Bucket.GetObjectToFile(key, localFilePath)
	if err != nil {
		slog.Error("AliOSS.DownloadFile, error:%v", err)
		return false
	}
	return true
}

func (s *AliOSS) Compare(key string, localFilePath string) (bool, error) {
	lsRes, err := s.Bucket.ListObjects(oss.Prefix(key))
	if err != nil {
		slog.Error("AliOSS.ListKeys, error:%v", err)
		return false, err
	}

	if len(lsRes.Objects) <= 0 {
		return false, nil
	}

	var obj oss.ObjectProperties
	for _, object := range lsRes.Objects {
		if key == object.Key {
			obj = object
			break
		}
	}

	info, err := os.Stat(localFilePath)
	if err != nil {
		return true, nil
	}

	localLastModified := info.ModTime()
	if obj.LastModified.Unix() > localLastModified.Unix() {
		return true, nil
	}
	return false, nil

}

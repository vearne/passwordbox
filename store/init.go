package store

import (
	"crypto/aes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vearne/passwordbox/consts"
	slog "github.com/vearne/passwordbox/log"
	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	GlobalStore *DatabaseStore
)

type DatabaseStore struct {
	DatabaseName string
	Hint         string
	Items        []*model.SimpleItem
	FullPath     string
	Key          []byte
	DataBaseIV   string
	DB           *sql.DB
	TempFile     string
}

func NewDatabaseStore(dataPath string, database *model.Database) *DatabaseStore {
	databaseName := strings.TrimSpace(database.Name)
	filename := utils.Sha256N(databaseName, consts.HashCount)
	fullpath := filepath.Join(dataPath, filename)

	s := DatabaseStore{}
	s.FullPath = fullpath
	s.DatabaseName = database.Name
	s.DataBaseIV = filename[0:aes.BlockSize]
	s.Key = utils.GenHMacKey([]byte(database.Password), []byte(s.DataBaseIV))
	s.Hint = database.Hint
	s.Items = make([]*model.SimpleItem, 0)

	return &s
}

func OpenDatabaseStore(dataPath string, database *model.Database) (*DatabaseStore, error) {
	databaseName := strings.TrimSpace(database.Name)
	filename := utils.Sha256N(databaseName, consts.HashCount)
	fullpath := filepath.Join(dataPath, filename)

	s := DatabaseStore{}
	s.FullPath = fullpath
	s.DatabaseName = database.Name
	s.DataBaseIV = filename[0:aes.BlockSize]
	s.Key = utils.GenHMacKey([]byte(database.Password), []byte(s.DataBaseIV))

	// copy disk file to tempfile
	// create temp file
	file, _ := ioutil.TempFile("", "*")
	err := file.Close()
	if err != nil {
		slog.Fatal("DatabaseStore-create temp file, %v", err)
		return nil, err
	}
	s.TempFile = file.Name()
	buff, err := ioutil.ReadFile(s.FullPath)
	if err != nil {
		slog.Error("open openDatabase error, %v", err)
		return nil, err
	}
	// Decrypt the entire file
	buff = utils.DecryptAesInCFB(buff, s.Key, []byte(s.DataBaseIV))
	err = ioutil.WriteFile(s.TempFile, buff, 0600)
	if err != nil {
		slog.Error("open openDatabase error, %v", err)
		return nil, err
	}
	s.DB, err = sql.Open("sqlite3", s.TempFile)
	if err != nil {
		slog.Error("DatabaseStore-open db, %v", err)
		return nil, err
	}
	return &s, nil
}

func (s *DatabaseStore) Init() error {
	var err error
	file, _ := ioutil.TempFile("", "*")
	err = file.Close()
	if err != nil {
		slog.Fatal("DatabaseStore-create temp file, %v", err)
		return err
	}
	s.TempFile = file.Name()
	s.DB, err = sql.Open("sqlite3", s.TempFile)
	if err != nil {
		slog.Error("DatabaseStore-open db, %v", err)
		return err
	}
	err = CreateTable(s.DB)
	if err != nil {
		slog.Error("DatabaseStore-operate db, %v", err)
		return err
	}
	err = InsertHint(s.DB, s.Hint)
	if err != nil {
		slog.Error("DatabaseStore-operate db, %v", err)
		return err
	}
	return nil
}

func (s *DatabaseStore) Close() error {
	var err error
	// close sqlite db
	err = s.DB.Close()
	if err != nil {
		slog.Error("close file error, %v", err)
		return err
	}
	// flush to disk
	buff, err := ioutil.ReadFile(s.TempFile)
	if err != nil {
		slog.Error("read temp file error, %v", err)
		return err
	}

	// Encrypt the entire file
	// ciphertext = AES-CFB(fileConent, key, DataBaseIV)
	buff = utils.EncryptAesInCFB(buff, s.Key, []byte(s.DataBaseIV))
	err = ioutil.WriteFile(s.FullPath, buff, 0600)
	if err != nil {
		slog.Error("write disk file error, %v", err)
		return err
	}

	// remove temp file
	err = os.Remove(s.TempFile)
	if err != nil {
		slog.Error("remove temp file error, %v", err)
		return err
	}
	return nil
}

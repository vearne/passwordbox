package store

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/resource"
	slog "github.com/vearne/simplelog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RestoreItem(c *cli.Context) error {
	fmt.Println("--RestoreItem--")
	tagId := c.Int("tagId")
	slog.Debug("RestoreItem, tagId:%v", tagId)

	if tagId < 0 { // list all available backups
		slog.Debug("resource.DataPath:%v", resource.DataPath)
		items := getAllBackupItem()
		PrintBackups(items)
		return nil
	}
	// restore
	items := getAllBackupItem()
	if tagId > len(items) {
		slog.Error("tagId invalid")
		return nil
	}
	confirmRestore := false
	prompt := &survey.Confirm{
		Message: "confirm restore?",
	}
	err := survey.AskOne(prompt, &confirmRestore)
	if err != nil {
		fmt.Printf("survey.AskOne error, %v\n", err)
		return err
	}
	if !confirmRestore {
		return nil
	}

	slog.Info("1. RestoreItem-close DB")
	GlobalStore.Close()
	// delete
	err = os.Remove(GlobalStore.FullPath)
	if err != nil {
		slog.Error("os.Remove:%v", GlobalStore.FullPath)
		return err
	}
	oldName := filepath.Join(resource.DataPath, GlobalStore.FileName+"."+items[tagId-1].Tag)
	newName := GlobalStore.FullPath
	// rename
	slog.Info("2. RestoreItem-rename, oldName:%v, newName:%v", oldName, newName)
	err = os.Rename(oldName, newName)
	if err != nil {
		slog.Error("os.Rename:%v", oldName)
		return err
	}
	// upload
	key := filepath.Join(resource.GlobalOSS.GetDirPath(), GlobalStore.FileName)
	slog.Info("3. RestoreItem-upload, key:%v", key)
	resource.GlobalOSS.UploadFile(key, GlobalStore.FullPath)

	slog.Info("Restore success.Please login later...")
	resource.LoopExit = true
	return nil
}

func getAllBackupItem() []model.BackupItem {
	items := make([]model.BackupItem, 0)
	files, err := getAllBackupFiles(resource.DataPath, GlobalStore.FileName)
	if err != nil {
		slog.Error("RestoreItem-GetAllBackupFiles, %v", err)
		return items
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	counter := 1
	for _, fileName := range files {
		tempList := strings.Split(fileName, ".")
		if len(tempList) < 2 {
			continue
		}
		items = append(items, model.BackupItem{ID: counter, Tag: tempList[1]})
		counter++
	}
	return items
}

func getAllBackupFiles(dirPth string, prefix string) (files []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			continue
		} else {
			// 过滤指定格式
			ok := strings.HasPrefix(fi.Name(), prefix)
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	return files, nil
}

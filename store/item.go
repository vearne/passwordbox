package store

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli/v2"
	"github.com/vearne/passwordbox/consts"
	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/resource"
	"github.com/vearne/passwordbox/sc"
	"github.com/vearne/passwordbox/utils"
	slog "github.com/vearne/simplelog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"time"
)

func AddItem(c *cli.Context) error {
	fmt.Println("--AddItem--")
	var qs = []*survey.Question{
		{
			Name:     "title",
			Prompt:   &survey.Input{Message: "Please type Item's title:"},
			Validate: survey.Required,
		},
		{
			Name:     "account",
			Prompt:   &survey.Input{Message: "Please type Item's account:"},
			Validate: survey.Required,
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Please type Item's password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "comment",
			Prompt: &survey.Input{
				Message: "Please type Item's comment(optional):",
			},
		},
	}
	answers := model.DetailItem{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Printf("survey.Ask error, %v\n", err)
		return err
	}
	answers.ModifiedAt = time.Now().Format(time.RFC3339)
	PrintItems([]*model.DetailItem{&answers})

	err = InsertItem(GlobalStore.DB, ChangeToSimpleItem(&answers, GlobalStore.Key))
	if err != nil {
		fmt.Printf("InsertItem error,%v\n", err)
		return err
	}

	GlobalStore.Dirty = true
	fmt.Println("AddItem-save to file")

	err = SearchItem(c)
	if err != nil {
		fmt.Printf("SearchItem error,%v\n", err)
		return err
	}

	return nil
}

func DelItem(c *cli.Context) error {
	fmt.Println("--DeleteItem--")
	itemId := c.Int("itemId")

	item, err := GetItem(GlobalStore.DB, itemId)
	if err != nil {
		fmt.Printf("can't find %v\n", itemId)
		return nil
	}
	detailItem := ParseSimpleItem(item, GlobalStore.Key)
	PrintItems([]*model.DetailItem{detailItem})

	confirmDel := false
	prompt := &survey.Confirm{
		Message: "confirm delete?",
	}
	err = survey.AskOne(prompt, &confirmDel)
	if err != nil {
		fmt.Printf("survey.AskOne error, %v\n", err)
		return err
	}
	if confirmDel {
		GlobalStore.Dirty = true
		err = DeleteItem(GlobalStore.DB, itemId)
		if err != nil {
			fmt.Printf("delete item %v error, %v\n", itemId, err)
		} else {
			fmt.Printf("delete item %v success\n", itemId)
		}
	}

	// For user experience
	err = SearchItem(c)
	if err != nil {
		fmt.Printf("SearchItem error, %v\n", err)
		return err
	}
	return nil
}

func paddingStar(n int) string {
	buff := bytes.NewBuffer(make([]byte, 0))
	for i := 0; i < n; i++ {
		buff.Write([]byte("*"))
	}
	return buff.String()
}

func ModifyItem(c *cli.Context) error {
	fmt.Println("--ModifyItem--")
	itemId := c.Int("itemId")
	item, err := GetItem(GlobalStore.DB, itemId)
	if err != nil {
		fmt.Printf("can't find %v\n", itemId)
		return nil
	}
	detailItem := ParseSimpleItem(item, GlobalStore.Key)
	// These are using the default foreground colors
	color.Red("If you don't want to make changes, you can just press Enter!")
	password := paddingStar(len(detailItem.Password))
	var qs = []*survey.Question{
		{
			Name:   "title",
			Prompt: &survey.Input{Message: fmt.Sprintf("Please type Item's title:[%q]", detailItem.Title)},
		},
		{
			Name:   "account",
			Prompt: &survey.Input{Message: fmt.Sprintf("Please type Item's account:[%q]", detailItem.Account)},
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: fmt.Sprintf("Please type Item's password:[%q]", password),
			},
		},
		{
			Name: "comment",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("Please type Item's comment(optional):[%q]", detailItem.Comment),
			},
		},
	}
	answers := model.DetailItem{}

	// perform the questions
	err = survey.Ask(qs, &answers)
	if err != nil {
		slog.Error("survey error, %v", err)
		return err
	}
	dirty := false
	if len(answers.Title) > 0 {
		detailItem.Title = answers.Title
		dirty = true
	}
	if len(answers.Account) > 0 {
		detailItem.Account = answers.Account
		dirty = true
	}
	if len(answers.Password) > 0 {
		detailItem.Password = answers.Password
		dirty = true
	}
	if len(answers.Comment) > 0 {
		detailItem.Comment = answers.Comment
		dirty = true
	}

	if dirty {
		GlobalStore.Dirty = true
		detailItem.ModifiedAt = time.Now().Format(time.RFC3339)
		PrintItems([]*model.DetailItem{detailItem})
		err = UpdateItem(GlobalStore.DB, ChangeToSimpleItem(detailItem, GlobalStore.Key))
		if err != nil {
			fmt.Printf("UpdateItem error %v\n", err)
			return err
		}
	} else {
		color.Yellow("The item remains the same as before.")
		PrintItems([]*model.DetailItem{detailItem})
	}
	return nil
}

func ViewItem(c *cli.Context) error {
	fmt.Println("--ViewItem--")
	itemId := c.Int("itemId")
	item, err := GetItem(GlobalStore.DB, itemId)
	if err != nil {
		fmt.Printf("can't find %v\n", itemId)
		return nil
	}
	detailItem := ParseSimpleItem(item, GlobalStore.Key)
	PrintItems([]*model.DetailItem{detailItem})
	return nil
}

func OtpItem(c *cli.Context) error {
	fmt.Println("--OtpItem--")
	itemId := c.Int("itemId")
	item, err := GetItem(GlobalStore.DB, itemId)
	if err != nil {
		fmt.Printf("can't find %v\n", itemId)
		return nil
	}
	detailItem := ParseSimpleItem(item, GlobalStore.Key)
	key, err := otp.NewKeyFromURL(detailItem.Password)
	if err != nil {
		fmt.Printf("failed to parse url %v\n", detailItem.Password)
		return nil
	}
	switch key.Type() {
	case "totp":
		passcode, err := totp.GenerateCodeCustom(key.Secret(), time.Now(), totp.ValidateOpts{
			Period:    uint(key.Period()),
			Digits:    key.Digits(),
			Algorithm: key.Algorithm(),
		})
		if err != nil {
			slog.Error("generate code error, %v", err)
			return err
		}
		detailItem.Password = passcode
		PrintItems([]*model.DetailItem{detailItem})
	default:
		fmt.Printf("unsupported otp type %v\n", key.Type())
	}
	return nil
}

func SearchItem(c *cli.Context) error {
	fmt.Println("--SearchItem--")
	pageId := c.Int("pageId")
	keyword := c.String("keyword")
	slog.Debug("SearchItem, pageId:%v, keyword:%s", pageId, keyword)
	if pageId <= 0 {
		pageId = 1
	}
	result, err := Query(GlobalStore.DB, keyword, pageId, consts.PageSize)
	if err != nil {
		slog.Error("query db error, %v", err)
		return err
	}

	total, err := CountItems(GlobalStore.DB, keyword)
	if err != nil {
		slog.Error("query db error, %v", err)
		return err
	}
	fmt.Println("total:", total)
	fmt.Println("pageSize:", consts.PageSize, "currentPage:", pageId)
	PrintItems(ConvToItems(result))
	return nil
}

func Quit(c *cli.Context) error {
	fmt.Println("Save and Quit")

	if GlobalStore.Dirty {
		GlobalStore.Close()
		sc.CompareAndUpload(GlobalStore.FileName, GlobalStore.FullPath)
	}

	if GlobalStore.NeedBackup {
		timeStr := time.Now().Format(time.RFC3339)
		key := filepath.Join(resource.GlobalOSS.GetDirPath(), GlobalStore.FileName+"."+timeStr)
		resource.GlobalOSS.UploadFile(key, GlobalStore.FullPath)

		// 保证本地和remote一致
		resource.GlobalOSS.DownloadFile(key, GlobalStore.FullPath+"."+timeStr)

		files, err := getAllBackupFiles(resource.DataPath, GlobalStore.FileName)
		if err != nil {
			slog.Error("RestoreItem-GetAllBackupFiles, %v", err)
		}

		sort.Sort(sort.Reverse(sort.StringSlice(files)))
		if len(files) > resource.MaxBackupFileCount {
			for i := resource.MaxBackupFileCount; i < len(files); i++ {
				// 1. remove oss file
				key := filepath.Join(resource.GlobalOSS.GetDirPath(), filepath.Base(files[i]))
				err = resource.GlobalOSS.Delete(key)
				if err != nil {
					slog.Error("remove remote backup file:%v", err)
				}
				slog.Debug("remove remote file:%v", key)
				// 2. remove local file
				slog.Debug("remove local file:%v", files[i])
				err = os.Remove(files[i])
				if err != nil {
					slog.Error("remove local backup file:%v", err)
				}
			}
		}
	}

	resource.LoopExit = true
	return nil
}

func ModifyDBPassword(c *cli.Context) error {
	fmt.Println("Modify DB password")

	answers := struct {
		Password  string `survey:"password"`
		Password2 string `survey:"password2"`
	}{}

	flag := false
	for !flag {
		// perform the questions
		howtouse := `1) The length must be greater than or equal to 8
2) It must contain at least one lowercase character[a-z]
3) It must contain at least one uppercase character[A-Z]
4) It must contain at least one number[0-9]
5) It must contain at least one special character[+-=_&$#^]`
		color.Red(howtouse)
		// 1. password
		prompt := &survey.Password{
			Message: "Please type Database's new password:",
		}
		err := survey.AskOne(prompt, &answers.Password, survey.WithValidator(PasswordComplexityRequired))
		if err != nil {
			fmt.Printf("survey.AskOne error, %v\n", err)
			return err
		}
		// 2. password2
		prompt = &survey.Password{
			Message: "Please type Database's new password again:",
		}
		err = survey.AskOne(prompt, &answers.Password2, survey.WithValidator(survey.Required))
		if err != nil {
			fmt.Printf("survey.AskOne error, %v\n", err)
			return err
		}
		if answers.Password == answers.Password2 {
			flag = true
			fmt.Printf("\n")
		} else {
			fmt.Printf("The password entered 2 times is not the same!")
		}
	}

	oldKey := GlobalStore.Key
	GlobalStore.Key = utils.GenHMacKey([]byte(answers.Password), []byte(GlobalStore.DataBaseIV))
	// 将所有的item重新保存
	itemList, err := Query(GlobalStore.DB, "", 1, 10000)
	if err != nil {
		fmt.Printf("query err, %v\n", err)

		return err
	}

	GlobalStore.Dirty = true
	fmt.Println("Use the command[quit] to make the changes take effect.")

	for _, item := range itemList {
		detailItem := ParseSimpleItem(item, oldKey)

		slog.Debug("id:%v, title:%v", item.ID, item.Title)
		detailItem.ModifiedAt = time.Now().Format(time.RFC3339)
		err = UpdateItem(GlobalStore.DB, ChangeToSimpleItem(detailItem, GlobalStore.Key))
		if err != nil {
			fmt.Printf("UpdateItem error %v\n", err)
			return err
		}
	}
	return nil
}

func PasswordComplexityRequired(val interface{}) error {
	// the reflect value of the result
	value := reflect.ValueOf(val)

	if value.Kind() != reflect.String {
		return errors.New("string is required")
	}
	pwd := value.String()
	if !utils.IsSecurePassword(pwd) {
		return errors.New("password is too simple")
	}
	return nil
}

func Backup(c *cli.Context) error {
	GlobalStore.NeedBackup = true
	fmt.Println("Backup will be executed where it quit.")
	return nil
}

func ChangeToSimpleItem(answers *model.DetailItem, key []byte) *model.SimpleItem {
	bt, _ := json.Marshal(answers)

	itemIV := utils.GenRandIV()
	buffer := bytes.NewBuffer(make([]byte, 0))
	buffer.Write(itemIV)
	buffer.Write([]byte(utils.EncryptAesInCFB(bt, key, itemIV)))
	ic := base64.StdEncoding.EncodeToString(buffer.Bytes())
	item := model.SimpleItem{ID: answers.ID, Title: answers.Title, IVCiphertext: ic}
	return &item
}

func ParseSimpleItem(item *model.SimpleItem, key []byte) *model.DetailItem {
	result := model.DetailItem{}
	bt, _ := base64.StdEncoding.DecodeString(item.IVCiphertext)
	iv := bt[0:aes.BlockSize]
	plaintext := utils.DecryptAesInCFB(bt[aes.BlockSize:], key, iv)
	slog.Debug("ParseSimpleItem:%v", string(plaintext))
	err := json.Unmarshal(plaintext, &result)
	if err != nil {
		slog.Error("json.Unmarshal DetailItem error,%v\n", err)
	}
	result.ID = item.ID
	result.Title = item.Title
	return &result
}

func PrintItems(items []*model.DetailItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Title", "Account",
		"password", "Comment", "ModifiedAt"})

	for _, item := range items {
		table.Append([]string{strconv.Itoa(item.ID), item.Title, item.Account,
			item.Password, item.Comment, item.ModifiedAt,
		})
	}
	table.Render() // Send output
}

func PrintBackups(items []model.BackupItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Tag"})

	for _, item := range items {
		table.Append([]string{strconv.Itoa(item.ID), item.Tag})
	}
	table.Render() // Send output
}

func ConvToItems(items []*model.SimpleItem) []*model.DetailItem {
	result := make([]*model.DetailItem, 0)
	var di *model.DetailItem
	for _, item := range items {
		di = &model.DetailItem{}
		di.ID = item.ID
		di.Title = item.Title
		di.Account = "***"
		di.Password = "***"
		di.Comment = "***"
		di.ModifiedAt = "***"
		result = append(result, di)
	}
	return result
}

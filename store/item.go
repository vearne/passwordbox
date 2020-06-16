package store

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"github.com/vearne/passwordbox/consts"
	"github.com/vearne/passwordbox/sc"

	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/utils"
	slog "github.com/vearne/simplelog"
	"os"
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
		fmt.Println("error", err)
		return err
	}
	answers.ModifiedAt = time.Now().Format(time.RFC3339)
	if err != nil {
		return err
	}
	PrintItems([]*model.DetailItem{&answers})

	InsertItem(GlobalStore.DB, ChangeToSimpleItem(&answers))

	fmt.Println("AddItem-save to file")
	SearchItem(c)
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
	detailItem := ParseSimpleItem(item)
	PrintItems([]*model.DetailItem{detailItem})

	confirmDel := false
	prompt := &survey.Confirm{
		Message: "confirm delete?",
	}
	survey.AskOne(prompt, &confirmDel)
	if confirmDel {
		err = DeleteItem(GlobalStore.DB, itemId)
		if err != nil {
			fmt.Printf("delete item %v error, %v\n", itemId, err)
		} else {
			fmt.Printf("delete item %v success\n", itemId)
		}
	}
	// For user experience
	SearchItem(c)
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
		fmt.Println("can't find %v", itemId)
		return nil
	}
	detailItem := ParseSimpleItem(item)
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

	PrintItems([]*model.DetailItem{detailItem})
	if dirty {
		UpdateItem(GlobalStore.DB, ChangeToSimpleItem(detailItem))
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
	detailItem := ParseSimpleItem(item)
	PrintItems([]*model.DetailItem{detailItem})
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
	GlobalStore.Close()
	sc.CompareAndUpload(GlobalStore.FileName, GlobalStore.FullPath)
	return nil
}

func ChangeToSimpleItem(answers *model.DetailItem) *model.SimpleItem {
	bt, _ := json.Marshal(answers)

	itemIV := utils.GenRandIV()
	buffer := bytes.NewBuffer(make([]byte, 0))
	buffer.Write(itemIV)
	buffer.Write([]byte(utils.EncryptAesInCFB(bt, GlobalStore.Key, itemIV)))
	ic := base64.StdEncoding.EncodeToString(buffer.Bytes())
	item := model.SimpleItem{ID: answers.ID, Title: answers.Title, IVCiphertext: ic}
	return &item
}

func ParseSimpleItem(item *model.SimpleItem) *model.DetailItem {
	result := model.DetailItem{}
	bt, _ := base64.StdEncoding.DecodeString(item.IVCiphertext)
	iv := bt[0:aes.BlockSize]
	plaintext := utils.DecryptAesInCFB(bt[aes.BlockSize:], GlobalStore.Key, iv)
	slog.Debug("ParseSimpleItem:%v", string(plaintext))
	json.Unmarshal(plaintext, &result)
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

package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/peterh/liner"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"github.com/vearne/passwordbox/args"
	"github.com/vearne/passwordbox/consts"
	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/resource"
	"github.com/vearne/passwordbox/sc"
	"github.com/vearne/passwordbox/store"
	"github.com/vearne/passwordbox/utils"
	slog "github.com/vearne/simplelog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "passwordbox"
	app.Version = consts.Version
	MasterAuthor := &cli.Author{Name: "vearne", Email: "asdwoshiaotian@gmail.com"}
	app.Authors = []*cli.Author{MasterAuthor}
	app.Copyright = "(c)2020-? vearne"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "data",
			Aliases: []string{"c"},
			Value:   ".",
			Usage:   "Load data from `DIR`",
		},
		&cli.StringFlag{
			Name:    "loglevel",
			Aliases: []string{"l"},
			Usage:   "specify log level, optional: debug|info|warn|error",
			Value:   "info",
		},
		&cli.IntFlag{
			Name:  "maxBackupFileCount",
			Usage: "Maximum number of backup file retained",
			Value: 5,
		},
		&cli.StringFlag{
			Name: "oss",
			Usage: `--oss /etc/qingstor.yaml
					specify Object Storage Service address, 
					Note: pwbox identify cloud services by configuration file name.
					optional: qingstor.yaml`,
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "clear",
			Usage: "clear",
			Action: func(c *cli.Context) error {
				fmt.Print("\x1b[H\x1b[2J")
				return nil
			},
		},
		{
			Name:   "add",
			Usage:  "add",
			Action: store.AddItem,
		},
		{
			Name:   "delete",
			Usage:  "delete -itemId <itemId>",
			Action: store.DelItem,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:     "itemId",
					Required: true,
				},
			},
		},
		{
			Name:   "modify",
			Usage:  "modify -itemId <itemId>",
			Action: store.ModifyItem,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:     "itemId",
					Required: true,
				},
			},
		},
		{
			Name:   "view",
			Usage:  "view -itemId <itemId>",
			Action: store.ViewItem,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:     "itemId",
					Required: true,
				},
			},
		},
		{
			Name:      "search",
			Usage:     "search [-pageId <pageId>] [-keyword <keyword>]",
			UsageText: "pageId/keyword is optional.",
			Action:    store.SearchItem,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "pageId",
					Value: 1,
				},
				&cli.StringFlag{
					Name:  "keyword",
					Value: "",
				},
			},
		},
		{
			Name:   "backup",
			Action: store.Backup,
		},
		{
			Name:      "restore",
			Usage:     "restore [-tagId <tagId>]",
			UsageText: "Restore from backup data with specific tag.",
			Action:    store.RestoreItem,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "tagId",
					Value: -1,
				},
			},
		},
		{
			Name:   "quit",
			Action: store.Quit,
		},
		{
			Name: "help",
			Action: func(cxt *cli.Context) error {
				return cli.ShowAppHelp(cxt)
			},
		},
	}
	app.Action = MainLogic
	err := app.Run(os.Args)
	if err != nil {
		slog.Fatal("app run error, %v", err)
	}
}

func MainLogic(c *cli.Context) error {
	logLevel := c.String("loglevel")
	slog.Level = slog.LogMap[logLevel]

	maxBackupFileCount := c.Int("maxBackupFileCount")
	resource.MaxBackupFileCount = maxBackupFileCount
	if resource.MaxBackupFileCount <= 0 {
		resource.MaxBackupFileCount = 5
	}

	// check data directory exist?
	dataPath := c.String("data")
	if !utils.Exists(dataPath) {
		return cli.Exit("Data directory is not exist.", -1)
	}

	if !utils.IsDir(dataPath) {
		return cli.Exit("Data directory is not directory.", -1)
	}

	// datapath
	resource.DataPath = dataPath

	ossConfigFile := c.String("oss")
	if len(ossConfigFile) > 0 {
		viper.SetConfigFile(ossConfigFile)
		if err := viper.ReadInConfig(); err == nil {
			slog.Info("Using config file: %v", viper.ConfigFileUsed())
		} else {
			slog.Fatal("can't find config file, %v", err)
		}

		ossType := extractType(ossConfigFile)
		switch ossType {
		case "qingstor":
			oss := sc.QingStor{}
			err := viper.Unmarshal(&oss)
			if err != nil {
				slog.Fatal("can't parse oss config file, %v", err)
			}
			resource.GlobalOSS = &oss
		case "oss":
			oss := sc.AliOSS{}
			err := viper.Unmarshal(&oss)
			if err != nil {
				slog.Fatal("can't parse oss config file, %v", err)
			}
			resource.GlobalOSS = &oss
		default:
			slog.Fatal("Unsupport Cloud service providers, %v", ossType)
		}

		// init object storage service
		err := resource.GlobalOSS.Init()
		if err != nil {
			slog.Fatal("GlobalOSS init error:%v", err)
		}
		// sync from oss
		sc.CompareAndDownloadAll()
	}

LOGIN:
	fmt.Println("---- login database ----")
	database := ""
	promptDatabse := &survey.Input{
		Message: "Please type database's name:",
	}
	err := survey.AskOne(promptDatabse, &database, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Printf("survey.AskOne error, %v\n", err)
	}

	database = strings.TrimSpace(database)
	slog.Debug("database:%v", database)
	filename := utils.Sha256N(database, consts.HashCount)
	fullpath := filepath.Join(dataPath, filename)
	fmt.Println("fullpath", fullpath)
	slog.Debug("fullpath:%v", fullpath)
	if !utils.Exists(fullpath) {
		createFlag := false
		prompt := &survey.Confirm{
			Message: "Database is not exist.\nDo you like to create database now?",
		}
		err = survey.AskOne(prompt, &createFlag)
		if err != nil {
			fmt.Printf("survey.AskOne error, %v\n", err)
			createFlag = false
		}

		if !createFlag {
			return nil
		}

		// ---- create database ----
		err = createDatabase(dataPath)
		if err != nil {
			fmt.Printf("createDatabase error, %v\n", err)
			return err
		}
		goto LOGIN
	}

	password := ""
	promptPasswd := &survey.Password{
		Message: "Please type your password:",
	}
	err = survey.AskOne(promptPasswd, &password, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Printf("survey.AskOne error, %v\n", err)
	}

	db, err := store.OpenDatabaseStore(dataPath, &model.Database{Name: database, Password: password})
	if err != nil {
		slog.Fatal("openDatabase error, %v", err)
		os.Exit(1)
	}
	store.GlobalStore = db

	// Even if the database name or password is wrong, sqlite3 is still successfully opened,
	// and the error will not be reported until you actually query.
	db.Hint, err = store.GetHint(db.DB)
	if err != nil {
		slog.Debug("Get Hint error, %v", err)
		fmt.Printf("Decrypt error, Maybe DatabaseName or Password is invalid.\n")
		os.Exit(2)
	}

	info := color.New(color.FgRed, color.BgGreen).SprintFunc()
	fmt.Printf("Hint for database %v is %v", info(db.DatabaseName), info(db.Hint))

	line := liner.NewLiner()
	defer line.Close()

	msg := `
Tip: Up and down arrow keys can switch historical commands.
Tip: Ctrl + A jumps to the beginning of the command.
Tip: Ctrl + E jumps to the end of the command.
Tip: Type help for help.
		`
	fmt.Println(msg)
	// For user experience
	err = store.SearchItem(c)
	if err != nil {
		fmt.Printf("SearchItem error, %v\n", err)
	}

	for {
		commandLine, err := line.Prompt(store.GlobalStore.DatabaseName + " > ")
		if err != nil {
			slog.Error("commandLine:%v, error:%v", commandLine, err)
		}
		slog.Debug("commandLine:%v", commandLine)
		line.AppendHistory(commandLine)

		cmdArgs := args.Parse(commandLine)
		if len(cmdArgs) <= 0 {
			continue
		}
		s := []string{os.Args[0]}
		s = append(s, cmdArgs...)

		cmd := cmdArgs[0]
		if !utils.FindInSlice(cmd, []string{
			"clear", "add", "delete", "quit",
			"modify", "view", "search",
			"backup", "restore", "help"}) {
			fmt.Println("unknow command", cmd)
			continue
		}

		err = c.App.Run(s)
		if err != nil {
			fmt.Println("App.Run error", s)
		}

		if resource.LoopExit {
			break
		}
	}
	return nil
}

func createDatabase(dataPath string) error {
	fmt.Println("---- create database ----")
	// the questions to ask
	var qs = []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{Message: "Please type database's name:"},
			Validate: survey.Required,
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Please type password:",
			},
			Validate: survey.Required,
		},
		{
			Name:     "hint",
			Prompt:   &survey.Input{Message: "Please type hint:"},
			Validate: survey.Required,
		},
	}
	answers := model.Database{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println("error", err)
		return err
	}

	st := store.NewDatabaseStore(dataPath, &answers)
	err = st.Init()
	if err != nil {
		slog.Error("init store error,%v", err)
		return err
	}
	err = st.Close()
	if err != nil {
		slog.Error("close store error,%v", err)
		return err
	}

	return nil
}

// /abc/def/qingstor.yaml
func extractType(localfilepath string) string {
	_, filename := filepath.Split(localfilepath)
	itemList := strings.Split(filename, ".")
	return itemList[0]
}

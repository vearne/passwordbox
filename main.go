package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
	"github.com/vearne/passwordbox/args"
	"github.com/vearne/passwordbox/consts"
	slog "github.com/vearne/passwordbox/log"
	"github.com/vearne/passwordbox/model"
	"github.com/vearne/passwordbox/store"
	"github.com/vearne/passwordbox/utils"
	"os"
	"path/filepath"
	"strings"
)

var (
	Version = "v0.0.1"
)

func main() {
	app := cli.NewApp()
	app.Name = "passwordbox"
	app.Version = Version
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
			Value:   "debug",
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
			Usage:     "search [-pageId <pageId>] [-keyword <keyword>s]",
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
			Name:   "quit",
			Action: store.Quit,
		},
		{
			Name: "help",
			Action: func(cxt *cli.Context) error {
				cli.ShowAppHelp(cxt)
				return nil
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
	// check data directory exist?
	dataPath := c.String("data")
	logLevel := c.String("loglevel")
	slog.Level = slog.LogMap[logLevel]

	if !utils.Exists(dataPath) {
		return cli.Exit("Data directory is not exist.", -1)
	}

	if !utils.IsDir(dataPath) {
		return cli.Exit("Data directory is not directory.", -1)
	}

LOGIN:
	fmt.Println("---- login database ----")
	database := ""
	promptDatabse := &survey.Input{
		Message: "Please type database's name:",
	}
	survey.AskOne(promptDatabse, &database, survey.WithValidator(survey.Required))

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
		survey.AskOne(prompt, &createFlag)
		if !createFlag {
			return nil
		}

		// ---- create database ----
		createDatabase(dataPath)
		goto LOGIN
	}

	password := ""
	promptPasswd := &survey.Password{
		Message: "Please type your password:",
	}
	survey.AskOne(promptPasswd, &password, survey.WithValidator(survey.Required))

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
	for {
		msg := `
Tip: Up and down arrow keys can switch historical commands.
Tip: Ctrl + A jumps to the beginning of the command.
Tip: Ctrl + E jumps to the end of the command.
Tip: Type help for help.
		`
		fmt.Println(msg)
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
		if cmdArgs[0] == "quit" {
			store.Quit(c)
			break
		}
		if !utils.FindInSlice(cmd, []string{"clear", "add", "delete",
			"modify", "view", "search", "help"}) {
			fmt.Println("unknow command", cmd)
			continue
		}

		c.App.Run(s)

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
			Name:   "hint",
			Prompt: &survey.Input{Message: "Please type hint[optional]:"},
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

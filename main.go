package main

import (
	"flag"
	"fmt"
	surveyCore "github.com/AlecAivazis/survey/v2/core"
	"github.com/PierreKieffer/http-tanker/pkg/cli"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"github.com/mgutz/ansi"
	"os"
	"os/user"
	"strings"
)

func init() {
	//Override survey colors
	surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
		switch style {
		case "white":
			if Is256ColorSupported() {
				return fmt.Sprintf("\x1b[%d;5;%dm", 38, 242)
			} else {
				return ansi.ColorCode("default")
			}
		default:
			return ansi.ColorCode(style)
		}
	}
}

func main() {

	localUser, err := user.Current()

	databaseDir := flag.String("db", fmt.Sprintf("%v/http-tanker", localUser.HomeDir), "tanker database directory")

	database := &core.Database{
		DatabaseDir:  *databaseDir,
		DatabaseFile: fmt.Sprintf("%s/http-tanker-data.json", *databaseDir),
	}

	err = database.InitDB()
	if err != nil {
		fmt.Println(err)
	}

	app := &cli.App{
		SigChan:  make(chan cli.Signal),
		Database: database,
	}

	app.Run()

}

func Is256ColorSupported() bool {
	if strings.Contains(os.Getenv("TERM"), "256") || strings.Contains(os.Getenv("COLORTERM"), "256") {
		return true
	}
	return false
}

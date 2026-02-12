package main

import (
	"flag"
	"fmt"
	surveyCore "github.com/AlecAivazis/survey/v2/core"
	"github.com/PierreKieffer/http-tanker/pkg/cli"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	tankerMcp "github.com/PierreKieffer/http-tanker/pkg/mcp"
	"github.com/mgutz/ansi"
	"os"
	"os/user"
)

func init() {
	//Override survey colors
	surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
		switch style {
		case "white":
			if color.Is256ColorSupported() {
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current user: %v\n", err)
		os.Exit(1)
	}

	databaseDir := flag.String("db", fmt.Sprintf("%v/.http-tanker", localUser.HomeDir), "tanker database directory")
	mcpMode := flag.Bool("mcp", false, "start as MCP server (stdio transport)")
	flag.Parse()

	database := &core.Database{
		DatabaseDir:  *databaseDir,
		DatabaseFile: fmt.Sprintf("%s/http-tanker-data.json", *databaseDir),
	}

	err = database.InitDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	if *mcpMode {
		if err := tankerMcp.Serve(database); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	app := &cli.App{
		SigChan:  make(chan cli.Signal),
		Database: database,
	}

	app.Run()

}

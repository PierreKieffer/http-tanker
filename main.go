package main

import (
	"flag"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/cli"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"os/user"
)

func init() {
}

func main() {
	localUser, err := user.Current()

	databaseDir := flag.String("db", fmt.Sprintf("%v/tanker", localUser.HomeDir), "tanker database directory")

	database := &core.Database{
		DatabaseDir:  *databaseDir,
		DatabaseFile: fmt.Sprintf("%s/tanker-data.json", *databaseDir),
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

package cli

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PierreKieffer/http-tanker/pkg/core"
)

/*
Home
Display Home menu options
*/
func (app *App) Home() error {

	core.DrawBox("Home Menu", nil)
	var menu = []*survey.Question{
		{
			Name: "home",
			Prompt: &survey.Select{
				Message: "Select :",
				Options: []string{SigBrowse, SigCreate, SigAbout, SigExit},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Home string
	}{}

	err := survey.Ask(menu, &answers)

	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	sig := Signal{
		Meta: "home",
		Sig:  answers.Home,
	}

	app.SigChan <- sig

	return nil
}

/*
Requests
Display all available requests previously created by the user
*/
func (app *App) Requests() error {

	reqList := make([]string, 0, len(app.Database.Data)+1)
	reqList = append(reqList, SigBackHome)
	displayToName := make(map[string]string, len(app.Database.Data))
	for name, r := range app.Database.Data {
		label := "[" + r.Method + "] " + name + " - " + r.URL
		reqList = append(reqList, label)
		displayToName[label] = name
	}

	core.DrawBox("Requests", nil)
	var menu = []*survey.Question{
		{
			Name: "requests",
			Prompt: &survey.Select{
				Message: "Select :",
				Options: reqList,
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Requests string
	}{}

	err := survey.Ask(menu, &answers)

	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	selected := answers.Requests
	if name, ok := displayToName[selected]; ok {
		selected = name
	}

	sig := Signal{
		Sig:     SigReqSelect,
		Meta:    selected,
		Display: true,
	}

	app.SigChan <- sig

	return nil
}

/*
Request
Display options after selecting a request
*/
func (app *App) Request(reqName string, display bool) error {

	var menu = []*survey.Question{
		{
			Name: "request",
			Prompt: &survey.Select{
				Options: []string{SigRun, SigCurl, SigEdit, SigDelete, SigBackRequests, SigExit},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Request string
	}{}

	if display {
		app.Database.Display(reqName)
	}

	err := survey.Ask(menu, &answers)

	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	sig := Signal{
		Meta: reqName,
		Sig:  answers.Request,
	}

	app.SigChan <- sig

	return nil

}

/*
About
*/
func (app *App) About() error {

	Banner()
	fmt.Println(string(aboutBytes))
	fmt.Println("")

	var menu = []*survey.Question{
		{
			Name: "home",
			Prompt: &survey.Select{
				Options: []string{SigBackHome},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Home string
	}{}

	err := survey.Ask(menu, &answers)
	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	sig := Signal{
		Sig: SigHome,
	}

	app.SigChan <- sig
	return nil
}

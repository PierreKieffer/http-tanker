package cli

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"os"
)

type App struct {
	SigChan  chan Signal
	Database *core.Database
}

type Signal struct {
	Meta string
	Sig  string
}

/*
Run
Main run application method
Consumes the channel, and acts accordingly
*/
func (app *App) Run() {

	go app.Home()

	for {
		sig := <-app.SigChan
		switch sig.Sig {
		case "Requests":
			go app.Requests()
		case "Home":
			go app.Home()
		case "Exit":
			os.Exit(1)
		case "Create request":
			go app.Create()

		default:
			switch sig.Meta {
			case "reqSelect":
				go app.Request(sig.Sig)
			case "reqCreate":
				go app.Request(sig.Sig)
			}

		}
	}
}

/*
Home
Display Home menu options
*/
func (app *App) Home() error {

	var menu = []*survey.Question{
		{
			Name: "home",
			Prompt: &survey.Select{
				Message: "---- Home Menu ----",
				Options: []string{"Requests", "Create request", "Options", "Home", "Exit"},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Home string
	}{}

	err := survey.Ask(menu, &answers)

	if err != nil {
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

	var reqList = []string{"Home"}
	for r, _ := range app.Database.Data {
		reqList = append(reqList, r)
	}

	var menu = []*survey.Question{
		{
			Name: "requests",
			Prompt: &survey.Select{
				Message: "---- Requests ----",
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
		return err
	}

	sig := Signal{
		Meta: "reqSelect",
		Sig:  answers.Requests,
	}

	app.SigChan <- sig

	return nil
}

/*
Request
Display options after selecting a request
*/
func (app *App) Request(reqName string) error {

	var menu = []*survey.Question{
		{
			Name: "request",
			Prompt: &survey.Select{
				Options: []string{"Home", "Run", "Details", "Edit", "Delete"},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Request string
	}{}

	app.Database.Display(reqName)

	err := survey.Ask(menu, &answers)

	if err != nil {
		return err
	}

	sig := Signal{
		Meta: "reqAction",
		Sig:  answers.Request,
	}

	app.SigChan <- sig

	return nil

}

/*
Create
Request creation workflow
*/
func (app *App) Create() error {

	// Enter request generic values
	var menu = []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Name : ",
			},
			Validate: survey.Required,
		},
		{
			Name: "method",
			Prompt: &survey.Select{
				Message: "Method : ",
				Options: []string{"GET", "POST"},
			},
			Validate: survey.Required,
		},
		{
			Name: "url",
			Prompt: &survey.Input{
				Message: "URL : ",
			},
			Validate: survey.Required,
		},
	}

	genericAnswer := struct {
		Name   string
		Method string
		Url    string
	}{}

	err := survey.Ask(menu, &genericAnswer)

	if err != nil {
		return err
	}

	// Enter request body and headers values
	var body = []*survey.Question{}

	switch genericAnswer.Method {
	case "GET":
		body = []*survey.Question{
			{
				Name: "params",
				Prompt: &survey.Input{
					Message: fmt.Sprintf("%s \n", `Params (Enter the parameters in the format {"key": "value"}, default = {}) : `),
					Default: "{}",
				},
				Validate: func(val interface{}) error {
					var jsonData map[string]interface{}
					err := json.Unmarshal([]byte(val.(string)), &jsonData)
					if err != nil {
						return fmt.Errorf("Wrong input format")
					}
					return nil
				},
			},
		}

	case "POST":
		body = []*survey.Question{
			{
				Name: "payload",
				Prompt: &survey.Input{
					Message: fmt.Sprintf("%s \n", `Payload (Enter the payload in the format {"key": "value"}, default = {}) : `),
					Default: "{}",
				},
				Validate: func(val interface{}) error {
					var jsonData map[string]interface{}
					err := json.Unmarshal([]byte(val.(string)), &jsonData)
					if err != nil {
						return fmt.Errorf("Wrong input format")
					}
					return nil
				},
			},
		}

	}

	var headers = []*survey.Question{
		{
			Name: "headers",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("%s \n", `Headers (Enter the headers in the format {"key": "value"}, default = {}) : `),
				Default: "{}",
			},
			Validate: func(val interface{}) error {
				var jsonData map[string]interface{}
				err := json.Unmarshal([]byte(val.(string)), &jsonData)
				if err != nil {
					return fmt.Errorf("Wrong input format")
				}
				return nil
			},
		},
	}

	body = append(body, headers...)

	var bodyAnswers = struct {
		Params  string
		Payload string
		Headers string
	}{}

	err = survey.Ask(body, &bodyAnswers)

	if err != nil {
		return err
	}

	sig := Signal{
		Meta: "reqCreate",
		Sig:  genericAnswer.Name,
	}

	// Build Request object
	var R = core.Request{
		Name:   genericAnswer.Name,
		Method: genericAnswer.Method,
		URL:    genericAnswer.Url,
	}

	switch R.Method {
	case "GET":
		if bodyAnswers.Params != "" {
			var jsonData map[string]interface{}
			json.Unmarshal([]byte(bodyAnswers.Params), &jsonData)
			R.Params = jsonData

		} else {
			R.Params = map[string]interface{}{}
		}
	case "POST":
		if bodyAnswers.Payload != "" {
			var jsonData map[string]interface{}
			json.Unmarshal([]byte(bodyAnswers.Payload), &jsonData)
			R.Payload = jsonData

		} else {
			R.Payload = map[string]interface{}{}
		}
	}

	if bodyAnswers.Headers != "" {
		var jsonData map[string]interface{}
		json.Unmarshal([]byte(bodyAnswers.Headers), &jsonData)
		R.Headers = jsonData

	} else {
		R.Headers = map[string]interface{}{}
	}

	// Save request in local database
	app.Database.Data[R.Name] = R
	app.Database.Save()

	app.SigChan <- sig

	return nil
}

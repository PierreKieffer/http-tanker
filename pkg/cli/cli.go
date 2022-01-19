package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"os"
	"reflect"
)

var (
	version string = "edge"

	//go:embed assets/*
	assets embed.FS
)

type App struct {
	SigChan  chan Signal
	Database *core.Database
}

type Signal struct {
	Meta    string
	Sig     string
	Display bool
}

/*
Run
Main run application method
Consumes the channel, and acts accordingly
*/
func (app *App) Run() {

	Banner()
	go app.Home()

	for {
		sig := <-app.SigChan
		switch sig.Sig {
		case "Home", "Back to Home Menu":
			Banner()
			go app.Home()
		case "Browse requests", "Back to requests":
			Banner()
			go app.Requests()
		case "Exit":
			fmt.Print("\033[H\033[2J")
			os.Exit(1)
		case "Create request":
			Banner()
			go app.Create()
		case "Run":
			Banner()
			go app.RunRequest(sig.Meta)
		case "reqSelect":
			switch sig.Meta {
			case "Home", "Back to Home Menu":
				Banner()
				go app.Home()
			case "Browse requests", "Back to requests":
				Banner()
				go app.Requests()
			case "Exit":
				fmt.Print("\033[H\033[2J")
				os.Exit(1)
			default:
				Banner()
				go app.Request(sig.Meta, sig.Display)
			}

		case "reqCreate":
			Banner()
			go app.Request(sig.Meta, sig.Display)
		case "Edit":
			Banner()
			go app.Edit(sig.Meta)
		case "Delete":
			Banner()
			go app.Delete(sig.Meta)
		case "About":
			go app.About()

		}
	}
}

/*
Home
Display Home menu options
*/
func (app *App) Home() error {

	home := ` -------------
   | Home Menu |
   -------------
	`
	var menu = []*survey.Question{
		{
			Name: "home",
			Prompt: &survey.Select{
				Message: home,
				Options: []string{"Browse requests", "Create request", "About", "Exit"},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Home string
	}{}

	err := survey.Ask(menu, &answers)

	if err != nil {
		fmt.Printf("ERROR : %v", err)
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

	var reqList = []string{"Back to Home Menu"}
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
		Sig:     "reqSelect",
		Meta:    answers.Requests,
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
				Options: []string{"Run", "Edit", "Delete", "Back to requests", "Exit"},
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
RunRequest
Execute HTTP request, display response
*/
func (app *App) RunRequest(reqName string) error {
	r := app.Database.Data[reqName]
	resp, err := r.CallHTTP()
	if err != nil {
		app.ErrorHandler(err)
		// fmt.Printf("ERROR : %v \n", err)
		// sig := Signal{
		// Meta: reqName,
		// Sig:  "reqSelect",
		// }
		// app.SigChan <- sig
		return err
	}

	// Ask if user wants to inspect response
	var menu = []*survey.Question{
		{
			Name: "inspectResponse",
			Prompt: &survey.Confirm{
				Message: "Inspect response in editor ?",
				Default: false,
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		InspectResponse bool
	}{}

	err = survey.Ask(menu, &answers)
	if err != nil {
		fmt.Printf("ERROR : %v", err)
		return err
	}

	switch answers.InspectResponse {
	case true:
		/*
		   Inspect Response
		   Open response in editor to inspect
		*/
		var content string
		var menu = &survey.Editor{
			FileName:      "http-tanker-response-inspector*.json",
			Default:       resp,
			AppendDefault: true,
			HideDefault:   true,
		}

		err := survey.AskOne(menu, &content)
		if err != nil {
			return err
		}
	}

	sig := Signal{
		Meta:    reqName,
		Sig:     "reqSelect",
		Display: true,
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
					Message: fmt.Sprintf("%s \n", `Params (Enter the string parameters in the format {"key": "value"}, default = {}) : `),
					Default: "{}",
				},
				Validate: func(val interface{}) error {
					var jsonData map[string]interface{}
					err := json.Unmarshal([]byte(val.(string)), &jsonData)
					if err != nil {
						return fmt.Errorf("Wrong input format")
					}
					for k, v := range jsonData {
						if reflect.TypeOf(v).String() != "string" {
							return fmt.Errorf("Wront value type for param %v : %v. Type must be a string", k, reflect.TypeOf(v).String())
						}
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
		Sig:     "reqCreate",
		Meta:    genericAnswer.Name,
		Display: true,
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

/*
Edit
*/
func (app *App) Edit(reqName string) error {

	app.Database.Display(reqName)

	req := app.Database.Data[reqName]
	jsonReq, _ := json.MarshalIndent(req, "", "    ")

	content := ""

	var menu = &survey.Editor{
		// Message:       "Edit : ",
		FileName:      "http-tanker-edit*.json",
		Default:       string(jsonReq),
		AppendDefault: true,
		HideDefault:   true,
	}

	err := survey.AskOne(menu, &content)
	if err != nil {
		return err
	}
	var updateReq core.Request
	json.Unmarshal([]byte(content), &updateReq)

	app.Database.Data[updateReq.Name] = updateReq
	app.Database.Save()
	app.Database.Load()

	sig := Signal{
		Sig:     "reqCreate",
		Meta:    updateReq.Name,
		Display: true,
	}

	message := fmt.Sprintf("The request %v has been edited successfully", reqName)
	fmt.Println("")
	fmt.Println(string(color.ColorGreen), message, string(color.ColorReset))
	app.SigChan <- sig

	return nil

}

/*
Delete
*/
func (app *App) Delete(reqName string) error {

	var menu = []*survey.Question{
		{
			Name: "confirmDelete",
			Prompt: &survey.Confirm{
				Message: fmt.Sprintf("This will delete the request : %v. Continue ?", reqName),
				Default: false,
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		ConfirmDelete bool
	}{}

	err := survey.Ask(menu, &answers)
	if err != nil {
		fmt.Printf("ERROR : %v", err)
		return err
	}

	switch answers.ConfirmDelete {
	case true:
		app.Database.Delete(reqName)
		message := fmt.Sprintf("The request %v was successfully deleted", reqName)
		fmt.Println("")
		fmt.Println(string(color.ColorGreen), message, string(color.ColorReset))
		fmt.Println("")
	}

	menu = []*survey.Question{
		{
			Name: "back",
			Prompt: &survey.Select{
				// showingHelp: false,
				Options: []string{"Back to Home Menu", "Back to requests", "Exit"},
			},
			Validate: survey.Required,
		},
	}

	back := struct {
		Back string
	}{}

	err = survey.Ask(menu, &back)
	if err != nil {
		fmt.Printf("ERROR : %v", err)
		return err
	}

	sig := Signal{
		Sig: back.Back,
	}

	app.SigChan <- sig
	return nil
}

/*
About
*/
func (app *App) About() error {

	Banner()
	aboutBuffer, _ := assets.ReadFile("assets/about")
	fmt.Println(string(aboutBuffer))
	fmt.Println("")

	var menu = []*survey.Question{
		{
			Name: "home",
			Prompt: &survey.Select{
				// showingHelp: false,
				Options: []string{"Back to Home Menu"},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Home string
	}{}

	err := survey.Ask(menu, &answers)
	if err != nil {
		fmt.Printf("ERROR : %v", err)
		return err
	}

	sig := Signal{
		Sig: "Home",
	}

	app.SigChan <- sig
	return nil
}

/*
Banner
*/
func Banner() {
	fmt.Print("\033[H\033[2J")
	bannerBuffer, _ := assets.ReadFile("assets/banner")
	fmt.Println(string(bannerBuffer))
	fmt.Println(string(color.ColorGrey), fmt.Sprintf("  version : %v", version), string(color.ColorReset))
	fmt.Print("\n")
}

/*
Error handler
*/

func (app *App) ErrorHandler(err error) {
	fmtError := fmt.Sprintf("ERROR : %v", err.Error())
	fmt.Println(string(color.ColorRed), fmtError, string(color.ColorReset))
}

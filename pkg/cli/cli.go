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
	"strings"
	"time"
)

const (
	SigHome        = "Home"
	SigBackHome    = "Back to Home Menu"
	SigBrowse      = "Browse requests"
	SigBackRequests = "Back to requests"
	SigExit        = "Exit"
	SigCreate      = "Create request"
	SigRun         = "Run"
	SigReqSelect   = "reqSelect"
	SigReqCreate   = "reqCreate"
	SigEdit        = "Edit"
	SigDelete      = "Delete"
	SigCurl        = "cURL"
	SigAbout       = "About"
)

var (
	version string = "edge"

	//go:embed assets/*
	assets embed.FS

	bannerBytes, _ = assets.ReadFile("assets/banner")
	aboutBytes, _  = assets.ReadFile("assets/about")
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

func (app *App) handleNavigation(sig string) bool {
	switch sig {
	case SigHome, SigBackHome:
		Banner()
		go app.Home()
		return true
	case SigBrowse, SigBackRequests:
		Banner()
		go app.Requests()
		return true
	case SigExit:
		fmt.Print("\033[H\033[2J")
		os.Exit(0)
		return true
	}
	return false
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
		if app.handleNavigation(sig.Sig) {
			continue
		}
		switch sig.Sig {
		case SigCreate:
			Banner()
			go app.Create()
		case SigRun:
			Banner()
			go app.RunRequest(sig.Meta)
		case SigReqSelect:
			if !app.handleNavigation(sig.Meta) {
				Banner()
				go app.Request(sig.Meta, sig.Display)
			}
		case SigReqCreate:
			Banner()
			go app.Request(sig.Meta, sig.Display)
		case SigCurl:
			Banner()
			go app.ShowCurl(sig.Meta)
		case SigEdit:
			Banner()
			go app.Edit(sig.Meta)
		case SigDelete:
			Banner()
			go app.Delete(sig.Meta)
		case SigAbout:
			go app.About()
		}
	}
}

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
RunRequest
Execute HTTP request, display response
*/
func (app *App) RunRequest(reqName string) error {
	r := app.Database.Data[reqName]

	// Spinner pendant l'execution
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	stopSpinner := make(chan struct{})
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-stopSpinner:
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				fmt.Printf("\r%s %sExecuting request...%s", spinChars[i%len(spinChars)], color.ColorCyan, color.ColorReset)
				i++
			}
		}
	}()

	response, err := r.CallHTTP()
	close(stopSpinner)
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		fmtError := "ERROR : " + err.Error()
		fmt.Println(color.ColorRed, fmtError, color.ColorReset)

		var menu = []*survey.Question{
			{
				Name: "back",
				Prompt: &survey.Select{
					Options: []string{"Back to " + reqName + " request"},
				},
				Validate: survey.Required,
			},
		}

		answers := struct {
			Back string
		}{}

		err := survey.Ask(menu, &answers)
		if err != nil {
			app.ErrorHandler(err)
			return err
		}

		sig := Signal{
			Meta:    reqName,
			Sig:     SigReqSelect,
			Display: true,
		}

		app.SigChan <- sig
		return err
	}

	core.DisplayResponse(response)

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
		app.ErrorHandler(err)
		return err
	}

	if answers.InspectResponse {
		jsonResp, err := json.MarshalIndent(response, "", "    ")
		if err != nil {
			app.ErrorHandler(err)
			return err
		}
		var content string
		var menu = &survey.Editor{
			FileName:      "http-tanker-response-inspector*.json",
			Default:       string(jsonResp),
			AppendDefault: true,
			HideDefault:   true,
		}

		err = survey.AskOne(menu, &content)
		if err != nil {
			app.ErrorHandler(err)
			return err
		}
	}

	sig := Signal{
		Meta:    reqName,
		Sig:     SigReqSelect,
		Display: true,
	}
	app.SigChan <- sig
	return nil
}

/*
ShowCurl
Display formatted curl command for a request
*/
func (app *App) ShowCurl(reqName string) error {
	r := app.Database.Data[reqName]
	curlCmd := r.CurlCommand()

	core.DrawBox("cURL command", []string{curlCmd})

	var menu = []*survey.Question{
		{
			Name: "back",
			Prompt: &survey.Select{
				Options: []string{"Back to " + reqName + " request", SigBackRequests, SigBackHome},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Back string
	}{}

	err := survey.Ask(menu, &answers)
	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	switch answers.Back {
	case SigBackRequests, SigBackHome:
		app.SigChan <- Signal{Sig: answers.Back}
	default:
		app.SigChan <- Signal{Meta: reqName, Sig: SigReqSelect, Display: true}
	}
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
				Options: []string{"GET", "POST", "DELETE", "PUT"},
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
		app.ErrorHandler(err)
		return err
	}

	// Enter request body and headers values
	var body = []*survey.Question{}

	switch genericAnswer.Method {
	case "GET", "DELETE":
		body = []*survey.Question{
			{
				Name: "params",
				Prompt: &survey.Input{
					Message: fmt.Sprintf("%s \n", `Params (Enter the string parameters in {"key": "value"} format , default = {}) : `),
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
							return fmt.Errorf("Wrong value type for param %v : %v. Type must be a string", k, reflect.TypeOf(v).String())
						}
					}
					return nil
				},
			},
		}

	case "POST", "PUT":
		dfltPayload := map[string]interface{}{"foo": "bar"}
		jsonDfltPayload, _ := json.MarshalIndent(dfltPayload, "", "    ")
		body = []*survey.Question{
			{
				Name: "payload",
				Prompt: &survey.Editor{
					Message:       fmt.Sprintf("%s \n", `Payload (Enter the payload in json format {"key": "value"}) : `),
					FileName:      "http-tanker-post-payload*.json",
					Default:       string(jsonDfltPayload),
					HideDefault:   true,
					AppendDefault: true,
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
				Message: fmt.Sprintf("%s \n", `Headers (Enter the headers in json format {"key": "value"}, default = {}) : `),
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
		app.ErrorHandler(err)
		return err
	}

	// Ask for TLS verification skip
	insecureAnswer := struct {
		Insecure bool
	}{}
	insecureMenu := []*survey.Question{
		{
			Name: "insecure",
			Prompt: &survey.Confirm{
				Message: "Skip TLS certificate verification ?",
				Default: false,
			},
		},
	}
	err = survey.Ask(insecureMenu, &insecureAnswer)
	if err != nil {
		app.ErrorHandler(err)
		return err
	}

	sig := Signal{
		Sig:     SigReqCreate,
		Meta:    genericAnswer.Name,
		Display: true,
	}

	// Build Request object
	var R = core.Request{
		Name:     genericAnswer.Name,
		Method:   genericAnswer.Method,
		URL:      genericAnswer.Url,
		Insecure: insecureAnswer.Insecure,
	}

	switch R.Method {
	case "GET", "DELETE":
		if bodyAnswers.Params != "" {
			var jsonData map[string]interface{}
			json.Unmarshal([]byte(bodyAnswers.Params), &jsonData)
			R.Params = jsonData

		} else {
			R.Params = map[string]interface{}{}
		}
	case "POST", "PUT":
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
	if err := app.Database.Save(); err != nil {
		return err
	}

	app.SigChan <- sig

	return nil
}

/*
Edit
*/
func (app *App) Edit(reqName string) error {

	app.Database.Display(reqName)

	req := app.Database.Data[reqName]
	editorDefault, _ := json.MarshalIndent(req, "", "    ")

	for {
		content := ""

		menu := &survey.Editor{
			FileName:      "http-tanker-edit*.json",
			Default:       string(editorDefault),
			AppendDefault: true,
			HideDefault:   true,
		}

		err := survey.AskOne(menu, &content)
		if err != nil {
			app.ErrorHandler(err)
			return err
		}
		var updateReq core.Request
		if err := json.Unmarshal([]byte(content), &updateReq); err != nil {
			fmt.Println(color.ColorRed, fmt.Sprintf("Invalid JSON: %v", err.Error()), color.ColorReset)
			editorDefault = []byte(content)
			continue
		}

		if updateReq.Name != reqName {
			delete(app.Database.Data, reqName)
		}
		app.Database.Data[updateReq.Name] = updateReq
		if err := app.Database.Save(); err != nil {
			return err
		}

		sig := Signal{
			Sig:     SigReqCreate,
			Meta:    updateReq.Name,
			Display: true,
		}

		message := "The request " + reqName + " has been edited successfully"
		fmt.Println("")
		fmt.Println(color.ColorGreen, message, color.ColorReset)
		app.SigChan <- sig

		return nil
	}
}

/*
Delete
*/
func (app *App) Delete(reqName string) error {

	var menu = []*survey.Question{
		{
			Name: "confirmDelete",
			Prompt: &survey.Confirm{
				Message: "This will delete the request : " + reqName + ". Continue ?",
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
		app.ErrorHandler(err)
		return err
	}

	if answers.ConfirmDelete {
		if err := app.Database.Delete(reqName); err != nil {
			app.ErrorHandler(err)
			return err
		}
		message := "The request " + reqName + " was successfully deleted"
		fmt.Println("")
		fmt.Println(color.ColorGreen, message, color.ColorReset)
		fmt.Println("")
	}

	menu = []*survey.Question{
		{
			Name: "back",
			Prompt: &survey.Select{
				Options: []string{SigBackHome, SigBackRequests, SigExit},
			},
			Validate: survey.Required,
		},
	}

	back := struct {
		Back string
	}{}

	err = survey.Ask(menu, &back)
	if err != nil {
		app.ErrorHandler(err)
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

/*
Banner
*/
func Banner() {
	fmt.Print("\033[H\033[2J")
	hLine := strings.Repeat("─", core.BoxWidth)
	fmt.Printf("%s %s%s\n", color.ColorGrey, hLine, color.ColorReset)
	fmt.Print(string(bannerBytes))
	fmt.Printf(" %sversion : %v%s\n", color.ColorGrey, version, color.ColorReset)
	fmt.Printf("%s %s%s\n\n", color.ColorGrey, hLine, color.ColorReset)
}

/*
Error handler
*/

func (app *App) ErrorHandler(err error) error {
	fmtError := "ERROR : " + err.Error()
	fmt.Println(color.ColorRed, fmtError, color.ColorReset)

	menu := []*survey.Question{
		{
			Name: "back",
			Prompt: &survey.Select{
				Options: []string{SigBackHome, SigBackRequests, SigExit},
			},
			Validate: survey.Required,
		},
	}

	back := struct {
		Back string
	}{}

	if menuErr := survey.Ask(menu, &back); menuErr != nil {
		app.SigChan <- Signal{Sig: SigHome}
		return menuErr
	}

	sig := Signal{
		Sig: back.Back,
	}

	app.SigChan <- sig
	return nil
}

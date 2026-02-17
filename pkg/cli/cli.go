package cli

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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

type httpResult struct {
	resp core.Response
	err  error
}

type httpResultMsg httpResult

type spinnerModel struct {
	spinner spinner.Model
	done    bool
	result  httpResult
	callFn  func() (core.Response, error)
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		resp, err := m.callFn()
		return httpResultMsg{resp: resp, err: err}
	})
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case httpResultMsg:
		m.done = true
		m.result = httpResult(msg)
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + color.Yellow.Render("Executing request...")
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
Banner
*/
func Banner() {
	fmt.Print("\033[H\033[2J")
	hLine := strings.Repeat("─", core.BoxWidth)
	fmt.Println(color.Grey.Render(" " + hLine))
	fmt.Print(string(bannerBytes))
	fmt.Println(color.Grey.Render(fmt.Sprintf(" version: %v", version)))
	fmt.Println(color.Grey.Render(" " + hLine))
	fmt.Println()
}

/*
Error handler
*/

func (app *App) ErrorHandler(err error) error {
	fmt.Println(color.Red.Render("ERROR : " + err.Error()))

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

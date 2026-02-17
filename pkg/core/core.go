package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/PierreKieffer/http-tanker/pkg/color"
	"github.com/charmbracelet/lipgloss"
)

type AuthConfig struct {
	Type     string `json:"type"`               // "bearer", "basic", "api-key"
	Token    string `json:"token,omitempty"`     // pour bearer
	Username string `json:"username,omitempty"`  // pour basic
	Password string `json:"password,omitempty"`  // pour basic
	Key      string `json:"key,omitempty"`       // pour api-key
	Header   string `json:"header,omitempty"`    // nom du header pour api-key (défaut: "X-API-Key")
}

type Request struct {
	Name     string                 `json:"name"`
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
	Headers  map[string]interface{} `json:"headers"`
	Insecure bool                   `json:"insecure,omitempty"`
	Auth     *AuthConfig            `json:"auth,omitempty"`
}

type Database struct {
	DatabaseDir  string `json:"databaseDir"`
	DatabaseFile string `json:"databaseFile"`
	mu           sync.Mutex
	Data         map[string]Request `json:"data"`
}

/*
Init local database
*/
func (db *Database) InitDB() error {

	// Check database directory
	if _, err := os.Stat(db.DatabaseDir); os.IsNotExist(err) {
		err := os.Mkdir(db.DatabaseDir, 0750)
		if err != nil {
			return err
		}
	}

	err := db.Load()
	if err != nil {
		return err
	}

	return nil
}

/*
Load local database file
*/
func (db *Database) Load() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.loadLocked()
}

func (db *Database) loadLocked() error {

	// Check if database file exists
	if _, err := os.Stat(db.DatabaseFile); os.IsNotExist(err) {
		// Initialize json database with example data
		var data = map[string]Request{
			"get-example": {
				Name:   "get-example",
				Method: "GET",
				URL:    "https://httpbin.org/get",
				Params: map[string]interface{}{
					"foo":   "bar",
					"count": "42",
				},
				Headers: map[string]interface{}{
					"Accept": "application/json",
				},
			},
			"get-https-insecure": {
				Name:     "get-https-insecure",
				Method:   "GET",
				URL:      "https://self-signed.badssl.com/",
				Insecure: true,
				Headers: map[string]interface{}{
					"Accept": "text/html",
				},
			},
			"download-image-example": {
				Name:   "download-image-example",
				Method: "GET",
				URL:    "https://httpbin.org/image/png",
				Headers: map[string]interface{}{
					"Accept": "image/png",
				},
			},
			"post-example": {
				Name:   "post-example",
				Method: "POST",
				URL:    "https://httpbin.org/post",
				Payload: map[string]interface{}{
					"languages": []map[string]interface{}{
						{
							"name":            "Python",
							"staticallyTyped": false,
						},
						{
							"name":            "Javascript",
							"staticallyTyped": false,
						},
						{
							"name":            "Golang",
							"staticallyTyped": true,
						},
						{
							"name":            "Rust",
							"staticallyTyped": true,
						},
					},
					"foo":   "bar",
					"count": 42,
				},
				Headers: map[string]interface{}{
					"Content-Type":  "application/json",
					"Accept":        "application/json",
				},
			},
		}
		db.Data = data

		return db.saveLocked()
	}

	// Get data
	var data map[string]Request

	jsonFile, err := os.Open(db.DatabaseFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return err
	}

	db.Data = data

	return nil
}

/*
Save local database file
*/
func (db *Database) Save() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.saveLocked()
}

func (db *Database) saveLocked() error {
	buffer, err := json.Marshal(db.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(db.DatabaseFile, buffer, 0600)
}

/*
Reset local database file
*/
func (db *Database) Reset() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Data = map[string]Request{}
	return db.saveLocked()
}

/*
Delete a request
*/
func (db *Database) Delete(reqName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.Data, reqName)
	return db.saveLocked()
}

/*
Display request
*/
func (db *Database) Display(requestName string) {

	r := db.Data[requestName]

	var lines []string
	lines = append(lines, "Name   : "+r.Name)
	lines = append(lines, "Method : "+color.MethodStyle(r.Method).Render(r.Method))
	lines = append(lines, "URL    : "+r.URL)
	if len(r.Params) > 0 {
		jsonParams, _ := json.MarshalIndent(r.Params, "", "    ")
		lines = append(lines, "Params :\n"+string(jsonParams))
	}
	if len(r.Payload) > 0 {
		jsonPayload, _ := json.MarshalIndent(r.Payload, "", "    ")
		lines = append(lines, "Payload :\n"+string(jsonPayload))
	}
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.MarshalIndent(r.Headers, "", "    ")
		lines = append(lines, "Headers :\n"+string(jsonHeaders))
	}
	if r.Insecure {
		lines = append(lines, "Insecure : true (TLS verification skipped)")
	}
	if r.Auth != nil {
		switch r.Auth.Type {
		case "bearer":
			lines = append(lines, "Auth     : Bearer "+maskSecret(r.Auth.Token))
		case "basic":
			lines = append(lines, "Auth     : Basic "+r.Auth.Username+":"+maskSecret(r.Auth.Password))
		case "api-key":
			header := r.Auth.Header
			if header == "" {
				header = "X-API-Key"
			}
			lines = append(lines, "Auth     : API Key ["+header+"] "+maskSecret(r.Auth.Key))
		}
	}
	DrawBox("Request details", lines)
}
func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}

const BoxWidth = 50

var titleBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("8")).
	Foreground(lipgloss.Color("12")).
	Width(BoxWidth).
	Padding(0, 1)

func DrawBox(title string, content []string) {
	// Title box
	fmt.Println(titleBoxStyle.Render(title))

	// Content
	for _, line := range content {
		for _, sub := range strings.Split(line, "\n") {
			fmt.Println(" " + color.White.Render(sub))
		}
	}

	// Bottom separator
	hLine := strings.Repeat("─", BoxWidth)
	fmt.Println(color.Grey.Render(" " + hLine))
	fmt.Println()
}

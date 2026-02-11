package core

import (
	"encoding/json"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"io"
	"os"
	"strings"
	"sync"
)

type Request struct {
	Name     string                 `json:"name"`
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
	Headers  map[string]interface{} `json:"headers"`
	Insecure bool                   `json:"insecure,omitempty"`
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
				URL:    "http://localhost:8080/get",
				Params: map[string]interface{}{
					"foo":   "bar",
					"count": "42",
				},
				Headers: map[string]interface{}{
					"Authorization": "secret",
				},
			},
			"post-example": {
				Name:   "post-example",
				Method: "POST",
				URL:    "http://localhost:8080/post",
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
					"Authorization": "secret",
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
	if err := db.saveLocked(); err != nil {
		return err
	}
	return db.loadLocked()
}

/*
Display request
*/
func (db *Database) Display(requestName string) error {

	r := db.Data[requestName]

	var lines []string
	lines = append(lines, fmt.Sprintf("Name   : %s", r.Name))
	lines = append(lines, fmt.Sprintf("Method : %s%s%s", color.MethodColor(r.Method), r.Method, color.ColorReset))
	lines = append(lines, fmt.Sprintf("URL    : %s", r.URL))
	if len(r.Params) > 0 {
		jsonParams, _ := json.MarshalIndent(r.Params, "", "    ")
		lines = append(lines, fmt.Sprintf("Params :\n%s", string(jsonParams)))
	}
	if len(r.Payload) > 0 {
		jsonPayload, _ := json.MarshalIndent(r.Payload, "", "    ")
		lines = append(lines, fmt.Sprintf("Payload :\n%s", string(jsonPayload)))
	}
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.MarshalIndent(r.Headers, "", "    ")
		lines = append(lines, fmt.Sprintf("Headers :\n%s", string(jsonHeaders)))
	}
	if r.Insecure {
		lines = append(lines, "Insecure : true (TLS verification skipped)")
	}
	DrawBox("Request details", lines)

	return nil
}
const boxWidth = 50

func DrawBox(title string, content []string) {
	hLine := strings.Repeat("─", boxWidth)

	// Title box
	fmt.Printf("%s┌%s┐%s\n", color.ColorGrey, hLine, color.ColorReset)
	titlePad := boxWidth - len([]rune(title)) - 1
	if titlePad < 0 {
		titlePad = 0
	}
	fmt.Printf("%s│%s %s%*s%s│%s\n", color.ColorGrey, color.ColorBlue, title, titlePad, "", color.ColorGrey, color.ColorReset)
	fmt.Printf("%s└%s┘%s\n", color.ColorGrey, hLine, color.ColorReset)

	// Content
	for _, line := range content {
		for _, sub := range strings.Split(line, "\n") {
			fmt.Printf(" %s%s%s\n", color.ColorWhite, sub, color.ColorReset)
		}
	}

	// Bottom separator
	fmt.Printf("%s %s%s\n\n", color.ColorGrey, hLine, color.ColorReset)
}

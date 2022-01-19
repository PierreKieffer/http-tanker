package core

import (
	"encoding/json"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"io/ioutil"
	"os"
	"sync"
)

type Request struct {
	Name    string                 `json:"name"`
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Headers map[string]interface{} `json:"headers"`
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
		err := os.Mkdir(db.DatabaseDir, 0755)
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

	// Check if database file exists
	if _, err := os.Stat(db.DatabaseFile); os.IsNotExist(err) {
		// Initialize json database file
		jsonFile, err := os.OpenFile(db.DatabaseFile, os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		defer jsonFile.Close()

		var data = map[string]Request{}
		db.Data = data

		db.Save()

		return nil
	}

	// Get data
	var data map[string]Request

	jsonFile, err := os.Open(db.DatabaseFile)
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
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
	buffer, _ := json.Marshal(db.Data)
	ioutil.WriteFile(db.DatabaseFile, buffer, 0755)
	return nil
}

/*
Reset local database file
*/
func (db *Database) Reset() error {
	db.Data = map[string]Request{}
	db.Save()
	return nil
}

/*
Delete a request
*/
func (db *Database) Delete(reqName string) error {
	delete(db.Data, reqName)
	db.Save()
	db.Load()
	return nil
}

/*
Display request
*/
func (db *Database) Display(requestName string) error {

	r := db.Data[requestName]

	fmt.Println("")
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println(string(color.ColorBlue), "Request details : ", string(color.ColorReset))
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))

	name := fmt.Sprintf("Name : %s", r.Name)
	method := fmt.Sprintf("Method : %s", r.Method)
	url := fmt.Sprintf("URL : %s", r.URL)
	StringSeparatorDisplay(name)
	StringSeparatorDisplay(method)
	StringSeparatorDisplay(url)
	if len(r.Params) > 0 {
		jsonParams, _ := json.MarshalIndent(r.Params, "", "    ")
		params := fmt.Sprintf("Params :\n%s", string(jsonParams))
		StringSeparatorDisplay(params)
	}
	if len(r.Payload) > 0 {
		jsonPayload, _ := json.MarshalIndent(r.Payload, "", "    ")
		payload := fmt.Sprintf("Payload :\n%s", string(jsonPayload))
		StringSeparatorDisplay(payload)
	}
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.MarshalIndent(r.Headers, "", "    ")
		headers := fmt.Sprintf("Headers :\n%s", string(jsonHeaders))
		StringSeparatorDisplay(headers)
	}
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println("")

	return nil
}
func StringSeparatorDisplay(s string) {
	fmt.Println(string(color.ColorWhite), s, string(color.ColorReset))
}

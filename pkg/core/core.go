package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Request struct {
	Name    string                 `json:"name"`
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Params  map[string]interface{} `json:"params"`
	Payload map[string]interface{} `json:"payload"`
	Headers map[string]interface{} `json:"headers"`
}

type Database struct {
	DatabaseDir  string `json:"databaseDir"`
	DatabaseFile string `json:"databaseFile"`
	mu           sync.Mutex
	Data         map[string]Request `json:"data"`
}

/*
Init post-office local database
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
Save local database file
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
List all requests
*/
func (db *Database) List() error {

	return nil
}

/*
Edit a request
*/
func (r *Request) Edit() error {
	return nil
}

/*
Delete a request
*/
func (r *Request) Delete() error {
	return nil
}

/*
Display request
*/
func (db *Database) Display(requestName string) error {
	r := db.Data[requestName]
	fmt.Println("")
	fmt.Printf("Name : %s \n", r.Name)
	fmt.Printf("Method : %s \n", r.Method)
	fmt.Printf("URL : %s \n", r.URL)
	if len(r.Params) > 0 {
		jsonParams, _ := json.Marshal(r.Params)
		fmt.Printf("Params : %s \n", string(jsonParams))
	}
	if len(r.Payload) > 0 {
		jsonPayload, _ := json.Marshal(r.Payload)
		fmt.Printf("Payload : %s \n", string(jsonPayload))
	}
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.Marshal(r.Headers)
		fmt.Printf("Headers : %s \n", string(jsonHeaders))
	}
	fmt.Println("")
	return nil
}

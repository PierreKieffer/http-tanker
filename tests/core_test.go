package tests

import (
	"encoding/json"
	"github.com/PierreKieffer/http-tanker/pkg/core"
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Delete file if exist :

	err := os.Remove("/tmp/http-tanker-test/http-tanker-data-test.json")
	if err != nil {
		// file doesn't exist, so do nothing
	}

	database := &core.Database{
		DatabaseDir:  "/tmp/http-tanker-test",
		DatabaseFile: "/tmp/http-tanker-test/http-tanker-data-test.json",
	}

	err = database.InitDB()
	if err != nil {
		t.Errorf("TestInitDB failed")
	}

	jsonData, _ := json.Marshal(database.Data)

	must := `{"get-example":{"name":"get-example","method":"GET","url":"http://localhost:8080/get","params":{"count":"42","foo":"bar"},"headers":{"Authorization":"secret"}},"post-example":{"name":"post-example","method":"POST","url":"http://localhost:8080/post","payload":{"count":42,"foo":"bar","languages":[{"name":"Python","staticallyTyped":false},{"name":"Javascript","staticallyTyped":false},{"name":"Golang","staticallyTyped":true},{"name":"Rust","staticallyTyped":true}]},"headers":{"Authorization":"secret","Content-Type":"application/json"}}}`

	if must != string(jsonData) {
		t.Errorf("TestInitDB failed")
	}
}

func TestAddRequest(t *testing.T) {
	database := &core.Database{
		DatabaseDir:  "/tmp/http-tanker-test",
		DatabaseFile: "/tmp/http-tanker-test/http-tanker-data-test.json",
	}

	err := database.InitDB()
	if err != nil {
		t.Errorf("TestAddRequest failed")
	}

	newRequest := core.Request{
		Name:   "foobar",
		Method: "GET",
		URL:    "http://localhost:8080/get",
		Params: map[string]interface{}{
			"param_1": "amazing",
		},
		Headers: map[string]interface{}{
			"Authorization": "secret",
			"Useful":        "A useful header",
		},
	}

	database.Data[newRequest.Name] = newRequest

	err = database.Save()
	if err != nil {
		t.Errorf("TestAddRequest failed")
	}
	err = database.Load()
	if err != nil {
		t.Errorf("TestAddRequest failed")
	}

	jsonData, _ := json.Marshal(database.Data)

	must := `{"foobar":{"name":"foobar","method":"GET","url":"http://localhost:8080/get","params":{"param_1":"amazing"},"headers":{"Authorization":"secret","Useful":"A useful header"}},"get-example":{"name":"get-example","method":"GET","url":"http://localhost:8080/get","params":{"count":"42","foo":"bar"},"headers":{"Authorization":"secret"}},"post-example":{"name":"post-example","method":"POST","url":"http://localhost:8080/post","payload":{"count":42,"foo":"bar","languages":[{"name":"Python","staticallyTyped":false},{"name":"Javascript","staticallyTyped":false},{"name":"Golang","staticallyTyped":true},{"name":"Rust","staticallyTyped":true}]},"headers":{"Authorization":"secret","Content-Type":"application/json"}}}`

	if must != string(jsonData) {
		t.Errorf("TestAddRequest failed")
	}

}

func TestDeleteRequest(t *testing.T) {
	database := &core.Database{
		DatabaseDir:  "/tmp/http-tanker-test",
		DatabaseFile: "/tmp/http-tanker-test/http-tanker-data-test.json",
	}

	err := database.InitDB()
	if err != nil {
		t.Errorf("TestDeleteRequest failed")
	}

	database.Delete("foobar")

	jsonData, _ := json.Marshal(database.Data)

	must := `{"get-example":{"name":"get-example","method":"GET","url":"http://localhost:8080/get","params":{"count":"42","foo":"bar"},"headers":{"Authorization":"secret"}},"post-example":{"name":"post-example","method":"POST","url":"http://localhost:8080/post","payload":{"count":42,"foo":"bar","languages":[{"name":"Python","staticallyTyped":false},{"name":"Javascript","staticallyTyped":false},{"name":"Golang","staticallyTyped":true},{"name":"Rust","staticallyTyped":true}]},"headers":{"Authorization":"secret","Content-Type":"application/json"}}}`

	if must != string(jsonData) {
		t.Errorf("TestDeleteRequest failed")
	}
}

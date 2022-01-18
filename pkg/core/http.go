package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (r *Request) CallHTTP() error {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return err
	}

	switch r.Method {
	case "GET":
		if len(r.Params) > 0 {
			q := req.URL.Query()

			for k, v := range r.Params {
				q.Add(k, v.(string))
			}
			req.URL.RawQuery = q.Encode()
		}

	case "POST":
		jsonPayload, _ := json.Marshal(r.Payload)
		req, err = http.NewRequest(r.Method, r.URL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return err
		}

	}

	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			req.Header.Set(k, v.(string))
		}

	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	RespDisplay(resp)

	return nil
}

func RespDisplay(resp *http.Response) error {
	fmt.Printf("Status : %v \n", resp.Status)
	fmt.Printf("Status code : %v \n", resp.StatusCode)
	fmt.Printf("Protocol : %v \n", resp.Proto)
	if len(resp.Header) > 0 {
		jsonHeaders, _ := json.Marshal(resp.Header)
		fmt.Printf("Headers : %s \n", string(jsonHeaders))
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		fmt.Printf("Body : %v \n", string(bodyBytes))
	}

	return nil
}

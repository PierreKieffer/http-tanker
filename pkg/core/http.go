package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/color"
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
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println(string(color.ColorBlue), "Response details : ", string(color.ColorReset))
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	status := fmt.Sprintf("Status : %v", resp.Status)
	statusCode := fmt.Sprintf("Status code : %v", resp.StatusCode)
	proto := fmt.Sprintf("Protocol : %v", resp.Proto)
	StringSeparatorDisplay(status)
	StringSeparatorDisplay(statusCode)
	StringSeparatorDisplay(proto)
	if len(resp.Header) > 0 {
		jsonHeaders, _ := json.Marshal(resp.Header)
		headers := fmt.Sprintf("Headers : %s", string(jsonHeaders))
		StringSeparatorDisplay(headers)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		body := fmt.Sprintf("Body : %v", string(bodyBytes))
		StringSeparatorDisplay(body)
	}
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println("")

	return nil
}

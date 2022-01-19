package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"io/ioutil"
	"net/http"
	"time"
)

type Response struct {
	Status                string                 `json:"status,omitempty"`
	StatusCode            int                    `json:"statusCode,omitempty"`
	Proto                 string                 `json:"proto,omitempty"`
	Headers               http.Header            `json:"headers,omitempty"`
	JsonBody              map[string]interface{} `json:"jsonBody,omitempty"`
	Body                  string                 `json:"body,omitempty"`
	ExecutionTimeMillisec int64                  `json:"executionTimeMillisec,omitempty"`
}

func (r *Request) CallHTTP() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return "", err
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
			return "", err
		}

	}

	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			req.Header.Set(k, v.(string))
		}

	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	duration := time.Since(start)

	fmtResponse, _ := FmtResponse(resp, duration.Milliseconds())
	fmtStringResponse, _ := json.MarshalIndent(fmtResponse, "", "    ")

	return string(fmtStringResponse), nil
}

func FmtResponse(resp *http.Response, duration int64) (Response, error) {

	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println(string(color.ColorBlue), "Response details : ", string(color.ColorReset))
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	status := fmt.Sprintf("Status : %v", resp.Status)
	statusCode := fmt.Sprintf("Status code : %v", resp.StatusCode)
	proto := fmt.Sprintf("Protocol : %v", resp.Proto)
	execTime := fmt.Sprintf("Execution time : %v ms", duration)
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
	StringSeparatorDisplay(execTime)
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println("")

	var response = Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Headers:    resp.Header,
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(bodyBytes, &jsonResponse)
	if err != nil {
		response.Body = string(bodyBytes)
	} else {
		response.JsonBody = jsonResponse
	}
	response.ExecutionTime = duration

	return response, nil

}

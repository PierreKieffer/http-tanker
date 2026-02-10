package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	client := &http.Client{Timeout: 30 * time.Second}

	var body io.Reader
	switch r.Method {
	case "POST", "PUT":
		jsonPayload, err := json.Marshal(r.Payload)
		if err != nil {
			return "", err
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(r.Method, r.URL, body)
	if err != nil {
		return "", err
	}

	if len(r.Params) > 0 {
		q := req.URL.Query()
		for k, v := range r.Params {
			if s, ok := v.(string); ok {
				q.Add(k, s)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	response, err := BuildResponse(resp, duration.Milliseconds())
	if err != nil {
		return "", err
	}
	DisplayResponse(response)

	fmtStringResponse, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return "", err
	}

	return string(fmtStringResponse), nil
}

func BuildResponse(resp *http.Response, duration int64) (Response, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	response := Response{
		Status:                resp.Status,
		StatusCode:            resp.StatusCode,
		Proto:                 resp.Proto,
		Headers:               resp.Header,
		ExecutionTimeMillisec: duration,
	}

	var jsonResponse map[string]interface{}
	if json.Unmarshal(bodyBytes, &jsonResponse) == nil {
		response.JsonBody = jsonResponse
	} else {
		response.Body = string(bodyBytes)
	}

	return response, nil
}

func DisplayResponse(r Response) {
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println(string(color.ColorBlue), "Response details : ", string(color.ColorReset))
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	status := fmt.Sprintf("Status : %v", r.Status)
	statusCode := fmt.Sprintf("Status code : %v", r.StatusCode)
	proto := fmt.Sprintf("Protocol : %v", r.Proto)
	execTime := fmt.Sprintf("Execution time : %v ms", r.ExecutionTimeMillisec)
	StringSeparatorDisplay(status)
	StringSeparatorDisplay(statusCode)
	StringSeparatorDisplay(proto)
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.Marshal(r.Headers)
		headers := fmt.Sprintf("Headers : %s", string(jsonHeaders))
		StringSeparatorDisplay(headers)
	}
	if r.Body != "" {
		body := fmt.Sprintf("Body : %v", r.Body)
		StringSeparatorDisplay(body)
	} else if r.JsonBody != nil {
		jsonBody, _ := json.MarshalIndent(r.JsonBody, "", "    ")
		body := fmt.Sprintf("Body : %v", string(jsonBody))
		StringSeparatorDisplay(body)
	}
	StringSeparatorDisplay(execTime)
	fmt.Println(string(color.ColorGrey), "------------------------------------------------", string(color.ColorReset))
	fmt.Println("")
}

func (r *Request) CurlCommand() string {
	var parts []string
	parts = append(parts, "curl", "-X", r.Method)

	// Build URL with query params
	targetURL := r.URL
	if len(r.Params) > 0 {
		q := url.Values{}
		for k, v := range r.Params {
			if s, ok := v.(string); ok {
				q.Add(k, s)
			}
		}
		targetURL = targetURL + "?" + q.Encode()
	}
	parts = append(parts, fmt.Sprintf("'%s'", targetURL))

	// Headers
	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			if s, ok := v.(string); ok {
				parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", k, s))
			}
		}
	}

	// Body for POST/PUT
	switch r.Method {
	case "POST", "PUT":
		if len(r.Payload) > 0 {
			jsonPayload, _ := json.Marshal(r.Payload)
			parts = append(parts, "-d", fmt.Sprintf("'%s'", string(jsonPayload)))
		}
	}

	return strings.Join(parts, " \\\n  ")
}

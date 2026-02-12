package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/PierreKieffer/http-tanker/pkg/color"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

var (
	defaultClient = &http.Client{Timeout: 30 * time.Second}
	insecureClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func (r *Request) CallHTTP() (Response, error) {
	client := defaultClient
	if r.Insecure {
		client = insecureClient
	}

	var body io.Reader
	switch r.Method {
	case "POST", "PUT":
		jsonPayload, err := json.Marshal(r.Payload)
		if err != nil {
			return Response{}, err
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(r.Method, r.URL, body)
	if err != nil {
		return Response{}, err
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
		return Response{}, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	response, err := BuildResponse(resp, duration.Milliseconds())
	if err != nil {
		return Response{}, err
	}

	return response, nil
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
	statusColor := color.StatusCodeColor(r.StatusCode)
	var lines []string
	lines = append(lines, "Status         : "+statusColor+r.Status+color.ColorReset)
	lines = append(lines, "Status code    : "+statusColor+strconv.Itoa(r.StatusCode)+color.ColorReset)
	lines = append(lines, "Protocol       : "+r.Proto)
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.MarshalIndent(r.Headers, "", "    ")
		lines = append(lines, "Headers :\n"+string(jsonHeaders))
	}
	if r.Body != "" {
		lines = append(lines, "Body : "+r.Body)
	} else if r.JsonBody != nil {
		jsonBody, _ := json.MarshalIndent(r.JsonBody, "", "    ")
		lines = append(lines, "Body :\n"+string(jsonBody))
	}
	lines = append(lines, "Execution time : "+strconv.FormatInt(r.ExecutionTimeMillisec, 10)+" ms")
	DrawBox("Response details", lines)
}

func (r *Request) CurlCommand() string {
	parts := make([]string, 0, 6+2*len(r.Headers))
	parts = append(parts, "curl")
	if r.Insecure {
		parts = append(parts, "-k")
	}
	parts = append(parts, "-X", r.Method)

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
	parts = append(parts, "'"+targetURL+"'")

	// Headers
	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			if s, ok := v.(string); ok {
				parts = append(parts, "-H", "'"+k+": "+s+"'")
			}
		}
	}

	// Body for POST/PUT
	switch r.Method {
	case "POST", "PUT":
		if len(r.Payload) > 0 {
			jsonPayload, _ := json.Marshal(r.Payload)
			parts = append(parts, "-d", "'"+string(jsonPayload)+"'")
		}
	}

	return strings.Join(parts, " \\\n  ")
}

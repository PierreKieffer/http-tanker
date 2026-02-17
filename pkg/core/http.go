package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PierreKieffer/http-tanker/pkg/color"
)

type Response struct {
	Status                string                 `json:"status,omitempty"`
	StatusCode            int                    `json:"statusCode,omitempty"`
	Proto                 string                 `json:"proto,omitempty"`
	Headers               http.Header            `json:"headers,omitempty"`
	JsonBody              map[string]interface{} `json:"jsonBody,omitempty"`
	Body                  string                 `json:"body,omitempty"`
	ContentType           string                 `json:"contentType,omitempty"`
	BodySize              int64                  `json:"bodySize,omitempty"`
	ExecutionTimeMillisec int64                  `json:"executionTimeMillisec,omitempty"`
	savedFile             string
}

func (r *Response) IsBinaryContent() bool {
	return r.BodySize > 0 && r.savedFile != ""
}

func (r *Response) SaveToFile(path string) error {
	if r.savedFile == "" {
		return fmt.Errorf("no binary content to save")
	}
	// Try rename first (fast, no copy if same filesystem)
	if err := os.Rename(r.savedFile, path); err == nil {
		r.savedFile = ""
		return nil
	}
	// Fallback: copy + delete
	src, err := os.Open(r.savedFile)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	src.Close()
	os.Remove(r.savedFile)
	r.savedFile = ""
	return nil
}

func (r *Response) Cleanup() {
	if r.savedFile != "" {
		os.Remove(r.savedFile)
		r.savedFile = ""
	}
}

func IsTextContent(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.Index(ct, ";"); i != -1 {
		ct = strings.TrimSpace(ct[:i])
	}
	if strings.HasPrefix(ct, "text/") {
		return true
	}
	textTypes := []string{
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-javascript",
		"application/ecmascript",
		"application/xhtml+xml",
		"application/soap+xml",
		"application/rss+xml",
		"application/atom+xml",
		"application/svg+xml",
		"application/x-www-form-urlencoded",
	}
	for _, t := range textTypes {
		if ct == t {
			return true
		}
	}
	if strings.HasSuffix(ct, "+json") || strings.HasSuffix(ct, "+xml") {
		return true
	}
	return false
}

func formatSize(bytes int64) string {
	const (
		KB int64 = 1024
		MB       = KB * 1024
		GB       = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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
	case "POST", "PUT", "PATCH":
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

	if r.Auth != nil {
		switch r.Auth.Type {
		case "bearer":
			req.Header.Set("Authorization", "Bearer "+r.Auth.Token)
		case "basic":
			req.SetBasicAuth(r.Auth.Username, r.Auth.Password)
		case "api-key":
			header := r.Auth.Header
			if header == "" {
				header = "X-API-Key"
			}
			req.Header.Set(header, r.Auth.Key)
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
	contentType := resp.Header.Get("Content-Type")

	response := Response{
		Status:                resp.Status,
		StatusCode:            resp.StatusCode,
		Proto:                 resp.Proto,
		Headers:               resp.Header,
		ExecutionTimeMillisec: duration,
	}

	if IsTextContent(contentType) || contentType == "" {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return Response{}, err
		}
		var jsonResponse map[string]interface{}
		if json.Unmarshal(bodyBytes, &jsonResponse) == nil {
			response.JsonBody = jsonResponse
		} else {
			response.Body = string(bodyBytes)
		}
	} else {
		tmpFile, err := os.CreateTemp("", "http-tanker-*")
		if err != nil {
			return Response{}, fmt.Errorf("failed to create temp file: %w", err)
		}
		n, err := io.Copy(tmpFile, resp.Body)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpFile.Name())
			return Response{}, fmt.Errorf("failed to stream response body: %w", err)
		}
		response.ContentType = contentType
		response.BodySize = n
		response.savedFile = tmpFile.Name()
	}

	return response, nil
}

func DisplayResponse(r Response) {
	style := color.StatusCodeStyle(r.StatusCode)
	var lines []string
	lines = append(lines, "Status         : "+style.Render(r.Status))
	lines = append(lines, "Status code    : "+style.Render(strconv.Itoa(r.StatusCode)))
	lines = append(lines, "Protocol       : "+r.Proto)
	if len(r.Headers) > 0 {
		jsonHeaders, _ := json.MarshalIndent(r.Headers, "", "    ")
		lines = append(lines, "Headers :\n"+string(jsonHeaders))
	}
	if r.IsBinaryContent() {
		lines = append(lines, "Body           : [Binary content]")
		lines = append(lines, "Content-Type   : "+r.ContentType)
		lines = append(lines, "Size           : "+formatSize(r.BodySize))
	} else if r.Body != "" {
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

	// Auth
	if r.Auth != nil {
		switch r.Auth.Type {
		case "bearer":
			parts = append(parts, "-H", "'Authorization: Bearer "+r.Auth.Token+"'")
		case "basic":
			parts = append(parts, "-u", "'"+r.Auth.Username+":"+r.Auth.Password+"'")
		case "api-key":
			header := r.Auth.Header
			if header == "" {
				header = "X-API-Key"
			}
			parts = append(parts, "-H", "'"+header+": "+r.Auth.Key+"'")
		}
	}

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
	case "POST", "PUT", "PATCH":
		if len(r.Payload) > 0 {
			jsonPayload, _ := json.Marshal(r.Payload)
			parts = append(parts, "-d", "'"+string(jsonPayload)+"'")
		}
	}

	return strings.Join(parts, " \\\n  ")
}

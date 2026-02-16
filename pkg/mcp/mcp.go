package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PierreKieffer/http-tanker/pkg/core"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve creates the MCP server with all tools registered and starts stdio transport.
func Serve(db *core.Database) error {
	s := server.NewMCPServer(
		"http-tanker",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerTools(s, db)

	return server.ServeStdio(s)
}

func registerTools(s *server.MCPServer, db *core.Database) {
	s.AddTool(listRequestsTool(), listRequestsHandler(db))
	s.AddTool(getRequestTool(), getRequestHandler(db))
	s.AddTool(sendRequestTool(), sendRequestHandler(db))
	s.AddTool(sendCustomRequestTool(), sendCustomRequestHandler(db))
	s.AddTool(saveRequestTool(), saveRequestHandler(db))
	s.AddTool(deleteRequestTool(), deleteRequestHandler(db))
	s.AddTool(curlCommandTool(), curlCommandHandler(db))
}

// --- list_requests ---

func listRequestsTool() mcp.Tool {
	return mcp.NewTool("list_requests",
		mcp.WithDescription("List all saved HTTP requests with their names, methods, and URLs"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
	)
}

func listRequestsHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		type entry struct {
			Name   string `json:"name"`
			Method string `json:"method"`
			URL    string `json:"url"`
		}

		entries := make([]entry, 0, len(db.Data))
		for _, r := range db.Data {
			entries = append(entries, entry{
				Name:   r.Name,
				Method: r.Method,
				URL:    r.URL,
			})
		}

		return mcp.NewToolResultJSON(map[string]interface{}{
			"requests": entries,
		})
	}
}

// --- get_request ---

func getRequestTool() mcp.Tool {
	return mcp.NewTool("get_request",
		mcp.WithDescription("Get full details of a saved HTTP request by name"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the saved request")),
	)
}

func getRequestHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}

		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		r, ok := db.Data[name]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("request %q not found", name)), nil
		}

		return mcp.NewToolResultJSON(r)
	}
}

// --- send_request ---

func sendRequestTool() mcp.Tool {
	return mcp.NewTool("send_request",
		mcp.WithDescription("Execute a saved HTTP request by name and return the response. For binary responses (images, PDFs, archives...), only metadata is returned. Use output_file to save binary content to disk."),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the saved request to execute")),
		mcp.WithString("output_file", mcp.Description("File path to save binary response content (e.g. /tmp/image.png). Only used for binary responses.")),
	)
}

func sendRequestHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}

		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		r, ok := db.Data[name]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("request %q not found", name)), nil
		}

		resp, err := r.CallHTTP()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("HTTP request failed: %v", err)), nil
		}

		outputFile := request.GetString("output_file", "")
		return formatResponseResult(resp, outputFile)
	}
}

// --- send_custom_request ---

func sendCustomRequestTool() mcp.Tool {
	return mcp.NewTool("send_custom_request",
		mcp.WithDescription("Execute an ad-hoc HTTP request without saving it. For binary responses (images, PDFs, archives...), only metadata is returned. Use output_file to save binary content to disk."),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString("method", mcp.Required(), mcp.Description("HTTP method"), mcp.Enum("GET", "POST", "PUT", "DELETE")),
		mcp.WithString("url", mcp.Required(), mcp.Description("Target URL")),
		mcp.WithString("params", mcp.Description("Query parameters as a JSON object string, e.g. {\"key\": \"value\"}")),
		mcp.WithString("payload", mcp.Description("Request body as a JSON object string (for POST/PUT)")),
		mcp.WithString("headers", mcp.Description("HTTP headers as a JSON object string, e.g. {\"Content-Type\": \"application/json\"}")),
		mcp.WithBoolean("insecure", mcp.Description("Skip TLS certificate verification (default: false)")),
		mcp.WithString("output_file", mcp.Description("File path to save binary response content (e.g. /tmp/image.png). Only used for binary responses.")),
	)
}

func sendCustomRequestHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		method, err := request.RequireString("method")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: method"), nil
		}
		urlStr, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: url"), nil
		}

		r := core.Request{
			Method:   strings.ToUpper(method),
			URL:      urlStr,
			Headers:  map[string]interface{}{},
			Insecure: request.GetBool("insecure", false),
		}

		if err := parseOptionalJSON(request, "params", &r.Params); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid params JSON: %v", err)), nil
		}
		if err := parseOptionalJSON(request, "payload", &r.Payload); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid payload JSON: %v", err)), nil
		}
		if err := parseOptionalJSON(request, "headers", &r.Headers); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid headers JSON: %v", err)), nil
		}
		if r.Headers == nil {
			r.Headers = map[string]interface{}{}
		}

		resp, err := r.CallHTTP()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("HTTP request failed: %v", err)), nil
		}

		outputFile := request.GetString("output_file", "")
		return formatResponseResult(resp, outputFile)
	}
}

// --- save_request ---

func saveRequestTool() mcp.Tool {
	return mcp.NewTool("save_request",
		mcp.WithDescription("Save a new HTTP request to the database. This tool overwrites any existing request with the same name. Before calling this tool, you MUST ask the user if they want to add query parameters, a request body (payload), and/or headers. Always propose these optional fields explicitly during the creation workflow, even though they are not required."),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("name", mcp.Required(), mcp.Description("Unique name for the request")),
		mcp.WithString("method", mcp.Required(), mcp.Description("HTTP method"), mcp.Enum("GET", "POST", "PUT", "DELETE")),
		mcp.WithString("url", mcp.Required(), mcp.Description("Target URL")),
		mcp.WithString("params", mcp.Description("Query parameters as a JSON object string")),
		mcp.WithString("payload", mcp.Description("Request body as a JSON object string (for POST/PUT)")),
		mcp.WithString("headers", mcp.Description("HTTP headers as a JSON object string")),
		mcp.WithBoolean("insecure", mcp.Description("Skip TLS certificate verification (default: false)")),
	)
}

func saveRequestHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}
		method, err := request.RequireString("method")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: method"), nil
		}
		urlStr, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: url"), nil
		}

		r := core.Request{
			Name:     name,
			Method:   strings.ToUpper(method),
			URL:      urlStr,
			Headers:  map[string]interface{}{},
			Insecure: request.GetBool("insecure", false),
		}

		if err := parseOptionalJSON(request, "params", &r.Params); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid params JSON: %v", err)), nil
		}
		if err := parseOptionalJSON(request, "payload", &r.Payload); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid payload JSON: %v", err)), nil
		}
		if err := parseOptionalJSON(request, "headers", &r.Headers); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid headers JSON: %v", err)), nil
		}
		if r.Headers == nil {
			r.Headers = map[string]interface{}{}
		}

		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		db.Data[name] = r
		if err := db.Save(); err != nil {
			return nil, fmt.Errorf("failed to save database: %w", err)
		}

		return mcp.NewToolResultText(fmt.Sprintf("Request %q saved successfully", name)), nil
	}
}

// --- delete_request ---

func deleteRequestTool() mcp.Tool {
	return mcp.NewTool("delete_request",
		mcp.WithDescription("Delete a saved HTTP request by name"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the request to delete")),
	)
}

func deleteRequestHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}

		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		if _, ok := db.Data[name]; !ok {
			return mcp.NewToolResultError(fmt.Sprintf("request %q not found", name)), nil
		}

		if err := db.Delete(name); err != nil {
			return nil, fmt.Errorf("failed to delete request: %w", err)
		}

		return mcp.NewToolResultText(fmt.Sprintf("Request %q deleted successfully", name)), nil
	}
}

// --- curl_command ---

func curlCommandTool() mcp.Tool {
	return mcp.NewTool("curl_command",
		mcp.WithDescription("Generate the equivalent cURL command for a saved HTTP request"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the saved request")),
	)
}

func curlCommandHandler(db *core.Database) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: name"), nil
		}

		if err := db.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}

		r, ok := db.Data[name]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("request %q not found", name)), nil
		}

		return mcp.NewToolResultText(r.CurlCommand()), nil
	}
}

// --- helpers ---

func formatResponseResult(resp core.Response, outputFile string) (*mcp.CallToolResult, error) {
	contentType := resp.Headers.Get("Content-Type")
	if !core.IsTextContent(contentType) && contentType != "" {
		result := map[string]interface{}{
			"status":                resp.Status,
			"statusCode":            resp.StatusCode,
			"proto":                 resp.Proto,
			"headers":               resp.Headers,
			"contentType":           resp.ContentType,
			"bodySize":              resp.BodySize,
			"body":                  "[Binary content not included]",
			"executionTimeMillisec": resp.ExecutionTimeMillisec,
		}

		if outputFile != "" && resp.IsBinaryContent() {
			if err := resp.SaveToFile(outputFile); err != nil {
				result["saveError"] = err.Error()
			} else {
				result["savedTo"] = outputFile
			}
		}
		resp.Cleanup()

		return mcp.NewToolResultJSON(result)
	}
	return mcp.NewToolResultJSON(resp)
}

func parseOptionalJSON(request mcp.CallToolRequest, key string, target *map[string]interface{}) error {
	str := request.GetString(key, "")
	if str == "" {
		return nil
	}
	return json.Unmarshal([]byte(str), target)
}

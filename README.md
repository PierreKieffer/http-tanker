```
  _   _   _            _             _
 | |_| |_| |_ _ __ ___| |_ __ _ _ _ | |_____ _ _
 | ' \  _|  _| '_ \___|  _/ _` | ' \| / / -_) '_|
 |_||_\__|\__| .__/    \__\__,_|_||_|_\_\___|_|
             |_|
```

<div align="center">

A lightweight terminal HTTP client for API testing.

Create, manage and execute HTTP requests directly from your terminal. Includes a built-in MCP server for AI assistant integration.

<img src="./assets/tanker_demo.gif" width="1200" />

</div>

---

## Features

- **HTTP Methods** — Full support for GET, POST, PUT and DELETE requests with custom headers and JSON payloads
- **Request Management** — Create, edit, delete and browse saved requests. All persisted locally in JSON format
- **Response Inspector** — View response details including status, headers, body and execution time. Inspect in your editor
- **Binary Download** — Binary responses are streamed directly to disk without loading into memory. Save files to any location
- **cURL Export** — Generate the equivalent cURL command for any saved request. Copy and share effortlessly
- **HTTPS Insecure Mode** — Skip TLS certificate verification for testing with self-signed certificates

## Installation

### Pre-built binaries

Pre-built binaries for Linux and macOS (amd64 and arm64). No Go installation required.

Download the binary for your platform from the [`bin/`](./bin/) directory:

| Platform       | Binary                 |
|----------------|------------------------|
| Linux (amd64)  | `tanker-linux-amd64`   |
| Linux (arm64)  | `tanker-linux-arm64`   |
| macOS (amd64)  | `tanker-darwin-amd64`  |
| macOS (arm64)  | `tanker-darwin-arm64`  |

```bash
# Example for macOS arm64
cp bin/tanker-darwin-arm64 /usr/local/bin/tanker
chmod +x /usr/local/bin/tanker
```

### Build from source

Requires Go 1.21+.

```bash
git clone https://github.com/PierreKieffer/http-tanker.git
cd http-tanker
./build.sh v0.1.0
```

The binaries are generated in the `bin/` directory.

## Quick Start

Start http-tanker with the default database:

```bash
tanker
```

Use a custom database location:

```bash
tanker -db /path/to/custom/dir
```

Start the MCP server for AI assistant integration:

```bash
tanker --mcp
```

## CLI Examples

### Home menu

```
 ──────────────────────────────────────────────────
  _   _   _            _             _
 | |_| |_| |_ _ __ ___| |_ __ _ _ _ | |_____ _ _
 | ' \  _|  _| '_ \___|  _/ _` | ' \| / / -_) '_|
 |_||_\__|\__| .__/    \__\__,_|_||_|_\_\___|_|
             |_|
 version : edge
 ──────────────────────────────────────────────────

╭──────────────────────────────────────────────────╮
│ Home Menu                                        │
╰──────────────────────────────────────────────────╯
 ──────────────────────────────────────────────────

? Select :  [Use arrows to move, type to filter]
> Browse requests
  Create request
  About
  Exit
```

### Browse requests

```
╭──────────────────────────────────────────────────╮
│ Requests                                         │
╰──────────────────────────────────────────────────╯
 ──────────────────────────────────────────────────

? Select :  [Use arrows to move, type to filter]
> Back to Home Menu
  [GET] get-example - https://httpbin.org/get
  [GET] get-https-insecure - https://self-signed.badssl.com/
  [POST] post-example - https://httpbin.org/post
  [GET] download-image-example - https://httpbin.org/image/png
```

### Create request

```
? Name :  foobar
? Method :  POST
? URL :  https://foobar/post
? Payload (Enter the payload in json format {"key": "value"}) :
 [Enter to launch editor]
? Headers (Enter the headers in json format {"key": "value"}, default = {}) :
 {"Content-type": "application/json"}
? Skip TLS certificate verification ? (y/N) y
```

### Request details

```
╭──────────────────────────────────────────────────╮
│ Request details                                  │
╰──────────────────────────────────────────────────╯
 Name   : foobar
 Method : POST
 URL    : https://foobar/post
 Payload :
 {
     "foo": "bar"
 }
 Headers :
 {
     "Content-type": "application/json"
 }
 Insecure : true (TLS verification skipped)
 ──────────────────────────────────────────────────

?   [Use arrows to move, type to filter]
> Run
  cURL
  Edit
  Delete
  Back to requests
  Exit
```

### Response details

```
╭──────────────────────────────────────────────────╮
│ Response details                                 │
╰──────────────────────────────────────────────────╯
 Status         : 200 OK
 Status code    : 200
 Protocol       : HTTP/2.0
 Headers :
 {
     "Access-Control-Allow-Credentials": [
         "true"
     ],
     "Access-Control-Allow-Origin": [
         "*"
     ],
     "Content-Length": [
         "363"
     ],
     "Content-Type": [
         "application/json"
     ],
     "Date": [
         "Mon, 16 Feb 2026 13:56:06 GMT"
     ],
     "Server": [
         "gunicorn/19.9.0"
     ]
 }
 Body :
 {
     "args": {
         "count": "42",
         "foo": "bar"
     },
     "headers": {
         "Accept": "application/json",
         "Accept-Encoding": "gzip",
         "Host": "httpbin.org",
         "User-Agent": "Go-http-client/2.0",
         "X-Amzn-Trace-Id": "Root=1-699321f6-4de146b7397276213e32bb4d"
     },
     "origin": "5.48.194.113",
     "url": "https://httpbin.org/get?count=42&foo=bar"
 }
 Execution time : 426 ms
 ──────────────────────────────────────────────────

? Inspect response in editor ? (y/N) y
```

### Binary download

```
┌──────────────────────────────────────────────────┐
│ Response details                                 │
╰──────────────────────────────────────────────────╯
 Status         : 200 OK
 Content-Type   : image/png
 Size           : 8.1 KB
 Body           : [Binary content]
? Save file locally ? Yes
? Save to : ~/Downloads/image.png
 File saved to ~/Downloads/image.png
```

## MCP Server

http-tanker includes a built-in [MCP](https://modelcontextprotocol.io/) (Model Context Protocol) server, allowing AI assistants like Claude to manage and execute your HTTP requests through natural language.

### Available tools

| Tool | Description |
|------|-------------|
| `list_requests` | List all saved HTTP requests |
| `get_request` | Get full details of a saved request |
| `save_request` | Create or update a saved request |
| `delete_request` | Delete a saved request |
| `send_request` | Execute a saved request and return the response |
| `send_custom_request` | Execute an ad-hoc HTTP request without saving it |
| `curl_command` | Generate the equivalent cURL command for a saved request |

### Configuration (.mcp.json)

```json
{
  "mcpServers": {
    "http-tanker": {
      "type": "stdio",
      "command": "/path/to/tanker",
      "args": ["--mcp"]
    }
  }
}
```

### MCP Examples

Once configured, you can interact with http-tanker through your AI assistant:

- *"List all my saved requests"*
- *"Create a POST request to https://api.example.com/users with a JSON body"*
- *"Execute the get-example request"*
- *"Show me the cURL command for my post-example request"*
- *"Delete the request named old-test"*
- *"Download the image from download-image-example and save it to /tmp/image.png"*

The MCP server shares the same JSON database (`~/.http-tanker/http-tanker-data.json`) as the terminal UI, so requests created in one mode are available in the other.

## License

BSD 2-Clause License

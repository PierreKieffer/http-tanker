```
  _   _   _            _             _
 | |_| |_| |_ _ __ ___| |_ __ _ _ _ | |_____ _ _
 | ' \  _|  _| '_ \___|  _/ _` | ' \| / / -_) '_|
 |_||_\__|\__| .__/    \__\__,_|_||_|_\_\___|_|
             |_|
```

<div align="center">

A lightweight terminal HTTP client for API testing.

Create, manage and execute HTTP requests directly from your terminal.

<img src="./assets/tanker_demo.gif" width="1200" />

</div>

---

## Features

- **HTTP Methods** — GET, POST, PUT, DELETE
- **Request management** — Create, edit, delete and browse saved requests
- **Response inspector** — View response details (status, headers, body, execution time) and inspect in editor
- **cURL export** — Generate the equivalent cURL command for any saved request
- **HTTPS insecure mode** — Skip TLS certificate verification for self-signed certificates
- **Custom database path** — Store requests in a custom location with the `-db` flag
- **JSON persistence** — All requests are saved locally in JSON format

## Installation

### Pre-built binaries

Pre-built binaries are available for Linux and macOS (amd64 and arm64). No Go installation required.

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

## Usage

```bash
tanker
```

Custom database location:
```bash
tanker -db /path/to/custom/dir
```

## MCP Server

http-tanker includes a built-in [MCP](https://modelcontextprotocol.io/) (Model Context Protocol) server, allowing AI assistants like Claude to manage and execute your HTTP requests through natural language.

### Setup

Add the following configuration to your MCP client settings (e.g. `.mcp.json` for Claude Code):

```json
{
  "mcpServers": {
    "http-tanker": {
      "type": "stdio",
      "command": "go",
      "args": ["run", ".", "--mcp"]
    }
  }
}
```

Or using a pre-built binary:

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

### Usage examples

Once configured, you can interact with http-tanker through your AI assistant:

- *"List all my saved requests"*
- *"Create a POST request to https://api.example.com/users with a JSON body"*
- *"Execute the get-users request"*
- *"Show me the cURL command for my post-example request"*
- *"Delete the request named old-test"*

The MCP server shares the same JSON database (`~/.http-tanker/http-tanker-data.json`) as the terminal UI, so requests created in one mode are available in the other.

## Built with

- [AlecAivazis/survey](https://github.com/AlecAivazis/survey)

## License

BSD 2-Clause License

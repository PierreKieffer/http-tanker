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

[Documentation & Examples](<!-- TODO: add GitHub Pages URL -->)

</div>

---

## Installation

### Pre-built binaries

| Platform       | Binary                 |
|----------------|------------------------|
| Linux (amd64)  | `tanker-linux-amd64`   |
| Linux (arm64)  | `tanker-linux-arm64`   |
| macOS (amd64)  | `tanker-darwin-amd64`  |
| macOS (arm64)  | `tanker-darwin-arm64`  |

```bash
cp bin/tanker-darwin-arm64 /usr/local/bin/tanker
chmod +x /usr/local/bin/tanker
```

### Build from source

```bash
git clone https://github.com/PierreKieffer/http-tanker.git
cd http-tanker
./build.sh v0.1.0
```

## Quick Start

```bash
tanker
tanker -db /path/to/custom/dir
tanker --mcp
```

## MCP Configuration

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

## License

BSD 2-Clause License

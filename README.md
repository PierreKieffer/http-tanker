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

## Built with

- [AlecAivazis/survey](https://github.com/AlecAivazis/survey)

## License

BSD 2-Clause License

# AI Blog 部落格 CLI

A portable, zero-install AI development sandbox. Double-click to launch a browser-based terminal with Antigravity CLI, Python, Node.js, and Git — all self-contained, no system dependencies required.

## Features

- **Portable** — Everything lives in `./sandbox/`, no global installs
- **Web Terminal** — xterm.js in the browser with full PTY support (winpty on Windows, native PTY on macOS/Linux)
- **Double-click to start** — Auto-downloads tools on first run, opens browser
- **Image paste** — Paste screenshots directly into the terminal for Gemini analysis
- **Cross-platform** — Windows (amd64), macOS (amd64/arm64), Linux (amd64)
- **SKILL system** — Pre-installed marketing automation skills:
  - `marketing-content-factory` — friendly chat entry for non-technical users (5 modules: setup / write / examples / FAQ / create new persona)
  - `persona-writer` — generic writing engine (research → visuals → SEO → publish), reads persona configs from `personas/<slug>/persona.md`
  - `tool-status-checker` — environment verification + auto venv setup
  - Built-in persona: `mrs-lin-slow-travel` (林太, 58yo retired teacher, slow-travel cultural blog)

## Download (always latest)

- **Windows:** https://github.com/lionjin88912/ai-blog/releases/latest/download/ai-blog-windows-amd64.exe
- **All platforms:** https://github.com/lionjin88912/ai-blog/releases/latest

## Quick Start

### Windows

1. Download `ai-blog-windows-amd64.exe` from the link above
2. Double-click — it auto-downloads tools and opens a browser terminal
3. Click **Launch Antigravity**, then just start chatting

### macOS / Linux

```bash
chmod +x ai-blog-darwin-arm64   # or ai-blog-linux-amd64
./ai-blog-darwin-arm64
```

The browser opens automatically with a terminal connected to your shell.

## CLI Commands

```
ai-blog init      Configure API keys and workspace path
ai-blog setup     Download all tools to ./sandbox/
ai-blog shell     Open a local terminal with sandbox tools in PATH
ai-blog web       Open a browser-based terminal
ai-blog status    Show installed tool versions
ai-blog clean     Remove the sandbox directory
```

### Global Flags

```
-d, --dir string     Sandbox directory path (default "./sandbox")
-h, --help           Help for ai-blog
    --version        Print version
```

### `web` / `shell` Flags

```
-p, --port string    Web terminal port (default "8088")
    --shell string   Shell to use (macOS/Linux only)
                     Examples: --shell zsh, --shell /usr/local/bin/fish
                     Default: $SHELL or /bin/bash
```

> On Windows, the `--shell` flag is ignored — the portable Git Bash is always used.

## Architecture

```
ai-blog(.exe)          Single binary (Go)
├── cmd/                  CLI commands (cobra)
├── internal/
│   ├── config/           Config management (~/.ai-blog/config.json)
│   ├── toolchain/        Download & manage portable tools
│   └── web/              Web terminal server
│       ├── server.go         HTTP server + static files
│       ├── terminal.go       WebSocket ↔ shell bridge
│       ├── winpty_windows.go PTY via winpty.dll (Windows)
│       └── pty_unix.go       PTY via creack/pty (macOS/Linux)
└── sandbox/              Downloaded tools (auto-created)
    ├── bin/              Shim scripts (agy, git, node, python, uv)
    ├── node/             Node.js 22
    ├── python/           Python 3.12 (via uv)
    ├── git/              Portable Git for Windows
    ├── antigravity/      Antigravity CLI (agy)
    └── uv/               uv package manager
```

## Bundled Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Node.js | 22.x | Runtime for the Antigravity CLI |
| Antigravity CLI (agy) | latest | Google AI assistant |
| Python | 3.12 | Scripting, automation |
| uv | latest | Fast Python package manager |
| Git | Portable | Version control (Windows only, macOS/Linux use system git) |

## Development

### Build (Docker)

```bash
docker compose build --no-cache
docker compose run --rm build
```

Outputs all platform binaries to `./dist/`.

### Build (local, current platform only)

```bash
go build -o ai-blog .
```

### Project Structure

```
main.go              Entry point (double-click → web server, CLI → cobra)
cmd/                 CLI commands
internal/config/     Config with tests
internal/toolchain/  Tool download/install with tests
internal/web/        Web terminal (xterm.js + WebSocket + PTY)
```

## License

MIT

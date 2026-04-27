![togx](https://capsule-render.vercel.app/api?type=waving&height=180&color=0:0f766e,100:134e4a&text=togx%20v0.1.2&fontColor=ffffff&fontSize=42&desc=Pure%20Go%20%2B%20HTMX%20Todo%20App&descAlignY=72)

[![Release](https://img.shields.io/github/v/release/safwanehfaz/go-htmx-todo?style=for-the-badge)](https://github.com/safwanehfaz/go-htmx-todo/releases/tag/v0.1.2)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Workflow](https://img.shields.io/github/actions/workflow/status/safwanehfaz/go-htmx-todo/release.yml?style=for-the-badge&label=release%20workflow)](https://github.com/safwanehfaz/go-htmx-todo/actions/workflows/release.yml)
[![CLI](https://img.shields.io/badge/CLI-togx-0f766e?style=for-the-badge)](https://github.com/safwanehfaz/go-htmx-todo)

## Highlights

- New `togx` lifecycle CLI: `start`, `stop`, `quit`, `status`, `autostart`
- `stop --force` / `-f` for hard stop, graceful shutdown by default
- `Ctrl+C` exits cleanly and removes pid file
- Persistent todos stored in JSON (`$XDG_CONFIG_HOME/togx/todos.json`)
- Release pipeline keeps strict build gating before publish

## Architecture

```mermaid
flowchart LR
    C[CLI: togx] -->|start/stop/status| S[Server Process]
    B[Browser + HTMX] -->|HTTP| S
    S --> T[Templates]
    S --> D[JSON persistence]
```

## Quick Start

```bash
togx start --foreground
```

Open `http://127.0.0.1:8080`.

## Included binaries

- Linux: `amd64`, `arm64`, `armv7`
- macOS: `amd64`, `arm64`
- Windows: `amd64`, `arm64`
- Android/Termux: `arm64-v8a`, `armv7`

## Verify downloads

```bash
sha256sum -c checksums.txt
```

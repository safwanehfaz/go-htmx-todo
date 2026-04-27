# togx

[![Release](https://img.shields.io/github/v/release/safwanehfaz/go-htmx-todo?style=for-the-badge)](https://github.com/safwanehfaz/go-htmx-todo/releases)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Workflow](https://img.shields.io/github/actions/workflow/status/safwanehfaz/go-htmx-todo/release.yml?style=for-the-badge&label=release)](https://github.com/safwanehfaz/go-htmx-todo/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/safwanehfaz/go-htmx-todo?style=for-the-badge)](LICENSE)

Minimal Todo app with pure Go + HTMX, no backend framework, no ORM, and no extra Go libraries.

Repository: `https://github.com/safwanehfaz/go-htmx-todo`

Version source of truth: `VERSION`

## Why togx

- Fast startup, one binary, simple deployment
- Interactive server-rendered UI via HTMX
- Persistent todos in JSON (`$XDG_CONFIG_HOME/togx/todos.json`)
- CLI lifecycle controls (`start`, `stop`, `quit`, `status`, `autostart`)
- Colorized logs for runtime visibility

## Quick start

```bash
go build -o togx .
./togx start --foreground
```

Open `http://127.0.0.1:8080`.

## CLI commands

```bash
# foreground (Ctrl+C graceful stop)
togx start --foreground

# background
togx start --detach

# status
togx status

# graceful stop
togx stop

# force stop
togx stop --force
togx stop -f

# alias
togx quit

# linux autostart (systemd --user)
togx autostart enable
togx autostart status
togx autostart disable
```

## Runtime behavior

- `Ctrl+C` gracefully shuts down and removes pid file
- `--force` sends hard kill
- No hidden daemon unless you explicitly run `--detach` or `autostart enable`

## API routes

- `GET /` full page
- `POST /todos` add item, returns list fragment
- `POST /todos/{id}/toggle` toggle item, returns list fragment
- `DELETE /todos/{id}` delete item, returns list fragment
- `GET /healthz` simple health check

## Build targets

| Platform | Architectures |
|---|---|
| Linux | `amd64`, `arm64`, `armv7` |
| macOS | `amd64`, `arm64` |
| Windows | `amd64`, `arm64` |
| Android/Termux | `arm64-v8a`, `armv7` |

Build all locally:

```bash
task build:all
# or
bash scripts/build-all.sh
```

## Releases

- Tag format: `vX.Y.Z`
- CI builds archives + checksums and publishes GitHub Release assets
- If `.github/release-notes-vX.Y.Z.md` exists, CI uses it automatically for release body

## Versioning automation

```bash
task release
# or
bash scripts/release.sh
```

This helper will:

- read `VERSION`
- sync `packaging/aur/PKGBUILD` `pkgver`
- regenerate `packaging/aur/.SRCINFO`
- create `Release vX.Y.Z` commit (when needed)
- create and push `vX.Y.Z` tag

Safety gate in CI (tag builds):

- Git tag version, `VERSION`, and `PKGBUILD pkgver` must match
- mismatch => workflow fails and no release upload

## AUR

Files: `packaging/aur/PKGBUILD`, `packaging/aur/.SRCINFO`

Before publishing to AUR:

1. set final `sha256sums` in `PKGBUILD` (replace `SKIP`)
2. regenerate `.SRCINFO` with `makepkg --printsrcinfo > .SRCINFO`
3. push to your AUR package repo

## Project layout

- `main.go` app server, templates, CLI lifecycle, JSON persistence
- `Taskfile.yml` task runner commands
- `scripts/build-all.sh` cross-platform build script
- `scripts/release.sh` version+tag release helper
- `.github/workflows/release.yml` CI release workflow
- `packaging/aur/` AUR metadata

## License

MIT (`LICENSE`).

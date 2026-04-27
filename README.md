# todo-go-htmx

Repository: `https://github.com/safwanehfaz/go-htmx-todo`

Minimal Todo app built with pure Go (`net/http`, `html/template`) and HTMX.

No framework, no ORM, no backend plugin system, no extra Go libraries.

## Features

- Add, toggle, and delete todos without page reloads (HTMX requests)
- Server-rendered HTML templates
- In-memory todo storage with mutex for safe concurrent access
- Single binary deployment

## Stack

- Go 1.22+
- HTMX loaded from CDN in HTML
- Standard library only on backend

## Run locally

```bash
go run .
```

Open `http://localhost:8080`.

## Build

```bash
go build -o todo-go-htmx .
```

## HTTP routes

- `GET /` -> full page
- `POST /todos` -> add item, returns todo list fragment
- `POST /todos/{id}/toggle` -> toggle complete, returns todo list fragment
- `DELETE /todos/{id}` -> delete item, returns todo list fragment

## Cross-platform binaries

This repo includes:

- GitHub Actions workflow: `.github/workflows/release.yml`
- Local cross-build script: `scripts/build-all.sh`
- Go-task file (ninja-like task runner written in Go): `Taskfile.yml`

Targets produced:

- Linux: `x86_64`, `arm64`, `armv7`
- macOS: `x86_64`, `arm64`
- Windows: `x86_64`, `arm64`
- Android (Termux): `arm64-v8a`, `armeabi-v7a`

Android notes (Termux):

- Artifacts are built with `GOOS=android`
- `arm64-v8a` uses `GOARCH=arm64`
- `armeabi-v7a` uses `GOARCH=arm` and `GOARM=7`
- `armeabi-v7a` needs Android NDK clang for cgo; set `ANDROID_NDK_HOME` (or `CC_ANDROID_ARMV7`) to enable it. If not found, that one target is skipped.

## Local release builds

With Go-task:

```bash
task build:all
```

Or directly:

```bash
bash scripts/build-all.sh
```

Artifacts are written to `dist/`.

## GitHub release workflow

The release workflow runs on:

- tag push: `v*.*.*`
- manual dispatch

It compiles all targets, creates archives, uploads workflow artifacts, and (for tags) creates/updates a GitHub release with all binaries and a checksum file.

## AUR packaging (Arch Linux)

AUR files are included in `packaging/aur/`:

- `PKGBUILD`
- `.SRCINFO`

Steps before publishing to AUR:

1. Update `url` and source URL to your real GitHub repo.
2. Set real checksums (replace `SKIP`) using `sha256sum`.
3. Update `pkgver` for each new release.
4. Regenerate `.SRCINFO`:

```bash
cd packaging/aur
makepkg --printsrcinfo > .SRCINFO
```

Then publish to your AUR package repository.

## Project structure

- `main.go` - app server, templates, handlers, in-memory store
- `Taskfile.yml` - Go-task commands
- `scripts/build-all.sh` - local multi-target build script
- `.github/workflows/release.yml` - CI release automation
- `packaging/aur/` - AUR metadata

## License

MIT. See `LICENSE`.

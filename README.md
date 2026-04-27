# togx

Repository: `https://github.com/safwanehfaz/go-htmx-todo`

Version: `0.1.2` (source of truth in `VERSION`)

Minimal Todo app built with pure Go (`net/http`, `html/template`) and HTMX.

No framework, no ORM, no backend plugin system, no extra Go libraries.

## Features

- Add, toggle, and delete todos without page reloads (HTMX requests)
- Server-rendered HTML templates
- In-memory todo storage with mutex for safe concurrent access
- Single binary deployment (`togx`)
- Persistent todo storage in JSON (`$XDG_CONFIG_HOME/togx/todos.json`)
- CLI lifecycle controls: `start`, `stop`, `quit`, `status`, `autostart`
- Colorized terminal logs for key lifecycle/events

## Stack

- Go 1.22+
- HTMX loaded from CDN in HTML
- Standard library only on backend

## Run locally

```bash
go run . start --foreground
```

Open `http://127.0.0.1:8080`.

## Build

```bash
go build -o togx .

./togx start --foreground
```

## HTTP routes

- `GET /` -> full page
- `POST /todos` -> add item, returns todo list fragment
- `POST /todos/{id}/toggle` -> toggle complete, returns todo list fragment
- `DELETE /todos/{id}` -> delete item, returns todo list fragment

## CLI usage

```bash
# Start in foreground (Ctrl+C to stop)
togx start --foreground

# Start detached/background
togx start --detach

# Check status
togx status

# Graceful stop
togx stop

# Force stop
togx stop --force
togx stop -f

# Alias for stop
togx quit

# Linux systemd user autostart
togx autostart enable
togx autostart status
togx autostart disable
```

Notes:

- `Ctrl+C` triggers graceful shutdown and removes pid file.
- `stop --force` uses hard kill.
- No hidden background daemon is started unless you run `--detach` or `autostart enable`.

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
- `armeabi-v7a` needs Android NDK clang for cgo; set `ANDROID_NDK_HOME`/`ANDROID_NDK_ROOT` (or `CC_ANDROID_ARMV7`) to enable it.
- You can force failure when Android armv7 is unavailable by setting `REQUIRE_ANDROID_ARMV7=1`.

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

## Release versioning

- Canonical version lives in `VERSION`.
- Keep these in sync when releasing:
  - `VERSION`
  - `packaging/aur/PKGBUILD` (`pkgver`)
  - `packaging/aur/.SRCINFO` (regenerate from `PKGBUILD`)
- Suggested release flow:

```bash
ver="$(cat VERSION)"
git add VERSION packaging/aur/PKGBUILD packaging/aur/.SRCINFO
git commit -m "Release v${ver}"
git tag -a "v${ver}" -m "v${ver}"
git push && git push origin "v${ver}"
```

- Automated release helper:

```bash
bash scripts/release.sh
# or
task release
```

`scripts/release.sh` will:

- read `VERSION`
- sync `packaging/aur/PKGBUILD` `pkgver`
- regenerate `packaging/aur/.SRCINFO`
- create `Release vX.Y.Z` commit when needed
- create and push `vX.Y.Z` tag

Release workflow safety check:

- On tag builds, CI validates all three versions match:
  - Git tag (`vX.Y.Z`)
  - `VERSION`
  - `packaging/aur/PKGBUILD` `pkgver`
- If mismatched, release fails and no assets are published.

## Project structure

- `main.go` - app server, templates, CLI lifecycle manager, JSON persistence
- `Taskfile.yml` - Go-task commands
- `scripts/build-all.sh` - local multi-target build script
- `.github/workflows/release.yml` - CI release automation
- `packaging/aur/` - AUR metadata

## License

MIT. See `LICENSE`.

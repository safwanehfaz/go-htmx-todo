![togx](https://capsule-render.vercel.app/api?type=waving&height=180&color=0:0f766e,100:134e4a&text=togx%20v0.1.3&fontColor=ffffff&fontSize=42&desc=Pure%20Go%20%2B%20HTMX%20Todo%20App&descAlignY=72)

[![Release](https://img.shields.io/github/v/release/safwanehfaz/go-htmx-todo?style=for-the-badge)](https://github.com/safwanehfaz/go-htmx-todo/releases/tag/v0.1.3)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Workflow](https://img.shields.io/github/actions/workflow/status/safwanehfaz/go-htmx-todo/release.yml?style=for-the-badge&label=release)](https://github.com/safwanehfaz/go-htmx-todo/actions/workflows/release.yml)

## Highlights

- More interactive `README.md` with quick-start, command matrix, and release flow
- CI now publishes rich notes to GitHub Releases automatically from `.github/release-notes-vX.Y.Z.md`
- Versioning flow remains strict: tag, `VERSION`, and AUR `pkgver` must match

## Quick commands

```bash
togx start --foreground
togx start --detach
togx status
togx stop
togx stop --force
```

## Platform artifacts

- Linux: `amd64`, `arm64`, `armv7`
- macOS: `amd64`, `arm64`
- Windows: `amd64`, `arm64`
- Android/Termux: `arm64-v8a`, `armv7`

## Verify integrity

```bash
sha256sum -c checksums.txt
```

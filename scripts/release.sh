#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION_FILE="$ROOT_DIR/VERSION"
PKGBUILD_FILE="$ROOT_DIR/packaging/aur/PKGBUILD"

NO_PUSH=0

usage() {
  printf "Usage: %s [--no-push]\n" "$(basename "$0")"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-push)
      NO_PUSH=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf "Unknown option: %s\n" "$1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ ! -f "$VERSION_FILE" ]]; then
  printf "VERSION file not found at %s\n" "$VERSION_FILE" >&2
  exit 1
fi

VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf "Invalid VERSION: %s (expected x.y.z)\n" "$VERSION" >&2
  exit 1
fi

cd "$ROOT_DIR"

if [[ -n "$(git status --porcelain)" ]]; then
  printf "Working tree is not clean. Commit or stash changes first.\n" >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/v${VERSION}" >/dev/null; then
  printf "Tag v%s already exists locally.\n" "$VERSION" >&2
  exit 1
fi

sed -i "s/^pkgver=.*/pkgver=${VERSION}/" "$PKGBUILD_FILE"

(
  cd "$ROOT_DIR/packaging/aur"
  makepkg --printsrcinfo > .SRCINFO
)

git add VERSION packaging/aur/PKGBUILD packaging/aur/.SRCINFO

if ! git diff --cached --quiet; then
  git commit -m "Release v${VERSION}"
fi

git tag -a "v${VERSION}" -m "v${VERSION}"

if [[ "$NO_PUSH" == "1" ]]; then
  printf "Created commit/tag locally. Skipped push (--no-push).\n"
  exit 0
fi

git push
git push origin "v${VERSION}"

printf "Release v%s pushed.\n" "$VERSION"

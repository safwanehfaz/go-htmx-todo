#!/usr/bin/env bash

set -euo pipefail

APP_NAME="todo-go-htmx"
OUT_DIR="dist"

require_android_armv7="${REQUIRE_ANDROID_ARMV7:-0}"

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

build() {
  local goos="$1"
  local goarch="$2"
  local suffix="$3"
  local goarm="${4:-}"

  local target_dir="$OUT_DIR/$suffix"
  mkdir -p "$target_dir"

  local bin_name="$APP_NAME"
  if [[ "$goos" == "windows" ]]; then
    bin_name="${bin_name}.exe"
  fi

  echo "Building $suffix"
  if [[ -n "$goarm" ]]; then
    GOOS="$goos" GOARCH="$goarch" GOARM="$goarm" CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$target_dir/$bin_name" .
  else
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$target_dir/$bin_name" .
  fi

  if [[ "$goos" == "windows" ]]; then
    (cd "$OUT_DIR" && zip -rq "${suffix}.zip" "$suffix")
  else
    tar -C "$OUT_DIR" -czf "$OUT_DIR/${suffix}.tar.gz" "$suffix"
  fi
}

build_android_armv7() {
  local suffix="android-armv7"
  local target_dir="$OUT_DIR/$suffix"
  mkdir -p "$target_dir"

  local cc=""
  if [[ -n "${CC_ANDROID_ARMV7:-}" ]]; then
    cc="$CC_ANDROID_ARMV7"
  else
    local ndk_home="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
    if [[ -n "$ndk_home" ]]; then
      local candidate_glob
      for candidate_glob in "$ndk_home"/toolchains/llvm/prebuilt/*/bin/armv7a-linux-androideabi24-clang; do
        if [[ -x "$candidate_glob" ]]; then
          cc="$candidate_glob"
          break
        fi
      done
    fi
  fi

  if [[ -z "$cc" ]]; then
    if [[ "$require_android_armv7" == "1" ]]; then
      echo "ERROR: $suffix required but NDK clang not found."
      return 1
    fi
    echo "Skipping $suffix (NDK clang not found)."
    return 0
  fi

  echo "Building $suffix"
  GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 CC="$cc" go build -trimpath -ldflags='-s -w' -o "$target_dir/$APP_NAME" .
  tar -C "$OUT_DIR" -czf "$OUT_DIR/${suffix}.tar.gz" "$suffix"
}

# Linux
build linux amd64 linux-amd64
build linux arm64 linux-arm64
build linux arm linux-armv7 7

# macOS
build darwin amd64 darwin-amd64
build darwin arm64 darwin-arm64

# Windows
build windows amd64 windows-amd64
build windows arm64 windows-arm64

# Android (Termux)
build android arm64 android-arm64-v8a
build_android_armv7

echo "Done. Artifacts in $OUT_DIR"

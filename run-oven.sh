#!/bin/sh
set -e

BASE_URL="https://github.com/clay-doc/clay-oven/releases/latest/download"

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)
    case "$ARCH" in
      x86_64)  FILE="clay-oven-linux-amd64" ;;
      aarch64) FILE="clay-oven-linux-arm64" ;;
      *)       echo "Unsupported Linux architecture: $ARCH" >&2; exit 1 ;;
    esac
    ;;
  Darwin)
    case "$ARCH" in
      x86_64)  FILE="clay-oven-darwin-amd64" ;;
      arm64)   FILE="clay-oven-darwin-arm64" ;;
      *)       echo "Unsupported macOS architecture: $ARCH" >&2; exit 1 ;;
    esac
    ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT)
    case "$ARCH" in
      x86_64)  FILE="clay-oven-windows-amd64.exe" ;;
      *)       echo "Unsupported Windows architecture: $ARCH" >&2; exit 1 ;;
    esac
    ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

URL="${BASE_URL}/${FILE}"
DEST="./${FILE}"

echo "Downloading ${URL}..."
if command -v curl >/dev/null 2>&1; then
  curl -fSL -o "$DEST" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$DEST" "$URL"
else
  echo "Error: neither curl nor wget found" >&2
  exit 1
fi

chmod +x "$DEST"
exec "$DEST" "$@"

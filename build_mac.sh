#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

mkdir -p /tmp/gocache /tmp/gotmp
export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOTMPDIR="${GOTMPDIR:-/tmp/gotmp}"
export CGO_ENABLED="${CGO_ENABLED:-0}"

DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
mkdir -p "$DIST_DIR"

GO_BIN="${GO_BIN:-}"
if [ -z "$GO_BIN" ]; then
    if command -v go >/dev/null 2>&1; then
        GO_BIN="$(command -v go)"
    else
        echo "go compiler not found" >&2
        exit 1
    fi
fi

HOST_GOOS="$("$GO_BIN" env GOOS)"
HOST_GOARCH="$("$GO_BIN" env GOARCH)"

if [ "$HOST_GOOS" != "darwin" ]; then
    echo "build_mac.sh must run on macOS, current host: $HOST_GOOS/$HOST_GOARCH" >&2
    exit 1
fi

OUTPUT_PATH="$DIST_DIR/replive_${HOST_GOOS}_${HOST_GOARCH}"

echo "using go: $GO_BIN"
"$GO_BIN" version
echo "building native mac binary: $HOST_GOOS/$HOST_GOARCH"
echo "CGO_ENABLED=$CGO_ENABLED"

"$GO_BIN" test ./... -run TestDoesNotExist
"$GO_BIN" build -o "$OUTPUT_PATH" .

cp "$OUTPUT_PATH" "$ROOT_DIR/replive"
chmod +x "$ROOT_DIR/replive"

echo "build finished:"
echo "  dist: $OUTPUT_PATH"
echo "  runtime: $ROOT_DIR/replive"

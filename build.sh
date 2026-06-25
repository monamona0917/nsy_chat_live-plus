#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

mkdir -p /tmp/gocache /tmp/gotmp
export GOCACHE="${GOCACHE:-/tmp/gocache}"
export GOTMPDIR="${GOTMPDIR:-/tmp/gotmp}"
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

echo "using go: $GO_BIN"
"$GO_BIN" version

HOST_GOOS="$("$GO_BIN" env GOOS)"
HOST_GOARCH="$("$GO_BIN" env GOARCH)"
TARGETS="${TARGETS:-$HOST_GOOS/$HOST_GOARCH}"
LDFLAGS="${LDFLAGS:-}"
CGO_ENABLED_VALUE="${CGO_ENABLED:-0}"

echo "host target: $HOST_GOOS/$HOST_GOARCH"
echo "build targets: $TARGETS"

"$GO_BIN" test ./... -run '^$'

build_target() {
    local target="$1"
    local goos="${target%/*}"
    local goarch="${target#*/}"
    local -a build_args

    if [ -z "$goos" ] || [ -z "$goarch" ] || [ "$goos" = "$goarch" ]; then
        echo "invalid target: $target (expected os/arch)" >&2
        exit 1
    fi

    local output_name="replive_${goos}_${goarch}"
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    local output_path="$DIST_DIR/$output_name"

    echo "building $target -> $output_path"
    build_args=(build -o "$output_path" .)
    if [ -n "$LDFLAGS" ]; then
        build_args=(build -ldflags "$LDFLAGS" -o "$output_path" .)
    fi
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED="$CGO_ENABLED_VALUE" \
        "$GO_BIN" "${build_args[@]}"

    if [ "$goos" = "$HOST_GOOS" ] && [ "$goarch" = "$HOST_GOARCH" ] && [ "$goos" != "windows" ]; then
        cp "$output_path" "$ROOT_DIR/replive"
        chmod +x "$ROOT_DIR/replive"
        echo "updated runtime binary: $ROOT_DIR/replive"
    fi
}

for target in $TARGETS; do
    build_target "$target"
done

echo "build finished:"
ls -1 "$DIST_DIR"

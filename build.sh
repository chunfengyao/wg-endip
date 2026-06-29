#!/bin/bash
set -e

# Output directory
BUILD_DIR="build"
BINARY_NAME="wg-endip"

# Supported architectures list (Removed Android, pure cross-compilation focus)
ALL_ARCHS=(
    "darwin-amd64" "darwin-arm64"
    "linux-amd64" "linux-arm64" "linux-armv5" "linux-armv6" "linux-armv7"
    "linux-loong64" "linux-mips-softfloat" "linux-mips64-softfloat"
    "linux-mips64le" "linux-mipsle-softfloat" "linux-mipsle"
    "linux-ppc64le" "linux-riscv64" "linux-s390x"
    "windows-amd64" "windows-arm64"
)

usage() {
    echo "Usage: $0 [--all | <arch>]"
    echo "Supported architectures:"
    for arch in "${ALL_ARCHS[@]}"; do
        echo "  $arch"
    done
    exit 1
}

build_platform() {
    local arch=$1
    local GOOS GOARCH GOARM GOMIPS CGO_ENABLED=0
    
    # Mapping architecture to build environment
    # CGO_ENABLED=0 enables pure Go builds for maximum portability
    case $arch in
        darwin-amd64) GOOS=darwin; GOARCH=amd64 ;;
        darwin-arm64) GOOS=darwin; GOARCH=arm64 ;;
        linux-amd64) GOOS=linux; GOARCH=amd64 ;;
        linux-arm64) GOOS=linux; GOARCH=arm64 ;;
        linux-armv5) GOOS=linux; GOARCH=arm; GOARM=5 ;;
        linux-armv6) GOOS=linux; GOARCH=arm; GOARM=6 ;;
        linux-armv7) GOOS=linux; GOARCH=arm; GOARM=7 ;;
        linux-loong64) GOOS=linux; GOARCH=loong64 ;;
        linux-mips-softfloat) GOOS=linux; GOARCH=mips; GOMIPS=softfloat ;;
        linux-mips64-softfloat) GOOS=linux; GOARCH=mips64; GOMIPS=softfloat ;;
        linux-mips64le) GOOS=linux; GOARCH=mips64le ;;
        linux-mipsle-softfloat) GOOS=linux; GOARCH=mipsle; GOMIPS=softfloat ;;
        linux-mipsle) GOOS=linux; GOARCH=mipsle ;;
        linux-ppc64le) GOOS=linux; GOARCH=ppc64le ;;
        linux-riscv64) GOOS=linux; GOARCH=riscv64 ;;
        linux-s390x) GOOS=linux; GOARCH=s390x ;;
        windows-amd64) GOOS=windows; GOARCH=amd64 ;;
        windows-arm64) GOOS=windows; GOARCH=arm64 ;;
        *) echo "Unknown arch: $arch"; exit 1 ;;
    esac

    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=$CGO_ENABLED
    [ -n "$GOARM" ] && export GOARM=$GOARM || unset GOARM
    [ -n "$GOMIPS" ] && export GOMIPS=$GOMIPS || unset GOMIPS

    OUTPUT="$BUILD_DIR/${BINARY_NAME}_${arch}"
    
    echo "Building for $arch (CGO=$CGO_ENABLED) -> $OUTPUT"
    go build -ldflags="-s -w" -o "$OUTPUT" ./cmd/endpoint/endpoint.go
}

if [ "$1" == "--all" ]; then
    mkdir -p $BUILD_DIR
    for arch in "${ALL_ARCHS[@]}"; do
        build_platform "$arch"
    done
elif [ -n "$1" ]; then
    # Simple match check
    found=false
    for arch in "${ALL_ARCHS[@]}"; do
        if [ "$arch" == "$1" ]; then
            found=true
            break
        fi
    done
    
    if [ "$found" = true ]; then
        mkdir -p $BUILD_DIR
        build_platform "$1"
    else
        echo "Skipping/Unsupported: $1"
        exit 0
    fi
else
    usage
fi

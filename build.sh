#!/bin/bash
VERSION=$1

if [ -z "$VERSION" ]; then
        echo "You have to pass build version as arg"
        exit 1
fi

mkdir -p bin

PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

for PLATFORM in "${PLATFORMS[@]}"; do
        GOOS="${PLATFORM%/*}"
        GOARCH="${PLATFORM#*/}"
        OUTPUT="bin/tanker-${GOOS}-${GOARCH}"

        echo "Building ${OUTPUT}..."
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -trimpath -ldflags "-s -w -X github.com/PierreKieffer/http-tanker/pkg/cli.version=$VERSION" -o "$OUTPUT"
done

echo "Done"

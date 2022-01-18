#!/bin/bash
VERSION=$1

if [ -z "$VERSION" ]; then
        echo "You have to pass build version as arg"
        exit 1

fi

go build -ldflags "-X main.version=$VERSION" -o tanker

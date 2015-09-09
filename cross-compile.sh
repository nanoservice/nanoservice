#!/usr/bin/env bash

docker run --rm -it -e CGO_ENABLED=true -e LDFLAGS='-extldflags "-static"' -e COMPRESS_BINARY=true -v $(pwd):/src centurylink/golang-builder-cross

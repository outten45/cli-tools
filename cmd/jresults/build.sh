#!/usr/bin/env bash

BASEDIR=$(dirname "$0")
cd $BASEDIR

GOOS="${GOOS:-linux}"

env GOOS=${GOOS} go build \
    -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse HEAD`" \
    .

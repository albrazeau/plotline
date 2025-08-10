#!/bin/bash

set -euo pipefail
set -x

if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  VERSION=$(git describe --tags)
else
  COUNT=$(git rev-list --count HEAD)
  VERSION="0.0.0+build.$COUNT"
fi

COMMIT=$(git rev-parse --short HEAD)

BUILT=$(date -u +'%Y-%m-%dT%H:%M:%SZ')


docker build \
  --target prod \
  -t plotline-back:$COMMIT \
  --build-arg VERSION=$VERSION \
  --build-arg COMMIT=$COMMIT \
  --build-arg BUILT=$BUILT \
  .

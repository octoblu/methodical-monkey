#!/bin/bash

APP_NAME=methodical-monkey
TMP_DIR="/tmp/$APP_NAME-build"
IMAGE_NAME="local/$APP_NAME"

init() {
  rm -rf $TMP_DIR && \
    mkdir -p $TMP_DIR
}

build() {
  docker build --tag "$IMAGE_NAME:built" .
}

run() {
  docker run --rm \
    --volume "$TMP_DIR:/export/" \
    "$IMAGE_NAME:built" \
      cp "$APP_NAME" /export
}

copy() {
  cp "$TMP_DIR/$APP_NAME" .
  cp "$TMP_DIR/$APP_NAME" entrypoint/
}

package() {
  local tag="$1"
  docker build --tag "$IMAGE_NAME:$tag" entrypoint
}

fatal() {
  local message="$1"
  echo "$message"
  exit 1
}

main() {
  local tag="$1"
  if [ -z "$tag" ]; then
    tag="latest"
  fi
  init    || fatal "init failed"
  build   || fatal "build failed"
  run     || fatal "run failed"
  copy    || fatal "copy failed"
  package "$tag" || fatal "package failed"
}

main "$@"

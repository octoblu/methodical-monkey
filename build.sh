#!/bin/bash

APP_NAME=methodical-monkey
TMP_DIR="/tmp/$APP_NAME-build"
IMAGE_NAME=local/$APP_NAME

init() {
  rm -rf $TMP_DIR/ && \
    mkdir -p $TMP_DIR/
}

build() {
  docker build --tag $IMAGE_NAME:built .
}

run() {
  docker run --rm \
    --volume $TMP_DIR:/export/ \
    $IMAGE_NAME:built \
      cp $APP_NAME /export
}

copy() {
  cp $TMP_DIR/$APP_NAME .
  cp $TMP_DIR/$APP_NAME entrypoint/
}

package() {
  docker build --tag $IMAGE_NAME:latest entrypoint
}

panic() {
  local message=$1
  echo $message
  exit 1
}

main() {
  init    || panic "init failed"
  build   || panic "build failed"
  run     || panic "run failed"
  copy    || panic "copy failed"
  package || panic "package failed"
}

main

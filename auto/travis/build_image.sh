#!/bin/bash

if [[ "${TRAVIS_TAG}" == "" ]]; then
  echo "no travis tag, skipping"
  exit 0
fi

IMAGE_TAG=beacon-api:$TRAVIS_TAG
ARTIFACT_URL=https://github.com/dadleyy/beacon.api/releases/download/$TRAVIS_TAG/beacon-linux-amd64.tar.gz
DOCKERFILE=./auto/docker/Dockerfile
docker build --build-arg ARTIFACT_URL=$ARTIFACT_URL -t $IMAGE_TAG -f $DOCKERFILE .

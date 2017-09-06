#!/bin/bash


if [[ "${TRAVIS_TAG}" == "" ]]; then
  echo "no travis tag, skipping"
  exit 0
fi

IMAGE_TAG=beacon-api:$TRAVIS_TAG
ARTIFACT_URL=https://github.com/dadleyy/beacon.api/releases/download/$TRAVIS_TAG/beacon-linux-amd64.tar.gz
DOCKERFILE=./auto/docker/Dockerfile

echo "building docker image"
docker build --build-arg ARTIFACT_URL=$ARTIFACT_URL -t $IMAGE_TAG -f $DOCKERFILE .

echo "logging into docker registry"
docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

echo "tagging built docker image"
docker tag $IMAGE_TAG $DOCKER_USERNAME/$IMAGE_TAG

echo "pushing new image to registry"
docker push $DOCKER_USERNAME/$IMAGE_TAG

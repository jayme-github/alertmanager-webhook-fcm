#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push "jaymedh/alertmanager-webhook-fcm:${TRAVIS_TAG}"

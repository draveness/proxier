#!/bin/sh

# TODO: use pkg md5 hash instead of git commit sha256
OPERATOR_E2E_IMAGE_TAG=`git rev-parse --short HEAD`
curl --silent -f -lSL https://index.docker.io/v1/repositories/draveness/proxier-e2e/tags/$OPERATOR_E2E_IMAGE_TAG > /dev/null

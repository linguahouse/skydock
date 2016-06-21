#!/usr/bin/env bash

# we need to set dependency to compilare image.
MONGO_IMAGE_NAME=mongo
USERVICE_NAME="skydock"
DOCKER_IMAGE_NAME="skydock"
DOCKER_REPOSITORY="docker.mobibeam.com"


CURRENT_FOLDER="${PWD##*/}"
HOME_FOLDER=~

## include the functions
. $(dirname $0)/s-lib.sh


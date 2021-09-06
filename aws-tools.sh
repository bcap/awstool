#!/bin/bash

set -e -o pipefail 

cd $(dirname $0)

IMAGE="$(docker build -q .)"

docker run --rm -v ~/.aws:/root/.aws -a stdout -a stderr $IMAGE $@
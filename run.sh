#!/bin/bash

set -e -o pipefail -x

cd $(dirname $0)

docker run --rm -v ~/.aws:/root/.aws -a stdout -a stderr aws-tools $@
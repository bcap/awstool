#!/bin/bash

set -e -o pipefail 

cd $(dirname $0)/cmd/awstool

go install -v
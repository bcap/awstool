#!/bin/bash

set -e -o pipefail

cd $(dirname $0)

function log() {
    echo "=> $@" >&2
}

if [[ $(git status -s | wc -l) -gt 0 ]]; then 
    log "repo is dirty, commit/clean changes first"
    exit 1
fi

log "pushing changes to github"
git push 

log "building and pushing docker image"
IMG=bcap/awstool:latest
docker build -t $IMG . && docker push $IMG
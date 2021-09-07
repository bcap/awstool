#!/bin/bash

set -e -o pipefail

cd $(dirname $0)

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" exit

function log() {
    echo "=> $@" >&2
}

if [[ $(git status -s | wc -l) -gt 0 ]]; then 
    log "repo is dirty, commit/clean changes first"
    exit 1
fi

log "building and pushing docker image"
IMG=bcap/awstool:latest
docker build -t $IMG . && docker push $IMG

log "pushing changes to github"
git push 

log "building and creating github release"
(cd cmd/awstool && go build -v -o $TMPDIR/awstool)
gh release create $(date +%Y%m%d-%H%M) $TMPDIR/awstool < /dev/null
#!/bin/bash

set -e -o pipefail

cd $(dirname $0)

TMPDIR=$(mktemp -d)
mkdir $TMPDIR/bin
trap "rm -rf $TMPDIR" exit

function log() {
    echo "=> $@" >&2
}

if [[ $(git status -s | wc -l) -gt 0 ]]; then 
    log "repo is dirty, commit/clean changes first"
    exit 1
fi

log "running tests locally"
go test ./...

# based on https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
(
cd cmd/awstool
cat <<END > $TMPDIR/platforms
darwin  amd64 awstool-macos-amd64
darwin  arm64 awstool-macos-arm64
linux   amd64 awstool-linux-amd64
linux   arm64 awstool-linux-arm64
windows amd64 awstool-windows-amd64.exe
END
cat $TMPDIR/platforms | while read GOOS GOARCH OUT; do
    NAME=awstool-$GOOS-$GOARCH
    log "building binary for $GOOS $GOARCH"
    export GOOS GOARCH
    go build -o $TMPDIR/bin/$OUT
done
)

log "building and pushing docker image"
IMG=bcap/awstool:latest
docker build -t $IMG . && docker push $IMG

log "pushing changes to github"
git push 

log "creating github release"
gh release create $(date +%Y%m%d-%H%M) $TMPDIR/bin/* < /dev/null

log "pulling new tag"
git pull
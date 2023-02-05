.DEFAULT_GOAL := build

# 
# docker targets
#
build:
	docker build -t bcap/awstool:latest . 

shell: build
	docker run --rm -it -v ~/.aws:/root/.aws --entrypoint /bin/bash bcap/awstool:latest

pre-build:
	docker build --target pre-build -t bcap/awstool:pre-build . 
	
shellb: pre-build
	docker run --rm -it -v ~/.aws:/root/.aws --entrypoint /bin/bash bcap/awstool:pre-build

#
# executables for different archs and oses
#
dist: dist-macos-amd64 dist-macos-arm64 dist-linux-amd64 dist-linux-arm64 dist-windows-amd64 dist-windows-arm64

dist-macos-amd64:
	GOOS=darwin  GOARCH=amd64 go build -o dist/awstool-macos-amd64 cmd/awstool/*.go

dist-macos-arm64:
	GOOS=darwin  GOARCH=arm64 go build -o dist/awstool-macos-arm64 cmd/awstool/*.go

dist-linux-amd64:
	GOOS=linux   GOARCH=amd64 go build -o dist/awstool-linux-amd64 cmd/awstool/*.go

dist-linux-arm64:
	GOOS=linux   GOARCH=arm64 go build -o dist/awstool-linux-arm64 cmd/awstool/*.go

dist-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o dist/awstool-windows-amd64.exe cmd/awstool/*.go

dist-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -o dist/awstool-windows-arm64.exe cmd/awstool/*.go
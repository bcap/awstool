#!/bin/bash

docker run --rm -v ~/.aws:/root/.aws -a stdout -a stderr bcap/awstool:latest $@
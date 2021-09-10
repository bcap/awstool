# awstool

This is a tool I have been writing to help me out with common commands regarding AWS resources/APIS

Currently implemented commands are:
- `dump`: generates a single json dumping the results of many different description APIs from AWS
- `ec2 resolve`: resolves/finds ec2 instances by a given set of inputs. Prints a short summary of them with key data like id, tags, ips (public & private)
- `es resolve`: resolves/finds elasticsearch domains by a given set of inputs. Prints a short summary of them
- `es request`: sends requests to an elasticsearch domain

## Setup

The tool rely on having your AWS credentials properly configured. This is normally done while configuring the `aws` cli, which is normally done with:

    aws configure

If your organization uses SSO login, be sure to configure it with:

    aws configure sso

And then login with:

    aws sso login

In case you have multiple aws accounts, be sure to use the `--profile` flag in the above commands so you can configure/login to different accounts. The tool implemented here also supports the `--profile` flag

## Running latest release

Simplest way to run the tool is by using docker. You dont need to clone this repo for that:

    docker run --rm -v ~/.aws:/root/.aws -a stdout -a stderr bcap/awstool help

You can also download pre-built binaries from the [releases page](https://github.com/bcap/awstool/releases/)

## Building / Running / Installing

You can install in your system with the `install.sh` script, which will build the tool and copy it to `GOPATH/bin/awstool` (you can run `go env GOPATH` to figure out your GOPATH)

    ./install.sh
    awstool help

Alternatively you can use the `awstool.sh` script, which wraps building a docker image and running it:

    ./awstool.sh help

You can instead build and/or run directly with go with no system installation or docker images. This works best for a development workflow:

    go run cmd/awstool/*.go help

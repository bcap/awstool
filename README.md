# awstool

    % ./awstool.sh help
    Set of tools to help on aws operations

    Usage:
    awstool [command]

    Available Commands:
    completion  generate the autocompletion script for the specified shell
    dump        generates a single json dumping the results of many different description APIs from AWS
    help        Help about any command
    resolve     resolves ec2 instances by a given set of inputs

    Flags:
    -h, --help                         help for awstool
        --max-requests-in-flight int   How many requests in parallel are allowed to be executed against the AWS APIs at anypoint in time (default 50)
        --max-retries int              Maximum amount of retries we should allow for a particular request before it fails. See also --max-retry-time (default 50)
        --max-retry-time duration      Maximum amount of total time we wait for a request to be retried over and over in case of failures. See also --max-retries (default 10s)
    -p, --profile string               Use this AWS profile. Profiles are configured in ~/.aws/config. If not specified then the default profile will be used
    -q, --quiet                        Runs quiet, excluding even error messages. This overrides --verbosity
    -v, --verbosity count              Controls loggging verbosity. Can be specified multiple times (eg -vv) or a count can be passed in (--verbosity=2). Defaults to print error messages. See also --quiet

    Use "awstool [command] --help" for more information about a command.

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

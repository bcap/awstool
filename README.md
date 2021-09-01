# aws-tools

Tools for dealing with aws apis. Currently only one tool is implemented:

## `dump` 

Generates a single json dumping the results of many different description APIs from AWS. Eg: S3 buckets, instances, beanstalk stacks, opsworks, etc

## Setup

The commands rely on having your AWS credentials properly configured. This is normally done while configuring the `aws` cli, which is normally done with:

    aws configure

If your organization uses SSO login, be sure to configure it with:

    aws configure sso

And then login with:

    aws sso login

In case you have multiple aws accounts, be sure to use the `--profile` flag in the above commands so you can configure/login to different accounts. The tools implemented here also supports the `--profile` flag

## Running

There are a few ways to execute the tools:

Using local go: 

    go run cmd/dump/main.go -h

Using docker directly: 

    docker build -t aws-tools
    docker run -v ~/.aws:/root/.aws -a stdout -a stderr aws-tools dump -h

Using `run.sh`, which wraps building and running through docker:

    ./run.sh dump -h

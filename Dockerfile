FROM golang:alpine as build
RUN ["apk", "add", "build-base"]
COPY . /app
WORKDIR /app
RUN ["go", "test", "./..."]
WORKDIR /app/cmd/aws-tools
RUN ["go", "build"]
RUN ["./aws-tools", "help"]

FROM alpine
COPY --from=build /app/cmd/aws-tools/aws-tools /app/aws-tools
VOLUME /root/.aws
ENTRYPOINT [ "/app/aws-tools" ]
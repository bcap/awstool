FROM golang:alpine as build
COPY . /app
WORKDIR /app/cmd/aws-tools
RUN ["go", "build"]
RUN ["./aws-tools", "help"]

FROM alpine
COPY --from=build /app/cmd/aws-tools/aws-tools /app/aws-tools
VOLUME /root/.aws
ENTRYPOINT [ "/app/aws-tools" ]
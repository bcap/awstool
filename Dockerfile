FROM golang:alpine as build
RUN ["apk", "add", "build-base"]
COPY . /app
WORKDIR /app
RUN ["go", "test", "./..."]
WORKDIR /app/cmd/awstool
RUN ["go", "build"]
RUN ["./awstool", "help"]

FROM alpine
COPY --from=build /app/cmd/awstool/awstool /app/awstool
VOLUME /root/.aws
ENTRYPOINT [ "/app/awstool" ]
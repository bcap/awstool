FROM golang:alpine as build
WORKDIR /app
COPY . .
RUN ["go", "build", "-o", "bin/dump", "cmd/dump/main.go"]

FROM alpine
WORKDIR /app
ENV PATH="$PATH:/app"
COPY --from=build /app/bin/dump dump
VOLUME /root/.aws
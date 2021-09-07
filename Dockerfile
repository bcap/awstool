FROM golang:alpine as build

RUN ["apk", "add", "build-base"]

WORKDIR /app

# To speed up build process we first download required modules
# then later focus on building our source. That way changing a 
# file in the project doesnt cause redownload of dependencies
COPY go.mod go.sum ./
RUN go mod download -x

COPY . .
RUN go test -v ./...
WORKDIR /app/cmd/awstool
RUN go build -v -x
RUN ./awstool help

# finally the released image is minimal with only the compiled binary
FROM alpine
COPY --from=build /app/cmd/awstool/awstool /app/awstool
VOLUME /root/.aws
ENTRYPOINT [ "/app/awstool" ]
# base image for everything else
FROM alpine as base
RUN apk add go bash

# image with app source ready to be built
FROM base as pre-build
RUN apk add build-base
WORKDIR /app
## cache deps
COPY go.mod go.sum ./
RUN go mod download -x
### check cachebuilddeps.go for more context
COPY cachebuilddeps.go .
RUN go build -v ./...
RUN go vet -v ./...
RUN go test -v ./...
## copy everything else
COPY . .
RUN rm cachebuilddeps.go

# build
FROM pre-build as build
RUN go build -v ./...
RUN go vet -v ./...
RUN go test -v ./...
RUN go build -o awstool cmd/awstool/*.go
## check that the program can initialize
RUN ./awstool help

# finally the released image is minimal with only the compiled binary
FROM base
COPY --from=build /app/awstool /app/awstool
VOLUME /root/.aws
ENTRYPOINT [ "/app/awstool" ]
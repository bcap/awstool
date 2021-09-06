FROM golang:alpine as build
WORKDIR /app
COPY . .
RUN ["sh", "-c", "go build -o bin/aws-tools cmd/aws-tools/*.go"]

FROM alpine
COPY --from=build /app/bin/aws-tools /app/aws-tools
VOLUME /root/.aws
ENTRYPOINT [ "/app/aws-tools" ]
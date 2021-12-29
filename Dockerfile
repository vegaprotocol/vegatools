FROM golang:1.17-alpine AS builder
# Install git: required for fetching the Go dependencies.
RUN apk update && apk add --no-cache git
ENV GOPROXY=direct GOSUMDB=off
WORKDIR /go/src/project
ADD . .
RUN env CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o vegatools .

FROM alpine:3.14
RUN apk add --no-cache bash ca-certificates jq
COPY --from=builder /go/src/project/vegatools /usr/local/bin/
ENTRYPOINT ["vegatools"]

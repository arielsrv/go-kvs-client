ARG GO_VERSION=1.26.0
FROM golang:${GO_VERSION} AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=secret,id=netrc,target=/root/.netrc \
    go mod download
COPY . .
ARG COMMIT_HASH
RUN --mount=type=cache,target=/go/pkg/mod \
    go test -v ./... && \
    CGO_ENABLED=0 go build -ldflags="-X 'kvs.CommitHash=${COMMIT_HASH}'" ./...

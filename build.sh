#!/bin/bash

set -e

go get -u github.com/swaggo/swag/cmd/swag
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get github.com/envoyproxy/protoc-gen-validate
go get -u github.com/securego/gosec/v2/cmd/gosec

GOPATH=$(go env GOPATH)
export PATH=$PATH:${GOPATH}/bin

protoc \
  -I . \
  -I ${GOPATH}/src \
  -I ${GOPATH}/src/github.com/envoyproxy/protoc-gen-validate \
  --proto_path="./protobuf" \
  --go_out=":./" \
  --validate_out="lang=go:." \
  meta.proto general.proto settings.proto status.proto messages.proto contacts.proto users.proto backup.proto internal.proto

gosec -exclude G104,G404,G307,G402 ./...

swag init -g server.go -d controller --parseDependency --parseInternal #--parseDepth 1

go build cmd/main.go

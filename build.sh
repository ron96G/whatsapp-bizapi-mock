#!/bin/bash

set -e 

go get -u github.com/swaggo/swag/cmd/swag
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get github.com/envoyproxy/protoc-gen-validate

GOPATH=$(go env GOPATH)
export PATH=$PATH:${GOPATH}/bin

protoc \
  -I . \
  -I ${GOPATH}/src \
  -I ${GOPATH}/src/github.com/envoyproxy/protoc-gen-validate \
  --proto_path="./protobuf" \
  --go_out=":./" \
  --validate_out="lang=go:." \
  whatsapp.proto internal.proto

swag init -g server.go -d controller --parseDependency --parseInternal #--parseDepth 1

go build cmd/main.go

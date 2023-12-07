#!/bin/bash

go clean

GOPROXY="https://goproxy.cn,direct" GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o carizon-device-plugin

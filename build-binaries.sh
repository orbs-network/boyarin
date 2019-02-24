#!/bin/sh -xe

time go build -ldflags '-w -extldflags "-static"' -o boyar.bin -a ./boyar/main/main.go

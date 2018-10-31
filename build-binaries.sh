#!/bin/bash -x

CGO_ENABLED=0 time go build -ldflags '-w -extldflags "-static"' -o strelets.bin -a main.go

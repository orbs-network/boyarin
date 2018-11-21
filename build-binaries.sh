#!/bin/bash -x

CGO_ENABLED=0 time go build -ldflags '-w -extldflags "-static"' -o strelets.bin -a main.go

CGO_ENABLED=0 time go build -ldflags '-w -extldflags "-static"' -o boyar.bin -a ./boyar/main/main.go

CGO_ENABLED=0 time go test -ldflags '-w -extldflags "-static"' -o e2e.test -a -c ./test/e2e

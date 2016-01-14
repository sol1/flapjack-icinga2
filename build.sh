#!/bin/bash

set -e

export GOPATH=$(pwd):$GOPATH

go get github.com/garyburd/redigo/redis
go get gopkg.in/alecthomas/kingpin.v2

go test -v ./...
go build -o bin/flapjack-icinga2 -x github.com/sol1/flapjack_icinga2




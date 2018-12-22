#!/usr/bin/env bash
docker pull gmaurizio/go-ubuntu:1.11.4 || exit 4
docker pull gmaurizio/go-alpine:latest || exit 8

docker run --rm -it -v $GOPATH:/mnt --name goubuntu gmaurizio/go-ubuntu:1.11.4 \
/bin/bash -c 'cd /mnt/src/github.com/gus-maurizio/ && sre-agent/scripts/build.bash '

docker run --rm -it -v $GOPATH:/mnt --name goalpine gmaurizio/go-alpine:latest \
/bin/bash -c 'cd /mnt/src/github.com/gus-maurizio/ && sre-agent/scripts/build.bash '



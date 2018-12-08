#!/usr/bin/env bash
plugdir=${1:-plugins}
plugos=${2:-darwin}
plugarch=${3:-amd64}
export GOOS=$plugos
export GOARCH=$plugarch
echo go build -o $GOOS/$GOARCH/sreagent github.com/gus-maurizio/sre-agent
go build -o $GOOS/$GOARCH/sreagent github.com/gus-maurizio/sre-agent
for i in $(find ${plugdir} -type f -name 'plugin_*.go')
do
    echo compiling $i
    object=$(echo $i | sed 's/.go/.so/')
    echo go build -buildmode=plugin -o $GOOS/$GOARCH/$object $i
    CGO_ENABLED=0 go build -buildmode=plugin -o $GOOS/$GOARCH/$object $i
done
find $plugos -type f | xargs file

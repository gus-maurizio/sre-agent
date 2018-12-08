#!/usr/bin/env bash
plugdir=${1:-plugins}
plugos=${2:-darwin}
plugarch=${3:-amd64}
for i in $(find ${plugdir} -type f -name 'plugin_*.go')
do
    echo compiling $i
    object=$(echo $i | sed 's/.go/.so/')
    GOOS=$plugos
    GOARCH=$plugarch
    echo go build -buildmode=plugin -o $GOOS/$GOARCH/$object $i
    go build -buildmode=plugin -o $GOOS/$GOARCH/$object $i
done

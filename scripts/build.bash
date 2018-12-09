#!/usr/bin/env bash
plugdir=${1:-plugins}
plugos=${2:-darwin}
plugarch=${3:-amd64}
echo go build -o $plugos/$plugarch/sreagent github.com/gus-maurizio/sre-agent
go build -o $plugos/$plugarch/sreagent github.com/gus-maurizio/sre-agent
for i in $(find ${plugdir} -type f -name 'plugin_*.go')
do
    echo compiling $i
    object=$(echo $i | sed 's/.go/.so/')
    echo go build -buildmode=plugin -o $plugos/$plugarch/$object $i
    go build -buildmode=plugin -o $plugos/$plugarch/$object $i
done
find $plugos -type f | xargs file
find $plugos -type f | xargs -I {} ldd {}

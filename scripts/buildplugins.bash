#!/usr/bin/env bash
plugdir=${1:-plugins}
for i in $(find ${plugdir} -type f -name 'plugin_*.go'); do echo compiling $i; object=$(echo $i | sed 's/.go/.so/');echo go build -buildmode=plugin -o $object $i;go build -buildmode=plugin -o $object $i;done

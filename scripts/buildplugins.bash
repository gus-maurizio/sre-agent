#!/usr/bin/env bash
for i in $(find plugins/*.go); do echo compiling $i; object=$(echo $i | sed 's/.go/.so/');echo go build -buildmode=plugin -o $object $i;go build -buildmode=plugin -o $object $i;done

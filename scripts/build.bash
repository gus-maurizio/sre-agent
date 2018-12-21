#!/usr/bin/env bash
go build -buildmode=plugin -o $GOPATH/src/github.com/gus-maurizio/plugin_mem/$(uname -s)/plugin_mem.so $GOPATH/src/github.com/gus-maurizio/plugin_mem/plugin_mem.go
repo=github.com/gus-maurizio
pack="plugin_mem plugin_cpu plugin_disk plugin_load"
packages=${1:-$pack}
mod="sre-agent"
mainmod=${2:-$mod}

OSTYPE=$(uname -s)
[ "$OSTYPE" == "Linux"  ] && [ "$(cat /etc/os-release|grep ^ID=)" == "ID=alpine" ] && OSTYPE=Alpine

echo Building for $OSTYPE

for i in $packages
do 
  echo Building $i 
  echo go build -buildmode=plugin -o $GOPATH/src/$repo/$i/$OSTYPE/$i.so $GOPATH/src/$repo/$i/$i.go 
  go build -buildmode=plugin -o $GOPATH/src/$repo/$i/$OSTYPE/$i.so $GOPATH/src/$repo/$i/$i.go 
done
for i in $mainmod
do
  echo Building $i
  echo go build -o $GOPATH/src/$repo/$i/$OSTYPE/$i $GOPATH/src/$repo/$i
  go build -o $GOPATH/src/$repo/$i/$OSTYPE/$i $GOPATH/src/$repo/$i
done
find $GOPATH/src/github.com/gus-maurizio/*/$OSTYPE -type f|xargs file
[ "$(uname -s)" == "Darwin" ] && find $GOPATH/src/github.com/gus-maurizio/*/$OSTYPE -type f | xargs -I {} otool -L {}
[ "$(uname -s)" == "Linux"  ] && find $GOPATH/src/github.com/gus-maurizio/*/$OSTYPE -type f | xargs -I {} ldd {}

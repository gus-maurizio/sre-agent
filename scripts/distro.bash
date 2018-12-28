#!/usr/bin/env bash
repo=github.com/gus-maurizio
pack="plugin_mem plugin_cpu plugin_disk plugin_filesystem plugin_load plugin_net plugin_connections plugin_ibmmq"
packages=${1:-$pack}
mod="sre-agent"
mainmod=${2:-$mod}

OSTYPE=$(uname -s)
MOSTYPE=$(uname -s)
[ "$MOSTYPE" == "Linux"  ] && [ "$(cat /etc/os-release|grep ^ID=)" == "ID=alpine" ] && MOSTYPE=Alpine

echo Building for $OSTYPE

for i in $mainmod
do
  echo Building $i for $OSTYPE and OS $MOSTYPE 
  echo go build -o $GOPATH/src/$repo/$i/distro/$MOSTYPE/bin/$i $GOPATH/src/$repo/$i
  go build -o $GOPATH/src/$repo/$i/distro/$MOSTYPE/bin/$i $GOPATH/src/$repo/$i
  file $GOPATH/src/$repo/$i/distro/$MOSTYPE/bin/$i
done
[ "$(uname -s)" == "Linux" ] && [ "$MOSTYPE" == "Alpine" ] && echo "Done with $MOSTYPE" && exit 0

for i in $packages
do 
  echo Building $i for $OSTYPE and OS $MOSTYPE 
  echo go build -buildmode=plugin -o $GOPATH/src/$repo/$mainmod/distro/$OSTYPE/lib/$i.so $GOPATH/src/$repo/$i/$i.go 
  go build -buildmode=plugin -o $GOPATH/src/$repo/$mainmod/distro/$OSTYPE/lib/$i.so $GOPATH/src/$repo/$i/$i.go 
  file $GOPATH/src/$repo/$mainmod/distro/$OSTYPE/lib/$i.so
done
[ "$(uname -s)" == "Darwin" ] && find $GOPATH/src/github.com/gus-maurizio/$mainmod/distro  -type f | xargs -I {} otool -L {}
[ "$(uname -s)" == "Linux"  ] && find $GOPATH/src/github.com/gus-maurizio/$mainmod/distro  -type f | xargs -I {} ldd {}

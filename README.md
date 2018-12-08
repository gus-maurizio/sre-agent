# sre-agent
## Compiling this package for multiple platforms
In order to compile the agent, is it good to retrieve from Github along with all the dependencies.
The Go programming language comes with a rich toolchain that makes obtaining packages and building executables incredibly easy.
One of Go's most powerful features is the ability to cross-build executables for any Go-supported foreign platform.
This makes testing and package distribution much easier, because you don't need to have access to a specific platform in order to distribute your package for it.

In this particular case, our `sre-agent` (for monitoring compute environments with extensibility via plugins) present interesting opportunities for building.

I will assume your working laptop is running Mac OS X (Mojave as of this version), with GoLang (version go1.11.2 as of this writing) installed. Docker (18.09 CE) is useful for some aspects and testing to support the cross-platform versions.

### Docker images needed
Download the docker images for the supported operating systems:
- Ubuntu
  - 18.04 (bionic)
  - 18.10 (cosmic)
  - 19.04 (disco)
  - 14.04 (trusty)
  - 16.04 (xenial)
 - Alpine
  - 3.6
  - 3.7
  - 3.8
  - CentOS
   - centos:7.6.1810
   - centos:7.5.1804
   - centos:6.10

Verify all are properly loaded:
```
docker images|grep -e REPO -e ^ubuntu -e ^alpine -e ^centos
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
alpine              3.6                 94627dfbdf19        2 months ago        4.03MB
alpine              3.7                 34ea7509dcad        2 months ago        4.21MB
alpine              3.8                 196d12cf6ab1        2 months ago        4.41MB
centos              6.10                30e66b619e9f        8 weeks ago         194MB
centos              7.5.1804            76d6bc25b8a5        8 weeks ago         200MB
centos              7.6.1810            d5f224905a66        2 days ago          202MB
ubuntu              14.04               f17b6a61de28        2 weeks ago         188MB
ubuntu              16.04               a51debf7e1eb        2 weeks ago         116MB
ubuntu              18.04               93fd78260bd1        2 weeks ago         86.2MB
ubuntu              18.10               0bfd76efee03        2 weeks ago         73.7MB
ubuntu              19.04               d861a21f6090        2 weeks ago         74.9MB
```
### Installing from Version Control
Before we can create executables from a Go package, we have to obtain its source code. The go get tool can fetch packages from version control systems like GitHub.
Under the hood, go get clones packages into subdirectories of the $GOPATH/src/ directory.
Then, if applicable, it installs the package by building its executable and placing it in the $GOPATH/bin directory.
If you configured Go as described in the prerequisite tutorials, the $GOPATH/bin directory is included in your $PATH environmental variable, which ensures that you can use installed packages from anywhere on your system.
We will use your home directory for this instructions.

It's common to use go get with the -u flag, which instructs go get to obtain the package and its dependencies, or update those dependencies if they're already present on the machine.
This command can take a little to finish, since it needs to validate and download all the dependencies.
No output actually indicates that the command executed successfully.

```
mkdir -p ~/go/src
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
go get -u github.com/gus-maurizio/sre-agent
```
The command also builds the executable and places it in the $GOPATH/bin.

In order to build the plugins (either those provided with the distribution, or the ones you can create to extend the agent), a special script has been included in the `scripts` directory.
Executing the following command will compile **just for the architecture of your build system**, in this case Mac OS X. The last parameter to the `buildplugins.bash` script is the directory where the plugin source code resides. The plugin source **must start with plugin_**. The plugins can reside (*.so) in any directory. This can be specified in the YAML configuration file for the agent.
```
bash scripts/buildplugins.bash plugins
compiling plugins/plugin_cpuram.go
go build -buildmode=plugin -o plugins/plugin_cpuram.so plugins/plugin_cpuram.go
compiling plugins/plugin_network.go
go build -buildmode=plugin -o plugins/plugin_network.so plugins/plugin_network.go
compiling plugins/plugin_system.go
go build -buildmode=plugin -o plugins/plugin_system.so plugins/plugin_system.go
```

#### Rebuild the executable
To specify a different name or location for the executable, use the -o flag. Let's build an executable called sreagent
and place it in a build directory (will be created if it does not exist) within the current working directory:
```
$ GOOS=`echo "$(uname -s)"| tr '[:upper:]' '[:lower:]'`
$ GOARCH=amd64
$ go build -o $GOOS/$GOARCH/sreagent github.com/gus-maurizio/sre-agent
$ file $GOOS/$GOARCH/sreagent
darwin/amd64/sreagent: Mach-O 64-bit executable x86_64
$ otool -L  $GOOS/$GOARCH/sreagent
darwin/amd64/sreagent:
	/usr/lib/libobjc.A.dylib (compatibility version 1.0.0, current version 228.0.0)
	/System/Library/Frameworks/Foundation.framework/Versions/C/Foundation (compatibility version 300.0.0, current version 1560.12.0)
	/System/Library/Frameworks/IOKit.framework/Versions/A/IOKit (compatibility version 1.0.0, current version 275.0.0)
	/usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1252.200.5)
	/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation (compatibility version 150.0.0, current version 1560.12.0)
	/System/Library/Frameworks/Security.framework/Versions/A/Security (compatibility version 1.0.0, current version 58286.220.15)
```

### Building for a different OS and architecture
The go build command lets you build an executable file for any Go-supported target platform, on your platform.
This means you can test, release and distribute your application without building those executables on the target platforms you wish to use.

Cross-compiling works by setting required environment variables that specify the target operating system and architecture.
We use the variable GOOS for the target operating system, and GOARCH for the target architecture.
To build an executable, the command would take this form:
```
GOOS=target-OS GOARCH=target-architecture;go build -o $GOOS/$GOARCH/sreagent github.com/gus-maurizio/sre-agent
```
In our case we will build for the `linux` OS and `amd64` architecture:
```
GOOS=linux GOARCH=amd64;go build -o $GOOS/$GOARCH/sreagent github.com/gus-maurizio/sre-agent
```
Unfortunately this will only work in Mac OS X if you **do not have the need for CGO**.
If your plugin requires C code (like we indeed do), this will not work. Fortunately there is a solution!

### Using Docker to cross compile (and test!)
Download the docker image to compile: `golang:1.11.2-alpine3.8` and verify:
```
$ docker images|grep -e REPO -e ^golang
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
golang              1.11.2-alpine3.8    57915f96905a        5 weeks ago         310MB

$ docker run --rm -it -v $GOPATH:/tmp --name goalpine golang:1.11.2-alpine3.8 /bin/sh
/go # apk update
/go # apk add git gcc musl-dev bash file libc-dev
/go # go get -u github.com/gus-maurizio/sre-agent
/go # cd src/github.com/gus-maurizio/sre-agent
/go # find linux/ darwin/ -type f |xargs rm
/go # bash scripts/buildplugins.bash plugins linux amd64
go build -o linux/amd64/sreagent github.com/gus-maurizio/sre-agent
compiling plugins/plugin_cpuram.go
go build -buildmode=plugin -o linux/amd64/plugins/plugin_cpuram.so plugins/plugin_cpuram.go
# command-line-arguments
loadinternal: cannot find runtime/cgo
compiling plugins/plugin_network.go
go build -buildmode=plugin -o linux/amd64/plugins/plugin_network.so plugins/plugin_network.go
# command-line-arguments
loadinternal: cannot find runtime/cgo
compiling plugins/plugin_system.go
go build -buildmode=plugin -o linux/amd64/plugins/plugin_system.so plugins/plugin_system.go
# command-line-arguments
loadinternal: cannot find runtime/cgo
linux/amd64/plugins/plugin_network.so: ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, not stripped
linux/amd64/plugins/plugin_system.so:  ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, not stripped
linux/amd64/plugins/plugin_cpuram.so:  ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, not stripped
linux/amd64/sreagent:                  ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib/ld-musl-x86_64.so.1, not stripped
/go #
/go #


/go # cd $GOPATH/src/github.com/gus-maurizio/sre-agent
/go/src/github.com/gus-maurizio/sre-agent #

```

## Monitoring Memory Usage
The package net/http/pprof can be used.
```
go tool pprof http://localhost:9000/debug/pprof/heap
go tool pprof http://localhost:9000/debug/pprof/profile?seconds=30
go tool pprof http://localhost:9000/debug/pprof/block
wget http://localhost:9000/debug/pprof/trace?seconds=5
go tool pprof http://localhost:9000/debug/pprof/mutex
```
To view all available profiles, open http://localhost:9000/debug/pprof/ in your browser.
### Using Prometheus to monitor
The simplest way is to use the container version.

```
mkdir ~/prometheus-data

docker run --rm -it -p 9090:9090 --name prom -v ~/prometheus-data:/prometheus \
       prom/prometheus  --config.file=/prometheus/prometheus.yml

docker exec -ti prom /bin/sh
```
Where the configuration file will perform the collection as defined below:
```
global:
  scrape_interval:     5s
  evaluation_interval: 5s

scrape_configs:
  - job_name: 'sre-agent'
    metrics_path: /metrics
    scheme: http
    static_configs:
      - targets:
        - docker.for.mac.host.internal:9000

```
### Grafana for Dashboards
You can run Grafana in a container!
```
mkdir ~/grafana-data

#           -v ~/grafana-data/grafana.ini:/etc/grafana/grafana.ini \
docker run --rm -it -p 3000:3000 \
           --name grafana \
           -v ~/grafana-data:/var/lib/grafana \
            grafana/grafana
```
Add the Prometheus data source and load the dashboard: https://grafana.com/dashboards/6671.
Make sure you *DO NOT USE* localhost if you are running in Mac, use *docker.for.mac.host.internal* as the hostname (assuming you are running native the sre-agent). The Data Source for Prometheus must be defined the same way!.

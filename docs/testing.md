# Testing SRE-AGENT with Docker
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
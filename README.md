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

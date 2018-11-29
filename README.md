# sre-agent
Agent for monitoring compute environments with extensibility via plugins
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

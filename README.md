# nbminer_prometheus_exporter

This is a simple application which will sit in front of your NBMiner and export the statistics from the web endpoint
in a prometheus parsable format. This allows you to have your NBMiner stats in Prometheus and thus likely grafana.

## Usage

```bash
go get github.com/gaffo/nbminer_prometheus_exporter
nbminer_prometheus_exporter
```

## Arguments
```
Usage
  -host string
        host and port to bind to for which prometheus polls (default ":2112")
  -minter string
        the host and port where nbminer is exporting, {VALUE}/api/v1/status (default "http://localhost:22333")
  -polling_interval int
        The number of seconds to sleep between polling invervals (default 30)
```
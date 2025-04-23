# System OS Info Exporter

This exporter collects and exposes system information and package details as Prometheus metrics.

## Features

- Collects OS information (name, version, architecture, platform, kernel version).
- Collects installed package versions.
- Indicates if updates are available for installed packages.

## Usage

### Build the Exporter

```bash
go mod init system_os_info
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go build -o system_os_info system_os_info.go
```

### Run the Exporter

```bash
./system_os_info --address=0.0.0.0 --port=9101 --interval=30
```

- Default address: `0.0.0.0`
- Default port: `9101`
- Default interval: `30` minutes

### Metrics

- Visit `/metrics` to view the exported metrics.

### Example

```bash
curl http://localhost:9101/metrics
```
# System OS Info Exporter

This exporter collects and exposes system information and package details as Prometheus metrics.

## Features

- Collects OS information (name, version, architecture, platform, kernel version).
- Collects installed package versions.
- Indicates if updates are available for installed packages.
- Allows restricting CPU usage (in millicores) and memory usage (in MB).

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
./system_os_info --address=0.0.0.0 --port=9101 --interval=30 --resource.cpu=500 --resource.memory=512
```

### Flags

| Flag               | Default Value | Description                                                                 |
|--------------------|---------------|-----------------------------------------------------------------------------|
| `--address`        | `0.0.0.0`     | Address to bind the HTTP server.                                           |
| `--port`           | `9101`        | Port to bind the HTTP server.                                              |
| `--interval`       | `30`          | Interval (in minutes) to collect metrics.                                  |
| `--resource.cpu`   | `0`           | Maximum CPU usage in millicores (0 for no limit).                          |
| `--resource.memory`| `0`           | Maximum memory usage in MB (0 for no limit).                               |

### Metrics

- Visit `/metrics` to view the exported metrics.

### Example

```bash
curl http://localhost:9101/metrics
```
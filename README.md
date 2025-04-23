# System OS Info Exporter

This exporter collects and exposes system information and package details as Prometheus metrics.

## Features

- Collects OS information (name, version, architecture, platform, kernel version).
- Collects installed package versions.
- Indicates if updates are available for installed packages.
- Allows restricting CPU usage (in millicores) and memory usage (in MB).
- Optionally collects filesystem and process metrics.

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
| `--filesystem`     | `false`       | Enable collection of filesystem metrics. Disabled by default.              |
| `--process`        | `false`       | Enable collection of process metrics. Disabled by default.                 |
| `--debug`          | `false`       | Enable debug mode with detailed logs. Disabled by default.                 |

### Metrics

- **Default Enabled Metrics**:
  - OS information (`system_os_info`)
  - Installed package versions
  - Package update availability

- **Optional Metrics**:
  - Filesystem metrics (`system_filesystem_info`) - Enable with `--filesystem`.
  - Process metrics (`system_process_info`) - Enable with `--process`.

### Example

Run the exporter with filesystem and process metrics enabled in debug mode:
```bash
./system_os_info --address=0.0.0.0 --port=9101 --interval=30 --resource.cpu=500 --resource.memory=512 --filesystem --process --debug
```

Visit `http://<address>:<port>/metrics` to view the metrics.
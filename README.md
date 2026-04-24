# go-misc-exporter

A modular Prometheus exporter for collecting metrics from various sources: hardware monitoring, disk SMART data, ZFS, systemd, NUT, Aria2, and Ryzen CPU power stats.

## Collectors

| Collector | Build Tag | Description |
|-----------|-----------|-------------|
| `hwmon`   | `hwmon`   | Hardware monitoring (CPU temperature, fan speed, voltage, current) via lm-sensors |
| `smart`   | `smart`   | Disk health metrics from S.M.A.R.T. data (SATA/NVMe/SCSI) |
| `zfs`     | `zfs`     | ZFS filesystem metrics |
| `systemd` | `systemd` | systemd service state and properties |
| `nut`     | `nut`     | Network UPS Tools metrics |
| `aria2`   | `aria2`   | Aria2 download manager metrics |
| `ryzenadj`| `ryzenadj`| AMD Ryzen CPU power/temperature via ryzenadj (requires CGO) |

Each collector is compiled into the binary as a build tag, allowing you to customize which modules are included.

## Configuration

Configuration is loaded from a JSON file. The default path is `/etc/gme/conf.json`. Specify a custom path with `-c <file>`.

### Global Settings

The `exporter` section configures the HTTP server:

- `listen` (optional): The address to listen on, e.g. `:4403`.
- `path` (optional): The metrics endpoint path for the default prometheus registry. By default each collector registers under its own path (see per-collector).
- `log.level` (optional): Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). Default: `INFO`.
- `log.file` (optional): Log file path. Empty = stdout. Auto-set to journald when running under systemd.

### Per-Collector Configuration

Each collector supports these shared fields:

- `path` (optional): Prometheus metrics scrape path for this collector. Default: `/metrics` (shared).
  Separate paths let you expose different collectors on distinct endpoints (e.g. `/metrics/hwmon`, `/metrics/smart`).

---

#### hwmon

Hardware monitoring via lm-sensors (CPU temperature, fan speed, voltage, current).

```json
{
  "hwmon": {}
}
```

No options needed — discovers all sensors automatically. Requires `lmsensors` installed and configured on the host.

---

#### smart

Disk health metrics from S.M.A.R.T. data (SATA/NVMe/SCSI). Blacklists `loop`, `zram`, `zd`, `sr` devices by default.

```json
{
  "smart": {
    "skip": ["/dev/disk/by-id/some-disk", "loop0"]
  }
}
```

- `skip`: List of disk names or device paths to exclude. Absolute paths are resolved to basenames for matching.

---

#### zfs

ZFS filesystem metrics via `/dev/zfs` kernel interface.

```json
{
  "zfs": {}
}
```

- `skip`: List of ZFS pool names to exclude from scraping.
- `device` (optional): Path to the `/dev/zfs` control character device provided by the Linux kernel. Default: `/dev/zfs`. Usually no configuration needed.

---

#### systemd

systemd service state and properties.

```json
{
  "systemd": {
    "units": ["nginx.service"],
    "patterns": ["prometheus-*"],
    "states": ["failed", "activating"],
    "properties": ["LoadState", "ActiveState", "SubState", "ExecMainPID"],
    "timeout": 5.0
  }
}
```

- `units`: Exact unit names to monitor.
- `patterns`: Glob patterns for unit matching (e.g. `"prometheus-*"`).
- `states`: Filter units by state (`loaded`, `active`, `failed`, etc.). If empty, all units matching other filters are scraped.
- `properties`: List of systemd properties to collect as metrics. Default: common status properties.
- `timeout` (optional): Timeout in seconds for querying systemd via D-Bus. Default: 5s.

---

#### nut

Network UPS Tools metrics. Supports multiple NUT server connections.

```json
{
  "nut": {
    "servers": [
      {
        "host": "127.0.0.1",
        "port": 3493,
        "username": "admin",
        "password": "secret",
        "timeout": 5.0
      }
    ]
  }
}
```

- `servers`: Array of NUT server connections. If empty, connects to localhost:3493 with no auth.
  - `host` / `port`: NUT server address. Default: `localhost:3493`.
  - `username` / `password`: Authentication credentials.
  - `timeout` (optional): Connection timeout in seconds.

---

#### aria2

Aria2 download manager metrics via RPC.

```json
{
  "aria2": {
    "servers": [
      {
        "rpc": "http://127.0.0.1:6800/rpc",
        "secret": "my-token",
        "timeout": 5.0
      }
    ]
  }
}
```

- `servers`: Array of Aria2 RPC endpoints. If empty, connects to `http://127.0.0.1:6800/rpc` with no secret.
  - `rpc`: Aria2 JSON-RPC URL (must be HTTP or HTTPS).
  - `secret` (optional): Aria2 RPC secret token (auth via `--rpc-secret`).
  - `timeout` (optional): RPC request timeout in seconds.

---

#### ryzenadj

AMD Ryzen CPU power/temperature metrics via the ryzenadj tool (requires CGO, AMD CPU).

```json
{
  "ryzenadj": {}
}
```

No options needed — reads current power and temperature from sysfs. Requires `CGO_ENABLED=1` during build and an AMD Ryzen processor with the `amd_pstate` or `acpi_ppp` driver exposing ryzenadj-compatible interfaces.

---

Use `-h` to see which modules are built into your binary.

## Building

```bash
# Build all collectors (requires CGO for ryzenadj)
./build.sh

# Build specific collectors only
./build.sh hwmon smart zfs

# Manual build with build tags
CGO_ENABLED=0 go build -tags "hwmon,smart,zfs" -o /opt/gme/go-misc-exporter ./main
```

`build.sh` sets `GOAMD64=v3` for optimized AMD64 binaries. The output binary is placed at `/opt/gme/go-misc-exporter`.

For collectors that require CGO (`ryzenadj`), set `CGO_ENABLED=1` before building.

## Installation (systemd)

```bash
# Install systemd service and default config file
./go-misc-exporter -install

# Then enable and start the service
systemctl daemon-reload
systemctl enable --now go-misc-exporter.service   # or: gme.service
```

On Linux, the service is installed to `/etc/systemd/system/go-misc-exporter.service` with an alias `gme.service`. It auto-restarts on failure and sets `RestartSec=5s`.

## Runtime Flags

| Flag | Description |
|------|-------------|
| `-c <file>` | Path to the config file (default: `/etc/gme/conf.json`) |
| `-h` | Show help (list built modules) |
| `-install` | Install systemd service and default config file |

## License

MIT

# driftwatch

Lightweight daemon that detects configuration drift between deployed services and their source manifests.

## Installation

```bash
go install github.com/driftwatch/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/driftwatch/driftwatch.git && cd driftwatch && go build ./...
```

## Usage

Start the daemon pointed at your manifests directory:

```bash
driftwatch --manifests ./deploy/manifests --interval 60s
```

Example configuration file (`driftwatch.yaml`):

```yaml
manifests_dir: ./deploy/manifests
interval: 60s
notify:
  slack_webhook: https://hooks.slack.com/services/your/webhook/url
log_level: info
```

Run with a config file:

```bash
driftwatch --config driftwatch.yaml
```

When drift is detected, driftwatch logs a diff and optionally sends an alert:

```
[DRIFT] service/api-server — field "image" changed: v1.2.3 → v1.2.4
[DRIFT] service/worker    — replicas changed: 3 → 1
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--manifests` | `./manifests` | Path to source manifest directory |
| `--interval` | `30s` | How often to poll for drift |
| `--config` | `` | Path to config file |
| `--dry-run` | `false` | Log drift without sending alerts |

## License

MIT © driftwatch contributors
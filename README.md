# portwatch

A lightweight CLI daemon that monitors port usage and alerts on unexpected bindings or conflicts.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

Start the daemon to monitor all active ports:

```bash
portwatch start
```

Watch specific ports and get alerted on unexpected bindings:

```bash
portwatch watch --ports 8080,5432,6379 --interval 5s
```

Run a one-time snapshot of current port bindings:

```bash
portwatch scan
```

### Example Output

```
[INFO]  Watching ports: 8080, 5432, 6379
[ALERT] Unexpected binding detected: 0.0.0.0:8081 (PID 3921 — node)
[WARN]  Port conflict on :5432 — multiple listeners detected
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ports` | all | Comma-separated list of ports to watch |
| `--interval` | `10s` | Polling interval |
| `--log` | stdout | Path to log file |
| `--daemon` | false | Run as background daemon |

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)
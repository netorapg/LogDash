# LogDash

Point it at a directory and find out what happened. LogDash discovers log files, detects their format, and shows levels, a timeline, the most frequent messages and anomalies, in the terminal or in the browser.

One binary, no configuration, no services to run.

<!-- Add a screenshot of the TUI and one of the web dashboard here.
<p align="center"><img src="docs/tui.png" width="720"/></p>
-->

## Why

Real systems scatter logs across folders and formats. Investigating an incident usually means opening dozens of files, repeating `grep`, and rebuilding the sequence of events in your head, while warnings stay buried in the noise. LogDash gives you the aggregated picture first, so you know where to look.

## Install

Requires Go 1.21 or newer.

```bash
git clone https://github.com/netorapg/LogDash.git
cd LogDash
go build -o logdash ./cmd/logdash
```

## Usage

LogDash has three interfaces over the same analysis.

```bash
logdash analyze -path /var/log          # report printed to the terminal
logdash tui     -path /var/log          # interactive terminal UI
logdash web     -path /var/log          # dashboard at http://localhost:8080
```

### Common options

| Flag | Meaning | Default |
|---|---|---|
| `-path` | Root directory to search | current directory |
| `-last` | Only entries from the last duration (`30m`, `24h`, `7d`) | all |
| `-patterns` | File patterns, comma-separated | `*.log,*.txt` |
| `-depth` | Maximum directory depth | `10` |
| `-max-files` | Maximum number of files to process | no limit |
| `-top` | Number of top messages to show (`analyze`) | `10` |
| `-no-anomalies` | Disable anomaly detection | off |
| `-port` | Port for the web dashboard (`web`) | `8080` |

### Examples

```bash
# Quick health check of the last hour
logdash analyze -path /var/log -last 1h -max-files 5

# Deeper investigation
logdash analyze -path /app/logs -last 24h -verbose -top 20

# Browse interactively
logdash tui -path ./logs -last 7d
```

### TUI keys

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Next / previous view |
| `↑` `↓` or `k` `j` | Scroll |
| `PgUp` `PgDn` | Page up / down |
| `g` / `G` | Top / bottom |
| `r` | Reload |
| `?` | Help |
| `q` | Quit |

## What it does

- **Discovery:** recursive search with pattern filters, depth control and automatic exclusion of directories such as `.git` and `node_modules`.
- **Parsing:** each file is matched against the available parsers and the one with the highest confidence score wins. Python `logging` and generic timestamped logs are supported today, with a fallback for unknown formats.
- **Aggregation:** counts per level, a timeline in configurable buckets, and the most frequent messages.
- **Anomaly detection:** error bursts, spikes in the timeline, and repeated errors.

## Architecture

```
        CLI  ·  TUI  ·  Web
               │
        Application service
               │
   ┌───────────┼───────────┐
Discovery    Parser    Aggregator
```

The three interfaces share one service, which runs the pipeline Discovery → Parser → Aggregator. Each core package sits behind an interface, so parsers and outputs can be added without touching the rest. Details in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

```
cmd/logdash/         # entry point and command parsing
internal/core/       # discovery, parser, aggregator
internal/service/    # pipeline orchestration
internal/tui/        # terminal UI (Bubble Tea)
internal/web/        # HTTP server and dashboard template
```

## Development

The core was written test-first.

```bash
go test ./...
go test -cover ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Roadmap

- [x] Discovery, generic and Python logging parsers, aggregation, anomaly detection
- [x] Command-line report
- [x] Terminal UI
- [x] Web dashboard
- [ ] More parsers: JSON lines, syslog, nginx, Log4j
- [ ] Export to JSON / CSV
- [ ] Watch mode
- [ ] Index cache for large directories

## License

[MIT](LICENSE)

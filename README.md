# 📦 Filtrate Backups
Filtrate Backups is a Go utility for filtering SQL dump archives.
It unpacks `.tar.gz` backups, removes unwanted `INSERT` data for selected tables, and repacks a cleaned dump.

## 🚀 Features
- Streams dump files line-by-line.
- Handles very large SQL lines with configurable memory limits (`MAX_LINE_BYTES`).
- Runs once or as an internal scheduler (`MODE=schedule`, `SCHEDULE_EVERY=...`).
- Supports deployment as:
  - a containerized scheduler,
  - a scheduler near a dedicated S3 service (e.g. MinIO),
  - a system scheduler via `systemd` timer.
- Supports multiple configuration formats through strategy-based loading:
  - `.yaml/.yml`
  - `.toml`
  - `.json`
  - `.conf/.ini`
- Supports combined configuration sources (file + env + CLI overrides).
- Current runtime I/O works with local filesystem paths; S3 object I/O is planned as a separate task.

## ⚙️ Configuration sources (strategy)
The app uses layered config with strategy selection:
1. defaults,
2. config file (`--config` + `--config-format`),
3. environment variables,
4. CLI flags (highest priority).

CLI switches for strategy:
- `--config ./examples/config.yaml`
- `--config-format auto|yaml|toml|json|conf`
- `--config-strategy merge|file-only|env-only`

Default strategy is `merge`.

### Environment variables
```env
DUMPFILE="./data/source.tar.gz"
OUTPUT_FILE="./output/filtered_result.tar.gz"
TABLE_MAP="^tmp_:^log_"
TMP_DIR="./tmp"
MAX_LINE_BYTES=8388608
MODE="once"
SCHEDULE_EVERY="1h"
```

### Combined configuration example
Keep operational logic in YAML, and secrets/urgent overrides in env:

```bash
MODE=once TABLE_MAP='^tmp_:^audit_' go run . --config ./examples/config.yaml
```

## 🗂️ CLI usage
```bash
go run . --input ./dump.tar.gz --output ./output/filtered_result.tar.gz --skip '^tmp_:^log_'
```

Useful flags:
- `--mode once|schedule`
- `--every 30m`
- `--max-line-bytes 16777216`


## 🧪 Large dump generator (1GB)
For stress testing in near-real conditions, use the built-in generator:

```bash
go run ./cmd/dumpgen --output ./data/generated_dump_1gb.tar.gz --target-size 1GB --tables users,orders,events
```

This creates a `.tar.gz` archive with `dump.sql` inside, containing multiple `CREATE TABLE` and batched `INSERT` statements until the SQL payload reaches the target size. Payload values are random alphanumeric strings (not constant `x`).

Useful tuning flags:
- `--tables` (comma-separated list, e.g. `users,orders,events`)
- `--table` (legacy alias for single table)
- `--rows-per-insert` (default `1000`)
- `--value-size` (default `128`)
- `--seed` (optional random seed for reproducible data)
- `--target-size` supports `MB/GB` and `MiB/GiB`

## 🐳 Run in Docker (scheduler + standalone S3 service)
1. Fill `.env`.
2. Start stack:

```bash
docker compose up --build -d
```

`docker-compose.yml` starts:
- `cleaner` (scheduler in container)
- `minio` (separate S3-compatible service for backup infrastructure)

## 🖥️ Run as system scheduler
Systemd units are provided in `deploy/systemd/`:
- `mysql-dump-cleaner.service`
- `mysql-dump-cleaner.timer`

Example:
```bash
sudo cp deploy/systemd/mysql-dump-cleaner.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now mysql-dump-cleaner.timer
```

## 🔭 Roadmap
✅ Current: Basic filtering for MySQL-compatible dumps.

### 🛠️ Planned / progress
- ✅ Make a smaller CLI app with runtime flags.
- ✅ Add configuration validation (required fields, values, regex validation).
- ✅ Add strategy-based configuration loading from yaml/toml/json/conf.
- ✅ Add combined config mode (file + env + CLI).
- ✅ Extend platform usage: local, container scheduler, system scheduler.
- ✅ Make utility ready for use inside Docker containers.
- ✅ Add flexible run modes with dynamic scheduling configuration.
- ⏳ Refactor deeper into reusable packages.
- ⏳ Support other SQL dialects (PostgreSQL, MSSQL, etc).
- ⏳ Support more dump formats (plain SQL, CSV, binary).


## ✅ Task итог (pre-merge summary)
What is implemented in this iteration:
- Refactored runtime into explicit packages (`internal/app`, `internal/config`, `internal/pipeline`, `internal/filter`) and reduced `main.go` to bootstrap + graceful shutdown.
- Added schedule mode (`MODE=schedule`, `SCHEDULE_EVERY`) and runtime flags for operational control.
- Added strategy-based layered config (defaults -> file -> env -> CLI) with format support: YAML/TOML/JSON/CONF.
- Added stress dump generator `cmd/dumpgen` with 1GB-scale targets, multi-table generation, random payloads, and deterministic seed support.
- Added deployment artifacts for container and system scheduler (`Dockerfile`, `docker-compose.yml`, `deploy/systemd/*`).
- Added/updated tests for config/filter/generator behavior.

Current limitation to keep in mind before merge:
- Native S3 read/write in application runtime is not implemented yet (MinIO is present in infra examples, but app I/O path is currently local FS).

Recommended next PR:
- Add universal storage backend API and implement `s3://bucket/key` input/output using AWS SDK v2 with MinIO-compatible endpoint/path-style options.

## 📜 License
MIT.

# 📦 Filtrate Backups
Filtrate Backups is a Go utility for filtering SQL dump archives.
It unpacks `.tar.gz` backups, removes unwanted `INSERT` data for selected tables, and repacks a cleaned dump.

## 🚀 Features
- Streams dump files line-by-line.
- Handles very large SQL lines with configurable memory limits (`MAX_LINE_BYTES`).
- Works with local filesystem paths.
- Runs once or as an internal scheduler (`MODE=schedule`, `SCHEDULE_EVERY=...`).
- Supports deployment as:
  - a containerized scheduler,
  - a scheduler near a dedicated S3 service (e.g. MinIO),
  - a system scheduler via `systemd` timer.

## ⚙️ Configuration
Use `.env` or environment variables:

```env
DUMPFILE="./data/source.tar.gz"
OUTPUT_FILE="./output/filtered_result.tar.gz"
TABLE_MAP="^tmp_:^log_"                      # colon-separated regex patterns
TMP_DIR="./tmp"
MAX_LINE_BYTES=8388608
MODE="once"                                  # once | schedule
SCHEDULE_EVERY="1h"                          # required for schedule mode
```

## 🗂️ CLI usage
```bash
go run . --input ./dump.tar.gz --output ./output/filtered_result.tar.gz --skip '^tmp_:^log_'
```

Useful flags:
- `--mode once|schedule`
- `--every 30m`
- `--max-line-bytes 16777216`

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
- ✅ Extend platform usage: local, container scheduler, system scheduler.
- ✅ Make utility ready for use inside Docker containers.
- ✅ Add flexible run modes with dynamic scheduling configuration.
- ⏳ Refactor deeper into reusable packages.
- ⏳ Support other SQL dialects (PostgreSQL, MSSQL, etc).
- ⏳ Support more dump formats (plain SQL, CSV, binary).

## 📜 License
MIT.

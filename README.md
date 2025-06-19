# 📦 Filtrate Backups
Filtrate Backups is a simple, efficient Go utility for filtering SQL dump archives.
It unpacks .tar.gz backups, removes unwanted INSERT data for selected tables, and repacks the cleaned dump — saving space and speeding up imports.

## 🚀 Features
✅ Unpacks .tar.gz SQL dump archives
✅ Filters out INSERT lines matching given table patterns (supports regex)
✅ Streams files line-by-line
✅ Creates a new filtered .tar.gz archive
✅ Reports processing time and memory usage
✅ Cleans up all temporary files automatically

## ⚙️ Configuration
Use a .env file or environment variables:

```env
DUMPFILE="dump.tar.gz"       # Path to the input archive
TABLE_MAP="^tmp_:^log_"      # Colon-separated list of regex patterns to skip
TMP_DIR="./tmp"              # Directory for temporary files
```
## 🗂️ Usage
1️. Create and configure .env in your project root.

2️. Run:

```bash
go run main.go
```
## 3️⃣ After completion:

The filtered dump will be available as filtered_result.tar.gz in your TMP_DIR.

All temporary extraction files are automatically removed.

✅ Example TABLE_MAP patterns
```Pattern	Description
^tmp_	Skip all tables starting with tmp_
^b_tmp_:^log_	Skip b_tmp_* and log_*
^.*_backup$	Skip all tables ending with _backup
```
## Clean working directory
The tool automatically deletes all temporary files after packing the final archive.
To keep the output archive, it is copied outside the temporary folder before cleanup.

## 🔭 Roadmap
✅ Current: Basic filtering for MySQL-compatible dumps (plain SQL).

### 🛠️ Planned:

- Make smalless cli app

- Refactor the project structure into smaller reusable packages.

- Add configuration validation (required fields, allowed values).

- Extend the filter to support other SQL dialects (PostgreSQL, MSSQL, etc).

- Extend platform use

- Support more flexible dump formats (plain SQL, CSV, binary).

- Add more CLI flags and dynamic configuration.

### 📜 License
MIT — free for personal and commercial use.

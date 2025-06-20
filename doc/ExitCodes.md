# Exit codes

| Exit Code | Name (BSD `sysexits.h`)     | When to use                                                                       | Source                                                                           |
| --------- | --------------------------- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `0`       | `EX_OK`                     | ✅ **Success** — everything worked as expected.                                    | [POSIX exit(3)](https://man7.org/linux/man-pages/man3/exit.3.html)               |
| `1`       | *(Custom, general failure)* | ⚠️ **General error** — unexpected internal error.                                 | [Shell conventions](https://en.wikipedia.org/wiki/Exit_status#Shell_conventions) |
| `2`       | `EX_USAGE`                  | ⚠️ **Incorrect command usage** — invalid flag, bad arguments, unknown command.    | [sysexits(3)](https://man7.org/linux/man-pages/man3/sysexits.3.html)             |
| `66`      | `EX_NOINPUT`                | 📂 **Missing input file** — file not found or cannot be opened.                   | [sysexits(3)](https://man7.org/linux/man-pages/man3/sysexits.3.html)             |
| `74`      | `EX_IOERR`                  | 🗂️ **I/O error** — file exists but failed to read/write.                         | [sysexits(3)](https://man7.org/linux/man-pages/man3/sysexits.3.html)             |
| `126`     | *(Shell convention)*        | 🔒 **Permission denied** — cannot execute file/script due to lack of permissions. | [Shell conventions](https://en.wikipedia.org/wiki/Exit_status#Shell_conventions) |
| `127`     | *(Shell convention)*        | ❓ **Command not found** — external binary/script not found.                       | [Shell conventions](https://en.wikipedia.org/wiki/Exit_status#Shell_conventions) |
| `130`     | *(128 + 2)*                 | ⏹️ **Interrupted by Ctrl+C** (`SIGINT`).                                          | [Shell conventions](https://en.wikipedia.org/wiki/Exit_status#Shell_conventions) |

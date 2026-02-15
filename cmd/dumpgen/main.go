package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/d00p1/filtrate-backups/pkg/archive"
)

func main() {
	var (
		output        string
		tableName     string
		targetSizeRaw string
		rowsPerInsert int
		valueSize     int
	)

	flag.StringVar(&output, "output", "./data/generated_dump_1gb.tar.gz", "output tar.gz path")
	flag.StringVar(&tableName, "table", "bench_data", "table name for generated INSERTs")
	flag.StringVar(&targetSizeRaw, "target-size", "1GB", "target SQL payload size (e.g. 1GB, 1024MB, 500MiB)")
	flag.IntVar(&rowsPerInsert, "rows-per-insert", 1000, "rows per INSERT statement")
	flag.IntVar(&valueSize, "value-size", 128, "string payload size per row")
	flag.Parse()

	targetBytes, err := parseSize(targetSizeRaw)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if rowsPerInsert <= 0 || valueSize <= 0 {
		fmt.Println("Error: rows-per-insert and value-size must be > 0")
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	sqlBytes, err := generateTarGz(output, tableName, targetBytes, rowsPerInsert, valueSize)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("✅ generated: %s (sql bytes: %d)\n", output, sqlBytes)
}

func generateTarGz(output, tableName string, targetBytes int64, rowsPerInsert, valueSize int) (int64, error) {
	tmpDir, err := os.MkdirTemp("", "dumpgen-")
	if err != nil {
		return 0, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	sqlPath := filepath.Join(tmpDir, "dump.sql")
	sqlFile, err := os.Create(sqlPath)
	if err != nil {
		return 0, fmt.Errorf("create temp sql: %w", err)
	}

	sqlBytes, err := writeSQL(sqlFile, tableName, targetBytes, rowsPerInsert, valueSize)
	if closeErr := sqlFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return 0, err
	}

	out, err := os.Create(output)
	if err != nil {
		return 0, fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	if err := archive.Pack(tmpDir, gz); err != nil {
		return 0, fmt.Errorf("pack tar.gz: %w", err)
	}

	return sqlBytes, nil
}

func writeSQL(w io.Writer, tableName string, targetBytes int64, rowsPerInsert, valueSize int) (int64, error) {
	base := "CREATE TABLE `" + tableName + "` (id BIGINT PRIMARY KEY, payload TEXT);\n"
	n, err := io.WriteString(w, base)
	if err != nil {
		return 0, fmt.Errorf("write ddl: %w", err)
	}

	payload := strings.Repeat("x", valueSize)
	written := int64(n)
	var id int64 = 1

	for written < targetBytes {
		var b strings.Builder
		b.Grow(rowsPerInsert * (valueSize + 32))
		b.WriteString("INSERT INTO `")
		b.WriteString(tableName)
		b.WriteString("` VALUES ")
		for i := 0; i < rowsPerInsert; i++ {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("(")
			b.WriteString(strconv.FormatInt(id, 10))
			b.WriteString(",'")
			b.WriteString(payload)
			b.WriteString("')")
			id++
		}
		b.WriteString(";\n")

		line := b.String()
		n, err := io.WriteString(w, line)
		if err != nil {
			return written, fmt.Errorf("write insert: %w", err)
		}
		written += int64(n)
	}

	return written, nil
}

func parseSize(raw string) (int64, error) {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	ordered := []struct {
		suffix string
		mul    int64
	}{
		{"GIB", 1024 * 1024 * 1024},
		{"MIB", 1024 * 1024},
		{"KIB", 1024},
		{"TB", 1000 * 1000 * 1000 * 1000},
		{"GB", 1000 * 1000 * 1000},
		{"MB", 1000 * 1000},
		{"KB", 1000},
		{"B", 1},
	}

	for _, item := range ordered {
		if strings.HasSuffix(raw, item.suffix) {
			numPart := strings.TrimSpace(strings.TrimSuffix(raw, item.suffix))
			v, err := strconv.ParseInt(numPart, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size %q", raw)
			}
			if v <= 0 {
				return 0, fmt.Errorf("size must be > 0")
			}
			return v * item.mul, nil
		}
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unknown size format %q", raw)
	}
	if v <= 0 {
		return 0, fmt.Errorf("size must be > 0")
	}
	return v, nil
}

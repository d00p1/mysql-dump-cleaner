package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/d00p1/filtrate-backups/pkg/archive"
	"io"
	"os"
	"path"
	"regexp"
	"path/filepath"
)

type Config struct {
	Dump       string   `env:"DUMPFILE,required"`
	TablesSkip []string `env:"TABLE_MAP" envSeparator:":"`
	TmpDir     string   `env:"TMP_DIR"`
	output   string   `env:"OUTPUT_DIR" envDefault:"./output"`
	tokenSize int   `env:"TOKEN_SIZE" envDefault:"65536"` // buffer size for reading lines, default 64 KB
}

func main() {
	if err := filtrateDump(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func filtrateDump() error {
	godotenv.Load("./.env")

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}

	f, err := os.Open(cfg.Dump)
	if err != nil {
		return fmt.Errorf("failed to open dump: %w", err)
	}
	defer f.Close()

	cr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader error: %w", err)
	}
	defer cr.Close()

	// Правильный вызов MkdirTemp и использование пути:
	tmpDir, err := os.MkdirTemp(cfg.TmpDir, "cache")
	if err != nil {
		return fmt.Errorf("MkdirTemp error: %w", err)
	}

	// Unpack архив в tmpDir
	if err := archive.Unpack(cr, tmpDir); err != nil {
		return fmt.Errorf("Error unpack archive: %w", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read extracted files: %w", err)
	}

	// Создаем папку filtered внутри tmpDir
	filteredDir := path.Join(tmpDir, "filtered")
	if err := os.MkdirAll(filteredDir, 0755); err != nil {
		return fmt.Errorf("failed to create filtered dir: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		file, err := os.Open(path.Join(tmpDir, f.Name()))
		if err != nil {
			continue
		}
		defer file.Close()

		outPut, err := os.Create(path.Join(filteredDir, f.Name()))
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outPut.Close()

		if err := insertFilter(file, outPut, cfg.TablesSkip, cfg.tokenSize); err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}
	// Упаковать filtered/ обратно
	outputTarGz := filepath.Join(cfg.output, "filtered_result.tar.gz")
	if err := packToTarGz(filteredDir, outputTarGz); err != nil {
		return fmt.Errorf("failed to pack archive: %w", err)
	}

	defer os.RemoveAll(tmpDir)
	fmt.Printf("✅ New archive: %s\n", outputTarGz)
	return nil
}

func insertFilter(r io.Reader, w io.Writer, skipTables []string, bufferSize int) error {
	
	scanner := bufio.NewScanner(r)
	scanner.Buffer(nil, bufferSize) // Увеличиваем буфер до 10 МБ


	writer := bufio.NewWriter(w)

	defer writer.Flush()
	var skipPatterns []*regexp.Regexp
	
	for _, pat := range skipTables{
		re, err := regexp.Compile(pat)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pat, err)
		}
		skipPatterns = append(skipPatterns, re)
	}

	reInsert := regexp.MustCompile(`^INSERT INTO ` + "`?" + `([^` + "`" + ` ]*)` + "`?")

	insideInsertBlock := false

	var filteredLines, totalLines int

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		if matches := reInsert.FindStringSubmatch(line); matches != nil {
			currentTable := matches[1]

			// Проверить после цикла!
			skipThisInsert := false
			for _, re := range skipPatterns {
				if re.MatchString(currentTable) {
					skipThisInsert = true
					break
				}
			}

			if skipThisInsert {
				insideInsertBlock = true
				filteredLines++
				continue
			}
		}

		if insideInsertBlock {
			filteredLines++
			if len(line) > 0 && line[len(line)-1] == ';' {
				insideInsertBlock = false
			}
			continue
		}

		_, _ = writer.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %w", err)
	}

	fmt.Printf("✅ Skipped lines: %d из %d\n", filteredLines, totalLines)
	return nil
}

func packToTarGz(srcDir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	if err := archive.Pack(srcDir, gzWriter); err != nil {
		return fmt.Errorf("failed to pack directory: %w", err)
	}

	return nil
}

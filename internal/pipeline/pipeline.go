package pipeline

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/d00p1/filtrate-backups/internal/filter"
	"github.com/d00p1/filtrate-backups/pkg/archive"
)

type Options struct {
	InputPath    string
	OutputPath   string
	TablesSkip   []string
	TmpDir       string
	MaxLineBytes int
}

type Result struct {
	OutputPath    string
	TotalLines    int
	FilteredLines int
}

func Run(opts Options) (Result, error) {
	tmpDir, err := os.MkdirTemp(opts.TmpDir, "cache-")
	if err != nil {
		return Result{}, fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile, err := os.Open(opts.InputPath)
	if err != nil {
		return Result{}, fmt.Errorf("open input: %w", err)
	}
	defer inputFile.Close()

	gzReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return Result{}, fmt.Errorf("gzip reader error: %w", err)
	}
	defer gzReader.Close()

	if err := archive.Unpack(gzReader, tmpDir); err != nil {
		return Result{}, fmt.Errorf("unpack archive: %w", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return Result{}, fmt.Errorf("read extracted files: %w", err)
	}

	filteredDir := filepath.Join(tmpDir, "filtered")
	if err := os.MkdirAll(filteredDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create filtered dir: %w", err)
	}

	var totalLines, filteredLines int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(tmpDir, entry.Name())
		dstPath := filepath.Join(filteredDir, entry.Name())

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return Result{}, fmt.Errorf("open extracted file: %w", err)
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			srcFile.Close()
			return Result{}, fmt.Errorf("create filtered file: %w", err)
		}

		stats, err := filter.InsertFilter(srcFile, dstFile, opts.TablesSkip, opts.MaxLineBytes)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return Result{}, fmt.Errorf("filter %s: %w", entry.Name(), err)
		}

		totalLines += stats.TotalLines
		filteredLines += stats.FilteredLines
	}

	if err := packToTarGz(filteredDir, opts.OutputPath); err != nil {
		return Result{}, err
	}

	return Result{OutputPath: opts.OutputPath, TotalLines: totalLines, FilteredLines: filteredLines}, nil
}

func packToTarGz(srcDir, outputFile string) error {
	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	if err := archive.Pack(srcDir, gzWriter); err != nil {
		return fmt.Errorf("pack directory: %w", err)
	}
	return nil
}

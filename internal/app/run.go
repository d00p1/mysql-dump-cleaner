package app

import (
	"context"
	"fmt"
	"time"

	"github.com/d00p1/filtrate-backups/internal/config"
	"github.com/d00p1/filtrate-backups/internal/pipeline"
)

func Run(ctx context.Context, args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}

	runOnce := func() error {
		result, err := pipeline.Run(pipeline.Options{
			InputPath:    cfg.Input,
			OutputPath:   cfg.Output,
			TablesSkip:   cfg.TablesSkip,
			TmpDir:       cfg.TmpDir,
			MaxLineBytes: cfg.MaxLineBytes,
		})
		if err != nil {
			return err
		}

		fmt.Printf("✅ filtered lines: %d/%d\n", result.FilteredLines, result.TotalLines)
		fmt.Printf("✅ output: %s\n", result.OutputPath)
		return nil
	}

	if cfg.Mode == "once" {
		return runOnce()
	}

	ticker := time.NewTicker(cfg.ScheduleInterval)
	defer ticker.Stop()

	if err := runOnce(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := runOnce(); err != nil {
				fmt.Printf("run failed: %v\n", err)
			}
		}
	}
}

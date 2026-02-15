package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Input            string        `env:"DUMPFILE,required"`
	Output           string        `env:"OUTPUT_FILE" envDefault:"./output/filtered_result.tar.gz"`
	TablesSkipRaw    string        `env:"TABLE_MAP"`
	TmpDir           string        `env:"TMP_DIR" envDefault:"./tmp"`
	MaxLineBytes     int           `env:"MAX_LINE_BYTES" envDefault:"8388608"`
	ScheduleInterval time.Duration `env:"SCHEDULE_EVERY" envDefault:"0"`
	Mode             string        `env:"MODE" envDefault:"once"`

	TablesSkip []string
}

func Load(args []string) (Config, error) {
	_ = godotenv.Load("./.env")

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("env parse error: %w", err)
	}

	fs := flag.NewFlagSet("mysql-dump-cleaner", flag.ContinueOnError)
	fs.StringVar(&cfg.Input, "input", cfg.Input, "input dump path")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "output archive path")
	fs.StringVar(&cfg.TablesSkipRaw, "skip", cfg.TablesSkipRaw, "colon-separated regex list of tables to remove")
	fs.StringVar(&cfg.TmpDir, "tmp-dir", cfg.TmpDir, "tmp directory")
	fs.IntVar(&cfg.MaxLineBytes, "max-line-bytes", cfg.MaxLineBytes, "max bytes per SQL line")
	fs.DurationVar(&cfg.ScheduleInterval, "every", cfg.ScheduleInterval, "run as scheduler with interval, e.g. 30m")
	fs.StringVar(&cfg.Mode, "mode", cfg.Mode, "run mode: once or schedule")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.TablesSkip = splitPatterns(cfg.TablesSkipRaw)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func splitPatterns(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ":")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func validate(cfg Config) error {
	var allErrs []error

	if cfg.Mode != "once" && cfg.Mode != "schedule" {
		allErrs = append(allErrs, fmt.Errorf("MODE must be once or schedule, got %q", cfg.Mode))
	}
	if cfg.MaxLineBytes < 1024 {
		allErrs = append(allErrs, errors.New("MAX_LINE_BYTES must be >= 1024"))
	}
	if cfg.Mode == "schedule" && cfg.ScheduleInterval <= 0 {
		allErrs = append(allErrs, errors.New("SCHEDULE_EVERY (or --every) is required for schedule mode"))
	}
	if err := ensureDir(cfg.TmpDir); err != nil {
		allErrs = append(allErrs, fmt.Errorf("TMP_DIR error: %w", err))
	}
	for _, pat := range cfg.TablesSkip {
		if _, err := regexp.Compile(pat); err != nil {
			allErrs = append(allErrs, fmt.Errorf("invalid TABLE_MAP pattern %q: %w", pat, err))
		}
	}

	return errors.Join(allErrs...)
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return nil
}

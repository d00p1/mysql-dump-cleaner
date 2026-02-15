package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Input            string
	Output           string
	TablesSkipRaw    string
	TmpDir           string
	MaxLineBytes     int
	ScheduleInterval time.Duration
	Mode             string
	TablesSkip       []string
}

type bootstrapOptions struct {
	ConfigPath     string
	ConfigFormat   string
	ConfigStrategy string
}

func Load(args []string) (Config, error) {
	_ = godotenv.Load("./.env")

	boot, err := parseBootstrap(args)
	if err != nil {
		return Config{}, err
	}

	cfg := defaultConfig()

	if boot.ConfigPath != "" {
		strategy, err := ResolveStrategy(boot.ConfigFormat, boot.ConfigPath)
		if err != nil {
			return Config{}, err
		}
		fileValues, err := strategy.Load(boot.ConfigPath)
		if err != nil {
			return Config{}, fmt.Errorf("load config file: %w", err)
		}
		cfg.applyKeyValues(fileValues)
	}

	if boot.ConfigStrategy == "merge" || boot.ConfigStrategy == "env-only" {
		cfg.applyKeyValues(readKnownEnv())
	}

	if err := applyCLIOverrides(args, &cfg); err != nil {
		return Config{}, err
	}

	cfg.TablesSkip = splitPatterns(cfg.TablesSkipRaw)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func parseBootstrap(args []string) (bootstrapOptions, error) {
	boot := bootstrapOptions{ConfigFormat: "auto", ConfigStrategy: "merge"}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := func() string {
			if i+1 < len(args) {
				i++
				return args[i]
			}
			return ""
		}

		switch {
		case arg == "--config":
			boot.ConfigPath = next()
		case strings.HasPrefix(arg, "--config="):
			boot.ConfigPath = strings.TrimPrefix(arg, "--config=")
		case arg == "--config-format":
			boot.ConfigFormat = next()
		case strings.HasPrefix(arg, "--config-format="):
			boot.ConfigFormat = strings.TrimPrefix(arg, "--config-format=")
		case arg == "--config-strategy":
			boot.ConfigStrategy = next()
		case strings.HasPrefix(arg, "--config-strategy="):
			boot.ConfigStrategy = strings.TrimPrefix(arg, "--config-strategy=")
		}
	}

	if boot.ConfigStrategy != "merge" && boot.ConfigStrategy != "file-only" && boot.ConfigStrategy != "env-only" {
		return bootstrapOptions{}, fmt.Errorf("invalid --config-strategy=%q", boot.ConfigStrategy)
	}
	if boot.ConfigStrategy == "file-only" && boot.ConfigPath == "" {
		return bootstrapOptions{}, errors.New("--config is required when --config-strategy=file-only")
	}
	return boot, nil
}

func applyCLIOverrides(args []string, cfg *Config) error {
	fs := flag.NewFlagSet("mysql-dump-cleaner", flag.ContinueOnError)
	fs.StringVar(&cfg.Input, "input", cfg.Input, "input dump path")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "output archive path")
	fs.StringVar(&cfg.TablesSkipRaw, "skip", cfg.TablesSkipRaw, "colon-separated regex list of tables to remove")
	fs.StringVar(&cfg.TmpDir, "tmp-dir", cfg.TmpDir, "tmp directory")
	fs.IntVar(&cfg.MaxLineBytes, "max-line-bytes", cfg.MaxLineBytes, "max bytes per SQL line")
	fs.DurationVar(&cfg.ScheduleInterval, "every", cfg.ScheduleInterval, "run as scheduler with interval, e.g. 30m")
	fs.StringVar(&cfg.Mode, "mode", cfg.Mode, "run mode: once or schedule")

	var ignored string
	fs.StringVar(&ignored, "config", "", "")
	fs.StringVar(&ignored, "config-format", "", "")
	fs.StringVar(&ignored, "config-strategy", "", "")

	return fs.Parse(args)
}

func (cfg *Config) applyKeyValues(values map[string]string) {
	for key, value := range values {
		norm := normalizeKey(key)
		switch norm {
		case "DUMPFILE", "INPUT":
			if value != "" {
				cfg.Input = value
			}
		case "OUTPUT_FILE", "OUTPUT", "OUTPUT_DIR":
			if value != "" {
				cfg.Output = value
			}
		case "TABLE_MAP", "TABLES_SKIP", "SKIP", "SKIP_TABLES":
			cfg.TablesSkipRaw = normalizePatterns(value)
		case "TMP_DIR", "TMPDIR":
			if value != "" {
				cfg.TmpDir = value
			}
		case "MAX_LINE_BYTES", "TOKEN_SIZE":
			if parsed, err := parseInt(value); err == nil && parsed > 0 {
				cfg.MaxLineBytes = parsed
			}
		case "MODE":
			if value != "" {
				cfg.Mode = strings.ToLower(value)
			}
		case "SCHEDULE_EVERY", "EVERY", "INTERVAL":
			if d, err := time.ParseDuration(value); err == nil {
				cfg.ScheduleInterval = d
			}
		}
	}
}

func defaultConfig() Config {
	return Config{
		Output:           "./output/filtered_result.tar.gz",
		TmpDir:           "./tmp",
		MaxLineBytes:     8 * 1024 * 1024,
		ScheduleInterval: 0,
		Mode:             "once",
	}
}

func splitPatterns(raw string) []string {
	if raw == "" {
		return nil
	}
	raw = strings.ReplaceAll(raw, ",", ":")
	parts := strings.Split(raw, ":")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.Trim(strings.TrimSpace(p), "\"'")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func normalizePatterns(v string) string {
	clean := strings.TrimSpace(v)
	clean = strings.TrimPrefix(clean, "[")
	clean = strings.TrimSuffix(clean, "]")
	clean = strings.ReplaceAll(clean, "\"", "")
	clean = strings.ReplaceAll(clean, "'", "")
	clean = strings.ReplaceAll(clean, ",", ":")
	return clean
}

func validate(cfg Config) error {
	var allErrs []error

	if cfg.Input == "" {
		allErrs = append(allErrs, errors.New("DUMPFILE (or --input) is required"))
	}
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

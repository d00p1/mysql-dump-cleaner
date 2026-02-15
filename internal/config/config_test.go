package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromJSONAndEnvOverride(t *testing.T) {
	t.Setenv("DUMPFILE", "")
	t.Setenv("OUTPUT_FILE", "")
	t.Setenv("TABLE_MAP", "")
	t.Setenv("TMP_DIR", "")
	t.Setenv("MAX_LINE_BYTES", "")
	t.Setenv("MODE", "")
	t.Setenv("SCHEDULE_EVERY", "")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{"DUMPFILE":"/data/in.tar.gz","OUTPUT_FILE":"/data/out.tar.gz","TABLE_MAP":["^tmp_","^log_"],"MODE":"schedule","SCHEDULE_EVERY":"15m"}`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MODE", "once")

	cfg, err := Load([]string{"--config", cfgPath, "--config-format", "json"})
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.Mode != "once" {
		t.Fatalf("env should override file mode, got %q", cfg.Mode)
	}
	if cfg.Input != "/data/in.tar.gz" {
		t.Fatalf("unexpected input: %s", cfg.Input)
	}
	if len(cfg.TablesSkip) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(cfg.TablesSkip))
	}
}

func TestLoadFromYAMLWithCLIOverride(t *testing.T) {
	t.Setenv("DUMPFILE", "")
	t.Setenv("OUTPUT_FILE", "")
	t.Setenv("TABLE_MAP", "")
	t.Setenv("TMP_DIR", "")
	t.Setenv("MAX_LINE_BYTES", "")
	t.Setenv("MODE", "")
	t.Setenv("SCHEDULE_EVERY", "")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := "DUMPFILE: /yaml/in.tar.gz\nOUTPUT_FILE: /yaml/out.tar.gz\nMODE: once\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load([]string{"--config", cfgPath, "--input", "/cli/in.tar.gz"})
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Input != "/cli/in.tar.gz" {
		t.Fatalf("cli override failed, got %s", cfg.Input)
	}
}

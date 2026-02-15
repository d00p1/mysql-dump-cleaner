package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LoadStrategy interface {
	Load(path string) (map[string]string, error)
}

type strategyFunc func(path string) (map[string]string, error)

func (f strategyFunc) Load(path string) (map[string]string, error) {
	return f(path)
}

func ResolveStrategy(format, path string) (LoadStrategy, error) {
	if format == "auto" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yaml", ".yml":
			format = "yaml"
		case ".toml":
			format = "toml"
		case ".json":
			format = "json"
		case ".conf", ".cfg", ".ini":
			format = "conf"
		default:
			return nil, fmt.Errorf("unable to detect config format for %q", path)
		}
	}

	switch strings.ToLower(format) {
	case "yaml", "yml":
		return strategyFunc(loadYAML), nil
	case "toml":
		return strategyFunc(loadKVEquals), nil
	case "json":
		return strategyFunc(loadJSON), nil
	case "conf", "cfg", "ini":
		return strategyFunc(loadKVEquals), nil
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}
}

func readKnownEnv() map[string]string {
	keys := []string{"DUMPFILE", "OUTPUT_FILE", "TABLE_MAP", "TMP_DIR", "MAX_LINE_BYTES", "MODE", "SCHEDULE_EVERY"}
	res := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
			res[k] = v
		}
	}
	return res
}

func parseInt(v string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(v))
}

func normalizeKey(k string) string {
	k = strings.TrimSpace(strings.ToUpper(k))
	k = strings.ReplaceAll(k, ".", "_")
	k = strings.ReplaceAll(k, "-", "_")
	return k
}

func loadJSON(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(raw))
	for k, v := range raw {
		switch t := v.(type) {
		case string:
			out[k] = t
		case float64:
			out[k] = strconv.FormatInt(int64(t), 10)
		case bool:
			out[k] = strconv.FormatBool(t)
		case []any:
			parts := make([]string, 0, len(t))
			for _, item := range t {
				parts = append(parts, fmt.Sprint(item))
			}
			out[k] = strings.Join(parts, ":")
		default:
			out[k] = fmt.Sprint(t)
		}
	}
	return out, nil
}

func loadYAML(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.Trim(strings.TrimSpace(line[idx+1:]), "\"'")
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func loadKVEquals(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.Trim(strings.TrimSpace(line[idx+1:]), "\"'")
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

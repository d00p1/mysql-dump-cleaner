package filter

import (
	"bytes"
	"strings"
	"testing"
)

func TestInsertFilterLargeLines(t *testing.T) {
	longValue := strings.Repeat("x", 2*1024*1024)
	input := "CREATE TABLE `tmp_log` (id int);\n" +
		"INSERT INTO `tmp_log` VALUES (1, '" + longValue + "');\n" +
		"CREATE TABLE `users` (id int);\n"

	var out bytes.Buffer
	stats, err := InsertFilter(strings.NewReader(input), &out, []string{"^tmp_"}, 4*1024*1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out.String(), "INSERT INTO `tmp_log`") {
		t.Fatalf("tmp_log insert should be removed")
	}
	if !strings.Contains(out.String(), "CREATE TABLE `users`") {
		t.Fatalf("users DDL should remain")
	}
	if stats.FilteredLines == 0 {
		t.Fatalf("expected filtered lines > 0")
	}
}

func TestInsertFilterFailsWhenLineLimitExceeded(t *testing.T) {
	longValue := strings.Repeat("x", 2048)
	input := "INSERT INTO `tmp_log` VALUES ('" + longValue + "');\n"

	_, err := InsertFilter(strings.NewReader(input), &bytes.Buffer{}, []string{"^tmp_"}, 1024)
	if err == nil {
		t.Fatalf("expected line-limit error")
	}
}

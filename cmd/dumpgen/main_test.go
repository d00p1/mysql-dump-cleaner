package main

import "testing"

func TestParseSize(t *testing.T) {
	tests := map[string]int64{
		"1GB":   1000 * 1000 * 1000,
		"512MB": 512 * 1000 * 1000,
		"1GiB":  1024 * 1024 * 1024,
		"2048":  2048,
	}

	for in, want := range tests {
		got, err := parseSize(in)
		if err != nil {
			t.Fatalf("parseSize(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("parseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseSizeInvalid(t *testing.T) {
	if _, err := parseSize("0GB"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := parseSize("abc"); err == nil {
		t.Fatalf("expected error")
	}
}

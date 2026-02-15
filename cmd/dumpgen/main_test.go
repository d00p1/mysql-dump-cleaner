package main

import (
	"math/rand"
	"strings"
	"testing"
)

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

func TestParseTables(t *testing.T) {
	tables := parseTables("users, orders,users, events ,, ")
	if len(tables) != 3 {
		t.Fatalf("expected 3 tables, got %d", len(tables))
	}
	if tables[0] != "users" || tables[1] != "orders" || tables[2] != "events" {
		t.Fatalf("unexpected table list: %#v", tables)
	}
}

func TestRandomString(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	s := randomString(rng, 32)
	if len(s) != 32 {
		t.Fatalf("expected len 32, got %d", len(s))
	}
	if strings.Count(s, string(s[0])) == len(s) {
		t.Fatalf("string looks non-random: %q", s)
	}
}

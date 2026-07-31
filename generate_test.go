package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)

func TestGenerateIsDeterministic(t *testing.T) {
	a := generate(1, 5000, 24*time.Hour, testNow)
	b := generate(1, 5000, 24*time.Hour, testNow)

	if len(a) != len(b) {
		t.Fatalf("lengths differ: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("entry %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestGenerateSeedChangesCorpus(t *testing.T) {
	a := generate(1, 100, time.Hour, testNow)
	b := generate(2, 100, time.Hour, testNow)

	same := true
	for i := range a {
		if a[i] != b[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("different seeds produced an identical corpus")
	}
}

func TestErrorsAreRareButPresent(t *testing.T) {
	total := 10000
	entries := generate(1, total, 24*time.Hour, testNow)

	errorLines := 0
	refused := 0
	for _, e := range entries {
		if e.Level == "error" {
			errorLines++
		}
		if strings.Contains(e.Msg, "connection refused") {
			refused++
		}
	}

	if refused == 0 {
		t.Fatalf("no connection refused lines in %d", total)
	}
	if errorLines == 0 {
		t.Fatalf("no error lines in %d", total)
	}
	if errorLines > total*5/100 {
		t.Fatalf("error lines = %d of %d, want a rare minority", errorLines, total)
	}
}

func TestJSONLinesFormat(t *testing.T) {
	window := 24 * time.Hour
	entries := generate(1, 200, window, testNow)

	lines, err := jsonLines(entries)
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range bytes.Split(bytes.TrimSpace(lines), []byte("\n")) {
		var record map[string]string
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("invalid JSON line %q: %v", line, err)
		}
		for _, key := range []string{"_time", "_msg", "level", "service", "host"} {
			if record[key] == "" {
				t.Fatalf("line %q is missing %q", line, key)
			}
		}

		ts, err := time.Parse(time.RFC3339Nano, record["_time"])
		if err != nil {
			t.Fatalf("unparseable _time in %q: %v", line, err)
		}
		if ts.After(testNow) || ts.Before(testNow.Add(-window)) {
			t.Fatalf("_time %s is outside the window ending at %s", ts, testNow)
		}
	}
}

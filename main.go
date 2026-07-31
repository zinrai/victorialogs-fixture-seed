package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	addr := flag.String("addr", "", "VictoriaLogs jsonline insert endpoint, e.g. http://127.0.0.1:9481/insert/jsonline (required)")
	total := flag.Int("total", 1000000, "number of log lines to generate")
	window := flag.Duration("window", 24*time.Hour, "time range ending now over which log timestamps are spread")
	seed := flag.Int64("seed", 1, "random seed. The same seed and total produce the same lines")
	tenant := flag.String("tenant", "", "tenant to ingest into, as accountID or accountID:projectID, e.g. 42 or 7:9. Empty uses the default tenant")
	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s, commit: %s, date: %s\n", version, commit, date)
		return
	}

	if err := run(*addr, *total, *window, *seed, *tenant); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(addr string, total int, window time.Duration, seed int64, tenant string) error {
	if addr == "" {
		return fmt.Errorf("-addr is required")
	}
	if total < 1 {
		return fmt.Errorf("total must be at least 1")
	}
	if window <= 0 {
		return fmt.Errorf("window must be positive")
	}

	t, err := parseTenant(tenant)
	if err != nil {
		return err
	}

	target, err := insertURL(addr)
	if err != nil {
		return err
	}

	tenantNote := ""
	if t != nil {
		tenantNote = fmt.Sprintf(" tenant=%s:%s", t.account, t.project)
	}

	c := newCorpus(seed, window, time.Now().UTC())

	done := 0
	for done < total {
		n := min(batchSize, total-done)
		entries := make([]entry, 0, n)
		for i := 0; i < n; i++ {
			entries = append(entries, c.next())
		}

		lines, err := jsonLines(entries)
		if err != nil {
			return err
		}
		if err := pushBatch(target, t, lines); err != nil {
			return err
		}
		done += n
		fmt.Fprintf(os.Stderr, "seeded %d/%d lines to %s (window=%s seed=%d%s)\n", done, total, target, window, seed, tenantNote)
	}
	return nil
}

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// batchSize bounds client-side memory: lines are generated and inserted in
// batches instead of materializing the whole corpus.
const batchSize = 100000

type tenantID struct {
	account string
	project string
}

func parseTenant(s string) (*tenantID, error) {
	if s == "" {
		return nil, nil
	}

	account := s
	project := "0"
	if i := strings.Index(s, ":"); i >= 0 {
		account = s[:i]
		project = s[i+1:]
	}
	for _, part := range []string{account, project} {
		if _, err := strconv.ParseUint(part, 10, 32); err != nil {
			return nil, fmt.Errorf("invalid -tenant %q: want accountID or accountID:projectID as 32-bit unsigned integers", s)
		}
	}
	return &tenantID{account: account, project: project}, nil
}

func insertURL(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("parse -addr: %w", err)
	}

	// An explicit _stream_fields in -addr wins over the default.
	q := u.Query()
	if q.Get("_stream_fields") == "" {
		q.Set("_stream_fields", "service,host")
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func pushBatch(target string, t *tenantID, lines []byte) error {
	req, err := http.NewRequest("POST", target, bytes.NewReader(lines))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/stream+json")
	if t != nil {
		req.Header.Set("AccountID", t.account)
		req.Header.Set("ProjectID", t.project)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s responded %s: %s", target, resp.Status, bytes.TrimSpace(body))
	}
	return nil
}

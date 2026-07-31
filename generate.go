package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"time"
)

type entry struct {
	Time    string `json:"_time"`
	Msg     string `json:"_msg"`
	Level   string `json:"level"`
	Service string `json:"service"`
	Host    string `json:"host"`
}

type corpus struct {
	r      *rand.Rand
	v      *vocab
	sum    int
	window time.Duration
	now    time.Time
}

func newCorpus(seed int64, window time.Duration, now time.Time) *corpus {
	r := rand.New(rand.NewSource(seed))

	sum := 0
	for _, t := range templates {
		sum += t.weight
	}

	return &corpus{r: r, v: newVocab(r), sum: sum, window: window, now: now}
}

func (c *corpus) next() entry {
	n := c.r.Intn(c.sum)
	var tpl template
	for _, t := range templates {
		if n < t.weight {
			tpl = t
			break
		}
		n -= t.weight
	}

	host := pick(c.r, hosts[tpl.service])
	msg := tpl.render(c.r, c.v)
	ts := c.now.Add(-time.Duration(c.r.Int63n(int64(c.window))))

	return entry{
		Time:    ts.Format(time.RFC3339Nano),
		Msg:     msg,
		Level:   tpl.level,
		Service: tpl.service,
		Host:    host,
	}
}

func generate(seed int64, total int, window time.Duration, now time.Time) []entry {
	c := newCorpus(seed, window, now)

	entries := make([]entry, 0, total)
	for i := 0; i < total; i++ {
		entries = append(entries, c.next())
	}
	return entries
}

func jsonLines(entries []entry) ([]byte, error) {
	var buf bytes.Buffer
	for _, e := range entries {
		b, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

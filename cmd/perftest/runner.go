package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type step struct {
	name, url string
	markers   []string
}

func makeSteps(c config) []step {
	calendarURL := c.baseURL + "/book/" + url.PathEscape(c.userID)
	return []step{
		{"home", c.baseURL + "/", []string{"Планируйте встречи проще"}},
		{"calendar", calendarURL, []string{`id="booking-date"`}},
		{"calendar_date", calendarURL + "?" + url.Values{"date": {c.date}}.Encode(), []string{`id="booking-date"`, `value="` + c.date + `"`}},
	}
}

type sample struct {
	started         time.Time
	user, iteration int
	step, url       string
	status, bytes   int
	elapsed         time.Duration
	err             string
}

const maxResponseBytes = 8 << 20

func request(ctx context.Context, client *http.Client, s step, user, iteration int) sample {
	r := sample{started: time.Now(), user: user, iteration: iteration, step: s.name, url: s.url}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		r.err = err.Error()
		r.elapsed = time.Since(r.started)
		return r
	}
	req.Header.Set("User-Agent", "slot-booking-perftest/1.0")
	resp, err := client.Do(req)
	if err != nil {
		r.err = err.Error()
		r.elapsed = time.Since(r.started)
		return r
	}
	r.status = resp.StatusCode
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	closeErr := resp.Body.Close()
	r.bytes = len(body)
	r.elapsed = time.Since(r.started)
	switch {
	case readErr != nil:
		r.err = "read response: " + readErr.Error()
	case closeErr != nil:
		r.err = "close response: " + closeErr.Error()
	case len(body) > maxResponseBytes:
		r.err = "response exceeds 8 MiB limit"
	case resp.StatusCode != http.StatusOK:
		r.err = fmt.Sprintf("expected HTTP 200, got %d", resp.StatusCode)
	default:
		for _, marker := range s.markers {
			if !bytes.Contains(body, []byte(marker)) {
				r.err = "missing expected page content: " + marker
				break
			}
		}
	}
	return r
}

func preflight(ctx context.Context, client *http.Client, steps []step) error {
	for _, s := range steps {
		if r := request(ctx, client, s, 0, 0); r.err != "" {
			return fmt.Errorf("%s (%s): %s", s.name, s.url, r.err)
		}
	}
	return nil
}

type statistics struct {
	durations []float64
	failed    int
	bytes     int64
}

func (s *statistics) add(r sample) {
	s.durations = append(s.durations, float64(r.elapsed)/float64(time.Millisecond))
	s.bytes += int64(r.bytes)
	if r.err != "" {
		s.failed++
	}
}

type loadResult struct {
	started, finished time.Time
	interrupted       bool
	total             statistics
	byStep            map[string]*statistics
	errors            map[string]int
}

func waitFor(ctx context.Context, delay time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil
	}
}

func runLoad(parent context.Context, c config, client *http.Client, steps []step) (loadResult, error) {
	result := loadResult{byStep: make(map[string]*statistics), errors: make(map[string]int)}
	for _, s := range steps {
		result.byStep[s.name] = &statistics{}
	}
	file, err := os.Create(filepath.Join(c.out, "requests.csv"))
	if err != nil {
		return result, err
	}
	writer := csv.NewWriter(file)
	_ = writer.Write([]string{"started_at_utc", "user", "iteration", "step", "method", "url", "status", "duration_ms", "response_bytes", "error"})
	result.started = time.Now()
	// The deadline stops admission, not an already admitted HTTP request.
	// In-flight requests can finish within client.Timeout without artificial failures.
	admission, cancel := context.WithDeadline(parent, result.started.Add(c.duration))
	defer cancel()
	samples := make(chan sample, c.users)
	var workers sync.WaitGroup
	for user := 1; user <= c.users; user++ {
		workers.Add(1)
		go func(user int) {
			defer workers.Done()
			delay := time.Duration(float64(c.rampUp) * float64(user-1) / float64(c.users))
			if !waitFor(admission, time.Until(result.started.Add(delay))) {
				return
			}
			for iteration := 1; ; iteration++ {
				for _, s := range steps {
					if admission.Err() != nil {
						return
					}
					r := request(parent, client, s, user, iteration)
					samples <- r
					if !waitFor(admission, c.pause) {
						return
					}
				}
			}
		}(user)
	}
	go func() {
		workers.Wait()
		close(samples)
	}()
	for r := range samples {
		result.total.add(r)
		result.byStep[r.step].add(r)
		if r.err != "" {
			result.errors[r.step+": "+r.err]++
		}
		_ = writer.Write([]string{
			r.started.UTC().Format(time.RFC3339Nano), strconv.Itoa(r.user), strconv.Itoa(r.iteration), r.step,
			http.MethodGet, r.url, strconv.Itoa(r.status), fmt.Sprintf("%.3f", float64(r.elapsed)/float64(time.Millisecond)),
			strconv.Itoa(r.bytes), r.err,
		})
	}
	result.finished = time.Now()
	result.interrupted = parent.Err() != nil
	writer.Flush()
	writeErr := writer.Error()
	closeErr := file.Close()
	if writeErr != nil {
		return result, writeErr
	}
	return result, closeErr
}

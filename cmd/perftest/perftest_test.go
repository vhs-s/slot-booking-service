package main

import (
	"context"
	"encoding/csv"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type delayedBody struct {
	io.Reader
	delay  time.Duration
	closed bool
}

func (b *delayedBody) Read(p []byte) (int, error) {
	time.Sleep(b.delay)
	b.delay = 0
	return b.Reader.Read(p)
}

func (b *delayedBody) Close() error { b.closed = true; return nil }

func TestRequestIncludesBodyReadAndClosesIt(t *testing.T) {
	body := &delayedBody{Reader: strings.NewReader("expected page"), delay: 20 * time.Millisecond}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	got := request(context.Background(), client, step{"home", "http://example.test/", []string{"expected page"}}, 1, 1)
	if got.err != "" || got.elapsed < 20*time.Millisecond || !body.closed || got.bytes != len("expected page") {
		t.Fatalf("sample = %+v, closed = %t", got, body.closed)
	}
}

func TestRequestDetectsFailures(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		status     int
		wantError  string
	}{
		{"server_error", "expected", 500, "expected HTTP 200"},
		{"wrong_page", "login page", 200, "missing expected page content"},
		{"redirect", "expected", 302, "expected HTTP 200"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := &delayedBody{Reader: strings.NewReader(tt.body)}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.status, Body: body}, nil
			})}
			got := request(context.Background(), client, step{"test", "http://example.test/", []string{"expected"}}, 1, 1)
			if !strings.Contains(got.err, tt.wantError) || !body.closed {
				t.Fatalf("sample = %+v, closed = %t", got, body.closed)
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Millisecond, Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			<-r.Context().Done()
			return nil, r.Context().Err()
		})}
		got := request(context.Background(), client, step{name: "test", url: "http://example.test/"}, 1, 1)
		if got.err == "" || got.status != 0 {
			t.Fatalf("expected timeout: %+v", got)
		}
	})
}

func TestLoadDrainsConcurrentRequestsAfterAdmissionDeadline(t *testing.T) {
	const users = 3
	var started atomic.Int32
	allStarted := make(chan struct{})
	client := &http.Client{Timeout: time.Second, Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if started.Add(1) == users {
			close(allStarted)
		}
		select {
		case <-allStarted:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		// This response intentionally finishes after the admission deadline.
		time.Sleep(60 * time.Millisecond)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("page"))}, nil
	})}
	c := config{users: users, duration: 20 * time.Millisecond, out: t.TempDir()}
	steps := []step{{"test", "http://example.test/", []string{"page"}}}
	result, err := runLoad(context.Background(), c, client, steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.total.durations) != users || result.total.failed != 0 || result.finished.Sub(result.started) < 60*time.Millisecond {
		t.Fatalf("requests=%d failed=%d elapsed=%s", len(result.total.durations), result.total.failed, result.finished.Sub(result.started))
	}
	file, err := os.Open(filepath.Join(c.out, "requests.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil || len(rows) != users+1 {
		t.Fatalf("CSV rows=%d err=%v", len(rows), err)
	}
	if _, err := saveReport(c, steps, result); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(c.out, "summary.csv"))
	if err != nil {
		t.Fatal(err)
	}
	summary, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil || len(summary) != 3 || summary[2][0] != "TOTAL" || summary[2][1] != "3" || summary[2][3] != "0" {
		t.Fatalf("summary = %v, err = %v", summary, err)
	}
}

func TestLoadCancellationSavesFailedRequest(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		cancel()
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	c := config{users: 1, duration: time.Minute, out: t.TempDir()}
	result, err := runLoad(parent, c, client, []step{{name: "test", url: "http://example.test/"}})
	if err != nil || !result.interrupted || len(result.total.durations) != 1 || result.total.failed != 1 {
		t.Fatalf("interrupted=%t requests=%d failed=%d err=%v", result.interrupted, len(result.total.durations), result.total.failed, err)
	}
}

func TestCalculatePercentilesAndRate(t *testing.T) {
	s := statistics{failed: 3}
	for i := 100; i >= 1; i-- {
		s.durations = append(s.durations, float64(i))
	}
	m := calculate(s, 2)
	if math.Abs(m.mean-50.5) > 1e-9 || m.p50 != 50 || m.p95 != 95 || m.p99 != 99 || m.min != 1 || m.max != 100 || m.rps != 50 || m.errorPercent != 3 {
		t.Fatalf("metrics = %+v", m)
	}
	if s.durations[0] != 100 {
		t.Fatal("calculate modified the input samples")
	}
	if got := calculate(statistics{}, 0); got != (metrics{}) {
		t.Fatalf("empty metrics = %+v", got)
	}
}

func TestConfigRejectsInvalidExperiments(t *testing.T) {
	for _, args := range [][]string{
		{}, {"-user-id", "owner", "-users", "0"}, {"-user-id", "owner", "-date", "2026-02-30"},
		{"-user-id", "owner", "-duration", "1s", "-ramp-up", "1s"},
		{"-user-id", "owner", "-pause", "-1s"}, {"-user-id", "owner", "-base-url", "http://example.test/path"},
	} {
		if _, err := parseConfig(args, io.Discard); err == nil {
			t.Errorf("accepted invalid arguments: %v", args)
		}
	}
	if code := runCLI([]string{"-help"}, io.Discard, io.Discard); code != 0 {
		t.Fatalf("help exit code = %d", code)
	}
}

// Command perftest measures the public HTTP pages of a running booking service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	baseURL, userID, date, out string
	users                      int
	duration, pause, timeout   time.Duration
	rampUp                     time.Duration
}

func parseConfig(args []string, stderr io.Writer) (config, error) {
	var c config
	f := flag.NewFlagSet("perftest", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&c.baseURL, "base-url", "http://localhost:8080", "Running service URL")
	f.StringVar(&c.userID, "user-id", "", "Calendar owner's ID (required)")
	f.StringVar(&c.date, "date", time.Now().Format(time.DateOnly), "Calendar date, YYYY-MM-DD")
	f.IntVar(&c.users, "users", 1, "Concurrent virtual users")
	f.DurationVar(&c.duration, "duration", time.Minute, "Time window for starting requests")
	f.DurationVar(&c.pause, "pause", time.Second, "Pause after each request, per user")
	f.DurationVar(&c.rampUp, "ramp-up", 0, "Time over which virtual users are started")
	f.DurationVar(&c.timeout, "timeout", 5*time.Second, "Timeout for a complete HTTP request")
	f.StringVar(&c.out, "out", "", "New output directory (default: timestamp under results/performance)")
	if err := f.Parse(args); err != nil {
		return c, err
	}
	if f.NArg() != 0 {
		return c, fmt.Errorf("unexpected arguments: %s", strings.Join(f.Args(), " "))
	}
	c.baseURL = strings.TrimRight(c.baseURL, "/")
	u, err := url.Parse(c.baseURL)
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return c, errors.New("-base-url must be an http(s) origin, e.g. http://localhost:8080")
	}
	c.userID = strings.TrimSpace(c.userID)
	if c.userID == "" || strings.ContainsAny(c.userID, "/?#") || c.userID == "." || c.userID == ".." {
		return c, errors.New("provide -user-id with the calendar owner's ID from /book/<ID>")
	}
	if _, err := time.Parse(time.DateOnly, c.date); err != nil {
		return c, errors.New("-date must be a valid date in YYYY-MM-DD format")
	}
	if c.users < 1 || c.duration <= 0 || c.timeout <= 0 || c.pause < 0 || c.rampUp < 0 || c.rampUp >= c.duration {
		return c, errors.New("users, duration and timeout must be positive; pause >= 0; 0 <= ramp-up < duration")
	}
	if c.out == "" {
		c.out = filepath.Join("results", "performance", time.Now().Format("20060102-150405.000000000"))
	}
	return c, nil
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	c, err := parseConfig(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, "Configuration error:", err)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = c.users + 10
	transport.MaxIdleConnsPerHost = c.users
	transport.MaxConnsPerHost = c.users
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport: transport,
		Timeout:   c.timeout,
		// A redirect to a different page must not count as a successful page load.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	steps := makeSteps(c)
	fmt.Fprintln(stdout, "Preflight: checking all three pages (excluded from measurements)...")
	if err := preflight(ctx, client, steps); err != nil {
		fmt.Fprintln(stderr, "Preflight failed:", err)
		fmt.Fprintln(stderr, "Check that the service is running, the user ID is correct and the calendar is active.")
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(c.out), 0755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	// Refuse to mix a new run with an earlier experiment.
	if err := os.Mkdir(c.out, 0755); err != nil {
		fmt.Fprintln(stderr, "Choose a new -out directory:", err)
		return 1
	}
	fmt.Fprintf(stdout, "Running: users=%d duration=%s pause=%s ramp-up=%s timeout=%s\n", c.users, c.duration, c.pause, c.rampUp, c.timeout)
	fmt.Fprintln(stdout, "Only GET requests are sent. Ctrl+C stops the test and saves collected results.")
	result, err := runLoad(ctx, c, client, steps)
	if err != nil {
		fmt.Fprintln(stderr, "Measurement error:", err)
		return 1
	}
	report, err := saveReport(c, steps, result)
	if err != nil {
		fmt.Fprintln(stderr, "Report error:", err)
		return 1
	}
	fmt.Fprint(stdout, report)
	if result.interrupted || result.total.failed > 0 || len(result.total.durations) == 0 {
		return 1
	}
	return 0
}

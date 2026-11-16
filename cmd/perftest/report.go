package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type metrics struct {
	count, failed                          int
	errorPercent, rps, min, mean, p50, p95 float64
	p99, max                               float64
}

// Percentiles use the nearest-rank definition, including failed attempts.
func calculate(s statistics, elapsedSeconds float64) metrics {
	m := metrics{count: len(s.durations), failed: s.failed}
	if m.count == 0 {
		return m
	}
	values := append([]float64(nil), s.durations...)
	sort.Float64s(values)
	percentile := func(p float64) float64 { return values[int(math.Ceil(p*float64(len(values))))-1] }
	for _, value := range values {
		m.mean += value / float64(m.count)
	}
	m.min, m.max = values[0], values[len(values)-1]
	m.p50, m.p95, m.p99 = percentile(.50), percentile(.95), percentile(.99)
	m.errorPercent = 100 * float64(m.failed) / float64(m.count)
	if elapsedSeconds > 0 {
		m.rps = float64(m.count) / elapsedSeconds
	}
	return m
}

func saveReport(c config, steps []step, result loadResult) (string, error) {
	seconds := result.finished.Sub(result.started).Seconds()
	file, err := os.Create(filepath.Join(c.out, "summary.csv"))
	if err != nil {
		return "", err
	}
	writer := csv.NewWriter(file)
	_ = writer.Write([]string{"step", "requests", "successful", "failed", "error_percent", "min_ms", "mean_ms", "p50_ms", "p95_ms", "p99_ms", "max_ms", "requests_per_second"})
	var report strings.Builder
	fmt.Fprintln(&report, "\nРезультаты веб-теста производительности")
	fmt.Fprintf(&report, "Начало (UTC): %s\n", result.started.UTC().Format("2006-01-02T15:04:05.000Z"))
	fmt.Fprintf(&report, "Сервис: %s\nПользователь календаря: %s\nДата: %s\n", c.baseURL, c.userID, c.date)
	fmt.Fprintf(&report, "Виртуальные пользователи: %d; окно запуска запросов: %s; пауза: %s; ramp-up: %s; тайм-аут: %s\n", c.users, c.duration, c.pause, c.rampUp, c.timeout)
	fmt.Fprintf(&report, "Фактическое время наблюдения: %.3f с; прерван: %t\n", seconds, result.interrupted)
	fmt.Fprintf(&report, "Генератор: %s, %s/%s, логических CPU: %d\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	fmt.Fprintln(&report, "Сценарий (GET, ожидается HTTP 200 и содержимое нужной страницы):")
	for _, s := range steps {
		fmt.Fprintf(&report, "  %s: %s\n", s.name, s.url)
	}
	fmt.Fprintln(&report)
	table := tabwriter.NewWriter(&report, 0, 4, 2, ' ', 0)
	fmt.Fprintln(table, "Step\tRequests\tFailed\tErrors %\tAvg ms\tp95 ms\tMax ms\tReq/s")
	names := make([]string, 0, len(steps)+1)
	for _, s := range steps {
		names = append(names, s.name)
	}
	names = append(names, "TOTAL")
	for _, name := range names {
		stats := result.byStep[name]
		if name == "TOTAL" {
			stats = &result.total
		}
		m := calculate(*stats, seconds)
		f := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
		_ = writer.Write([]string{name, strconv.Itoa(m.count), strconv.Itoa(m.count - m.failed), strconv.Itoa(m.failed), f(m.errorPercent), f(m.min), f(m.mean), f(m.p50), f(m.p95), f(m.p99), f(m.max), f(m.rps)})
		fmt.Fprintf(table, "%s\t%d\t%d\t%.2f\t%.3f\t%.3f\t%.3f\t%.3f\n", name, m.count, m.failed, m.errorPercent, m.mean, m.p95, m.max, m.rps)
	}
	_ = table.Flush()
	fmt.Fprintln(&report, "\nВремя запроса включает получение тела ответа; проверки содержимого выполняются после измерения.")
	fmt.Fprintln(&report, "Статистика задержек включает успешные и неуспешные попытки; p95/p99: метод ближайшего ранга.")
	fmt.Fprintln(&report, "Req/s = число попыток / фактическое время наблюдения, включая завершение начатых запросов.")
	fmt.Fprintln(&report, "Предварительная проверка исключена. CSS, JavaScript, отрисовка браузера и создание бронирований не измеряются.")
	if len(result.errors) > 0 {
		fmt.Fprintln(&report, "\nОшибки:")
		keys := make([]string, 0, len(result.errors))
		for message := range result.errors {
			keys = append(keys, message)
		}
		sort.Strings(keys)
		for _, message := range keys {
			fmt.Fprintf(&report, "  %d x %s\n", result.errors[message], message)
		}
	}
	absOut, err := filepath.Abs(c.out)
	if err != nil {
		_ = file.Close()
		return "", err
	}
	fmt.Fprintf(&report, "\nФайлы: %s\n  requests.csv — отдельные запросы\n  summary.csv — итоговая таблица\n  report.txt — текст для отчёта\n", absOut)
	writer.Flush()
	writeErr := writer.Error()
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	text := report.String()
	return text, os.WriteFile(filepath.Join(c.out, "report.txt"), []byte(text), 0644)
}

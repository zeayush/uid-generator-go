// Package bench_test runs throughput benchmarks for uid-generator-go.
// Run with: go test -bench=. -benchmem -benchtime=5s ./bench/
package bench_test

import (
	"fmt"
	"testing"
	"time"

	"uid-generator-go/uid"
)

// BenchmarkSnowflake_SingleThread measures single-goroutine Snowflake throughput.
// Target: ≥ 4 000 000 IDs/second.
func BenchmarkSnowflake_SingleThread(b *testing.B) {
	sf, err := uid.NewSnowflake(1, time.Time{})
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.NextID() //nolint:errcheck
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ids/sec")
}

// BenchmarkULID_SingleThread measures single-goroutine ULID throughput.
func BenchmarkULID_SingleThread(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid.NewULID() //nolint:errcheck
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ids/sec")
}

// TestThroughput_4M verifies (informally) that Snowflake hits ≥ 4M IDs/sec.
// This test never fails on its own — it logs the actual throughput so you
// can check the target even without running the full benchmark suite.
func TestThroughput_4M(t *testing.T) {
	sf, err := uid.NewSnowflake(1, time.Time{})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	const target = 4_000_000
	const window = time.Second

	count := 0
	deadline := time.Now().Add(window)
	for time.Now().Before(deadline) {
		sf.NextID() //nolint:errcheck
		count++
	}

	fmt.Printf("Snowflake throughput: %d IDs/sec  (target ≥ %d)\n", count, target)
	if count < target {
		// Don't fail — CI hardware varies widely. Use -bench for authoritative numbers.
		t.Logf("throughput %d < %d target — run `go test -bench=BenchmarkSnowflake_SingleThread -benchtime=5s ./bench/` locally for accurate numbers", count, target)
	}
}

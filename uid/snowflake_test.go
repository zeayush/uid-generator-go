package uid

import (
	"sync"
	"testing"
	"time"
)

func TestNewSnowflake_InvalidMachineID(t *testing.T) {
	_, err := NewSnowflake(-1, time.Time{})
	if err == nil {
		t.Fatal("expected error for negative machineID, got nil")
	}

	_, err = NewSnowflake(MaxMachineID+1, time.Time{})
	if err == nil {
		t.Fatalf("expected error for machineID %d (> %d), got nil", MaxMachineID+1, MaxMachineID)
	}
}

func TestNewSnowflake_DefaultEpoch(t *testing.T) {
	sf, err := NewSnowflake(1, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sf.epoch != DefaultEpoch {
		t.Errorf("expected epoch %v, got %v", DefaultEpoch, sf.epoch)
	}
}

func TestNewSnowflake_CustomEpoch(t *testing.T) {
	custom := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	sf, err := NewSnowflake(0, custom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sf.epoch != custom {
		t.Errorf("expected epoch %v, got %v", custom, sf.epoch)
	}
}

func TestNextID_Unique(t *testing.T) {
	sf, _ := NewSnowflake(1, time.Time{})
	const n = 10_000
	seen := make(map[int64]bool, n)
	for i := 0; i < n; i++ {
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("unexpected error at i=%d: %v", i, err)
		}
		if seen[id] {
			t.Fatalf("duplicate ID at i=%d: %d", i, id)
		}
		seen[id] = true
	}
}

func TestNextID_Monotonic(t *testing.T) {
	sf, _ := NewSnowflake(1, time.Time{})
	prev, _ := sf.NextID()
	for i := 0; i < 5_000; i++ {
		id, _ := sf.NextID()
		if id <= prev {
			t.Fatalf("ID not monotonically increasing at i=%d: %d <= %d", i, id, prev)
		}
		prev = id
	}
}

func TestNextID_ConcurrentUnique(t *testing.T) {
	sf, _ := NewSnowflake(1, time.Time{})
	const n = 10_000
	ids := make([]int64, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id, _ := sf.NextID()
			ids[i] = id
		}()
	}
	wg.Wait()

	seen := make(map[int64]bool, n)
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate ID in concurrent test: %d", id)
		}
		seen[id] = true
	}
}

func TestNextID_SequenceExhaustion(t *testing.T) {
	// Generate more than MaxSequence IDs as fast as possible.
	// At some point the sequence overflows and the generator must park until
	// the next millisecond. All IDs must remain unique and monotonically increasing.
	sf, _ := NewSnowflake(1, time.Time{})
	const n = int(MaxSequence) * 3 // spans at least two sequence windows (~12 285 IDs)
	prev, _ := sf.NextID()
	seen := make(map[int64]bool, n)
	seen[prev] = true
	for i := 1; i < n; i++ {
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("error at i=%d: %v", i, err)
		}
		if id <= prev {
			t.Fatalf("non-monotonic at i=%d: %d <= %d", i, id, prev)
		}
		if seen[id] {
			t.Fatalf("duplicate at i=%d: %d", i, id)
		}
		seen[id] = true
		prev = id
	}
}

func TestDecomposeID(t *testing.T) {
	const wantMachine int64 = 42
	sf, _ := NewSnowflake(wantMachine, time.Time{})
	before := time.Now()
	id, _ := sf.NextID()
	after := time.Now()

	ts, machine, seq := DecomposeID(id, DefaultEpoch)

	if machine != wantMachine {
		t.Errorf("machineID: got %d, want %d", machine, wantMachine)
	}
	if seq < 0 || seq > MaxSequence {
		t.Errorf("sequence out of range [0, %d]: %d", MaxSequence, seq)
	}
	lo := before.Truncate(time.Millisecond)
	hi := after.Add(time.Millisecond)
	if ts.Before(lo) || ts.After(hi) {
		t.Errorf("timestamp %v outside expected window [%v, %v]", ts, lo, hi)
	}
}

func BenchmarkNextID(b *testing.B) {
	sf, _ := NewSnowflake(1, time.Time{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.NextID() //nolint:errcheck
	}
}

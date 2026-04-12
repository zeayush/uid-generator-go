// Package uid provides Snowflake and ULID unique ID generators.
package uid

import (
	"errors"
	"sync"
	"time"
)

// ─── Bit-layout constants ────────────────────────────────────────────────────
//
// A 64-bit Snowflake ID is composed as follows (bit 63 is the sign bit):
//
//	 63        22        12        0
//	 ┌──────────┬─────────┬────────┐
//	 │  41-bit  │ 10-bit  │ 12-bit │
//	 │ timestamp│machine  │sequence│
//	 └──────────┴─────────┴────────┘
//
// timestamp : milliseconds elapsed since the configured epoch (lasts ~69 years)
// machine   : unique ID for the generator instance  (max 1 023 generators)
// sequence  : counter that resets each millisecond  (max 4 095 per ms → ≥4M/s)
const (
	timestampBits = 41
	machineIDBits = 10
	sequenceBits  = 12

	// MaxMachineID is the largest valid machine ID (2^10 − 1 = 1 023).
	MaxMachineID int64 = (1 << machineIDBits) - 1
	// MaxSequence is the largest valid sequence number (2^12 − 1 = 4 095).
	MaxSequence int64 = (1 << sequenceBits) - 1

	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
)

// DefaultEpoch is 2020-01-01 00:00:00 UTC.
// IDs store milliseconds elapsed since this date, so IDs remain valid well
// beyond the year 2089 even with a 41-bit timestamp.
var DefaultEpoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// Snowflake generates unique 64-bit IDs.
// It is safe for concurrent use via the internal mutex.
type Snowflake struct {
	mu        sync.Mutex
	epoch     time.Time
	machineID int64
	sequence  int64
	lastMS    int64
}

// NewSnowflake creates a Snowflake generator.
//
//   - machineID must be in [0, MaxMachineID].
//   - Pass a zero time.Time{} to use DefaultEpoch.
func NewSnowflake(machineID int64, epoch time.Time) (*Snowflake, error) {
	if machineID < 0 || machineID > MaxMachineID {
		return nil, errors.New("machineID out of range [0, MaxMachineID]")
	}
	if epoch.IsZero() {
		epoch = DefaultEpoch
	}
	return &Snowflake{epoch: epoch, machineID: machineID}, nil
}

// currentMS returns the number of milliseconds elapsed since s.epoch.
// Must only be called while holding s.mu (or from within NextID).
func (s *Snowflake) currentMS() int64 {
	return time.Since(s.epoch).Milliseconds()
}

// waitNextMS spins until the clock advances past last, and returns the new ms.
// Call this when the sequence counter overflows MaxSequence.
func (s *Snowflake) waitNextMS(last int64) int64 {
	for {
		if ms := s.currentMS(); ms > last {
			return ms
		}
	}
}

// NextID returns the next globally unique 64-bit Snowflake ID.
// It is safe for concurrent use.
func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.currentMS()

	// Clock drift protection: if the wall clock moved backward, park until we
	// observe a millisecond strictly greater than the last emitted one.
	if now < s.lastMS {
		now = s.waitNextMS(s.lastMS)
	}

	if now == s.lastMS {
		s.sequence++
		if s.sequence > MaxSequence {
			// sequence exhausted — park until the clock ticks forward
			now = s.waitNextMS(s.lastMS)
			s.sequence = 0
			s.lastMS = now
		}
	} else {
		s.sequence = 0
		s.lastMS = now
	}

	id := (now << timestampShift) | (s.machineID << machineIDShift) | s.sequence
	return id, nil
}

// DecomposeID splits a Snowflake ID back into its three components.
// Pass the same epoch that was given to NewSnowflake.
func DecomposeID(id int64, epoch time.Time) (ts time.Time, machineID int64, sequence int64) {
	sequence = id & MaxSequence
	machineID = (id >> machineIDShift) & MaxMachineID
	tsMS := id >> timestampShift
	ts = epoch.Add(time.Duration(tsMS) * time.Millisecond)
	return ts, machineID, sequence
}

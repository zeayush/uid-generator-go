package uid

import (
	"strings"
	"testing"
	"time"
)

func TestNewULID_Length(t *testing.T) {
	u, err := NewULID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s := u.String(); len(s) != 26 {
		t.Errorf("expected length 26, got %d: %q", len(s), s)
	}
}

func TestNewULID_ValidAlphabet(t *testing.T) {
	u, err := NewULID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range u.String() {
		if !strings.ContainsRune(crockfordAlphabet, c) {
			t.Errorf("invalid character %q at position %d", c, i)
		}
	}
}

func TestULID_TimestampRoundtrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	u, err := NewULIDWithTime(now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := u.Time()
	if !got.Equal(now) {
		t.Errorf("time mismatch: got %v, want %v", got, now)
	}
}

func TestParseULID_Roundtrip(t *testing.T) {
	u, err := NewULID()
	if err != nil {
		t.Fatalf("NewULID error: %v", err)
	}
	s := u.String()
	parsed, err := ParseULID(s)
	if err != nil {
		t.Fatalf("ParseULID(%q) error: %v", s, err)
	}
	if u != parsed {
		t.Errorf("roundtrip mismatch:\n got  %v\n want %v", parsed, u)
	}
}

func TestParseULID_InvalidLength(t *testing.T) {
	cases := []string{"", "TOOSHORT", "THIS-STRING-IS-WAY-TOO-LONG-FOR-A-VALID-ULID"}
	for _, tc := range cases {
		_, err := ParseULID(tc)
		if err == nil {
			t.Errorf("expected error for %q, got nil", tc)
		}
	}
}

func TestULID_LexicographicOrder(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := t1.Add(1 * time.Millisecond)
	u1, err1 := NewULIDWithTime(t1)
	u2, err2 := NewULIDWithTime(t2)
	if err1 != nil || err2 != nil {
		t.Fatalf("NewULIDWithTime errors: %v %v", err1, err2)
	}

	if u1.Compare(u2) >= 0 {
		t.Errorf("expected u1 < u2 by Compare, got %d", u1.Compare(u2))
	}
	s1, s2 := u1.String(), u2.String()
	if s1 >= s2 {
		t.Errorf("string order wrong: %q >= %q", s1, s2)
	}
}

func TestULID_Unique(t *testing.T) {
	seen := make(map[ULID]bool, 1000)
	for i := 0; i < 1000; i++ {
		u, err := NewULID()
		if err != nil {
			t.Fatalf("unexpected error at i=%d: %v", i, err)
		}
		if seen[u] {
			t.Fatalf("duplicate ULID at i=%d", i)
		}
		seen[u] = true
	}
}

func TestDecodeBase32Byte_OCRSubstitutions(t *testing.T) {
	// Crockford alphabet: 0123456789ABCDEFGHJKMNPQRSTVWXYZ
	// Index:              0         1         2         3
	//                     0123456789012345678901234567890 1
	cases := []struct {
		in   byte
		want byte
		ok   bool
	}{
		// digits
		{'0', 0, true},
		{'9', 9, true},
		// uppercase letters (skipping I=18,L=21,O=24,U=30 positions)
		{'A', 10, true},
		{'Z', 31, true},
		// lowercase aliases
		{'a', 10, true},
		{'z', 31, true},
		// OCR aliases
		{'O', 0, true},  // O → 0
		{'o', 0, true},
		{'I', 1, true},  // I → 1
		{'i', 1, true},
		{'L', 1, true},  // L → 1
		{'l', 1, true},
		// invalid
		{'U', 0, false},
		{'u', 0, false},
		{'!', 0, false},
	}
	for _, tc := range cases {
		got, ok := decodeBase32Byte(tc.in)
		if ok != tc.ok {
			t.Errorf("decodeBase32Byte(%q): ok=%v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok && got != tc.want {
			t.Errorf("decodeBase32Byte(%q): got %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestULID_TimestampOverflow(t *testing.T) {
	// A timestamp larger than 2^48 - 1 must be rejected.
	overflow := time.UnixMilli(1<<48).UTC()
	_, err := NewULIDWithTime(overflow)
	if err == nil {
		t.Error("expected error for timestamp that overflows 48 bits, got nil")
	}
}

func BenchmarkNewULID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewULID() //nolint:errcheck
	}
}

package uid

import (
	"crypto/rand"
	"errors"
	"time"
)

// crockfordAlphabet is the 32-character Crockford Base32 encoding alphabet.
// It omits I, L, O, U to reduce transcription errors (e.g. 0 vs O, 1 vs I/L).
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ULID is a 128-bit Universally Unique Lexicographically Sortable Identifier.
//
// Raw byte layout (big-endian):
//
//	bytes  0-5  :  48-bit millisecond timestamp since Unix epoch
//	bytes 6-15  :  80-bit cryptographically random data
//
// The String() representation is exactly 26 Crockford Base32 characters.
// Two ULIDs generated in different milliseconds sort lexicographically by time.
type ULID [16]byte

// NewULID creates a ULID using the current UTC time.
func NewULID() (ULID, error) {
	return NewULIDWithTime(time.Now().UTC())
}

// NewULIDWithTime creates a ULID for a specific time (useful in tests to
// produce a predictable timestamp while keeping the random portion random).
func NewULIDWithTime(t time.Time) (ULID, error) {
	// TODO:
	//  1. ms := t.UnixMilli()
	//  2. Validate: ms must fit in 48 bits.
	//       const maxMS = (1 << 48) - 1
	//       if ms < 0 || ms > maxMS { return ULID{}, errors.New("timestamp overflows 48 bits") }
	//  3. Encode ms into bytes 0–5 (big-endian):
	//       var id ULID
	//       id[0] = byte(ms >> 40)
	//       id[1] = byte(ms >> 32)
	//       id[2] = byte(ms >> 24)
	//       id[3] = byte(ms >> 16)
	//       id[4] = byte(ms >> 8)
	//       id[5] = byte(ms)
	//  4. Fill bytes 6–15 with 10 cryptographically random bytes:
	//       if err := randomBytes(id[6:]); err != nil { return ULID{}, err }
	//  5. Return id, nil

	return ULID{}, errors.New("not implemented")
}

// String encodes the ULID as a 26-character Crockford Base32 string.
//
// Each output character represents exactly 5 bits of the 128-bit ULID value.
// The full bit-assignment reference (from the oklog/ulid spec):
//
//	dst[ 0] = crockfordAlphabet[(u[0]&0xE0)>>5]
//	dst[ 1] = crockfordAlphabet[u[0]&0x1F]
//	dst[ 2] = crockfordAlphabet[(u[1]&0xF8)>>3]
//	dst[ 3] = crockfordAlphabet[((u[1]&0x07)<<2)|((u[2]&0xC0)>>6)]
//	dst[ 4] = crockfordAlphabet[(u[2]&0x3E)>>1]
//	dst[ 5] = crockfordAlphabet[((u[2]&0x01)<<4)|((u[3]&0xF0)>>4)]
//	dst[ 6] = crockfordAlphabet[((u[3]&0x0F)<<1)|((u[4]&0x80)>>7)]
//	dst[ 7] = crockfordAlphabet[(u[4]&0x7C)>>2]
//	dst[ 8] = crockfordAlphabet[((u[4]&0x03)<<3)|((u[5]&0xE0)>>5)]
//	dst[ 9] = crockfordAlphabet[u[5]&0x1F]
//	dst[10] = crockfordAlphabet[(u[6]&0xF8)>>3]
//	... (pattern continues for bytes 6–15, producing dst[10]–dst[25])
//	dst[25] = crockfordAlphabet[u[15]&0x1F]
func (u ULID) String() string {
	// TODO:
	//  Declare var dst [26]byte, fill each position using the bit layout above,
	//  then return string(dst[:]).
	return ""
}

// ParseULID decodes a 26-character Crockford Base32 string into a ULID.
func ParseULID(s string) (ULID, error) {
	// TODO:
	//  1. if len(s) != 26 { return ULID{}, errors.New("invalid ULID: must be 26 characters") }
	//  2. Reverse the String() encoding:
	//       - for each output character call decodeBase32Byte(s[i])
	//       - pack the resulting 5-bit values back into [16]byte (inverse of String())
	//  3. Return ULID(bytes), nil
	_ = s
	return ULID{}, errors.New("not implemented")
}

// Time extracts the millisecond timestamp embedded in the ULID.
func (u ULID) Time() time.Time {
	// TODO: reconstruct ms from the first 6 bytes and return time.UnixMilli(ms).UTC()
	//   ms = int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 |
	//        int64(u[3])<<16 | int64(u[4])<<8  | int64(u[5])
	return time.Time{}
}

// Compare returns -1, 0, or +1 when u is less than, equal to, or greater than v.
// Comparison is byte-by-byte on the raw [16]byte, which equals time-then-random ordering.
func (u ULID) Compare(v ULID) int {
	// TODO: compare u[:] and v[:] byte-by-byte.
	//   Return -1 if u < v, 0 if equal, +1 if u > v.
	return 0
}

// decodeBase32Byte maps a single Crockford Base32 character to its 5-bit value.
// Returns (value, true) on success, (0, false) for invalid input.
//
// Crockford alias/substitution rules:
//
//	'0'-'9'           → 0-9
//	'A'-'Z' (no I,L,O,U) → 10-31   (see crockfordAlphabet for the exact mapping)
//	lowercase         → same as uppercase
//	'I', 'i', 'L', 'l' → 1         (OCR-confusion alias)
//	'O', 'o'          → 0           (OCR-confusion alias)
//	'U', 'u'          → invalid
func decodeBase32Byte(c byte) (byte, bool) {
	// TODO: implement the lookup described above.
	_ = c
	return 0, false
}

// randomBytes fills b with cryptographically secure random bytes.
// Declared as a package-level variable so tests can substitute a deterministic source.
var randomBytes = func(b []byte) error {
	_, err := rand.Read(b)
	return err
}

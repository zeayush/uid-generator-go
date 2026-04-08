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
	ms := t.UnixMilli()
	const maxMS = (1 << 48) - 1
	if ms < 0 || ms > maxMS {
		return ULID{}, errors.New("timestamp overflows 48 bits")
	}
	var id ULID
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	if err := randomBytes(id[6:]); err != nil {
		return ULID{}, err
	}
	return id, nil
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
	var dst [26]byte
	dst[0]  = crockfordAlphabet[(u[0]&0xE0)>>5]
	dst[1]  = crockfordAlphabet[u[0]&0x1F]
	dst[2]  = crockfordAlphabet[(u[1]&0xF8)>>3]
	dst[3]  = crockfordAlphabet[((u[1]&0x07)<<2)|((u[2]&0xC0)>>6)]
	dst[4]  = crockfordAlphabet[(u[2]&0x3E)>>1]
	dst[5]  = crockfordAlphabet[((u[2]&0x01)<<4)|((u[3]&0xF0)>>4)]
	dst[6]  = crockfordAlphabet[((u[3]&0x0F)<<1)|((u[4]&0x80)>>7)]
	dst[7]  = crockfordAlphabet[(u[4]&0x7C)>>2]
	dst[8]  = crockfordAlphabet[((u[4]&0x03)<<3)|((u[5]&0xE0)>>5)]
	dst[9]  = crockfordAlphabet[u[5]&0x1F]
	dst[10] = crockfordAlphabet[(u[6]&0xF8)>>3]
	dst[11] = crockfordAlphabet[((u[6]&0x07)<<2)|((u[7]&0xC0)>>6)]
	dst[12] = crockfordAlphabet[(u[7]&0x3E)>>1]
	dst[13] = crockfordAlphabet[((u[7]&0x01)<<4)|((u[8]&0xF0)>>4)]
	dst[14] = crockfordAlphabet[((u[8]&0x0F)<<1)|((u[9]&0x80)>>7)]
	dst[15] = crockfordAlphabet[(u[9]&0x7C)>>2]
	dst[16] = crockfordAlphabet[((u[9]&0x03)<<3)|((u[10]&0xE0)>>5)]
	dst[17] = crockfordAlphabet[u[10]&0x1F]
	dst[18] = crockfordAlphabet[(u[11]&0xF8)>>3]
	dst[19] = crockfordAlphabet[((u[11]&0x07)<<2)|((u[12]&0xC0)>>6)]
	dst[20] = crockfordAlphabet[(u[12]&0x3E)>>1]
	dst[21] = crockfordAlphabet[((u[12]&0x01)<<4)|((u[13]&0xF0)>>4)]
	dst[22] = crockfordAlphabet[((u[13]&0x0F)<<1)|((u[14]&0x80)>>7)]
	dst[23] = crockfordAlphabet[(u[14]&0x7C)>>2]
	dst[24] = crockfordAlphabet[((u[14]&0x03)<<3)|((u[15]&0xE0)>>5)]
	dst[25] = crockfordAlphabet[u[15]&0x1F]
	return string(dst[:])
}

// ParseULID decodes a 26-character Crockford Base32 string into a ULID.
func ParseULID(s string) (ULID, error) {
	if len(s) != 26 {
		return ULID{}, errors.New("invalid ULID: must be 26 characters")
	}
	var v [26]byte
	for i := 0; i < 26; i++ {
		b, ok := decodeBase32Byte(s[i])
		if !ok {
			return ULID{}, errors.New("invalid ULID: invalid character")
		}
		v[i] = b
	}
	var id ULID
	id[0]  = (v[0] << 5) | v[1]
	id[1]  = (v[2] << 3) | (v[3] >> 2)
	id[2]  = ((v[3] & 0x03) << 6) | (v[4] << 1) | (v[5] >> 4)
	id[3]  = ((v[5] & 0x0F) << 4) | (v[6] >> 1)
	id[4]  = ((v[6] & 0x01) << 7) | (v[7] << 2) | (v[8] >> 3)
	id[5]  = ((v[8] & 0x07) << 5) | v[9]
	id[6]  = (v[10] << 3) | (v[11] >> 2)
	id[7]  = ((v[11] & 0x03) << 6) | (v[12] << 1) | (v[13] >> 4)
	id[8]  = ((v[13] & 0x0F) << 4) | (v[14] >> 1)
	id[9]  = ((v[14] & 0x01) << 7) | (v[15] << 2) | (v[16] >> 3)
	id[10] = ((v[16] & 0x07) << 5) | v[17]
	id[11] = (v[18] << 3) | (v[19] >> 2)
	id[12] = ((v[19] & 0x03) << 6) | (v[20] << 1) | (v[21] >> 4)
	id[13] = ((v[21] & 0x0F) << 4) | (v[22] >> 1)
	id[14] = ((v[22] & 0x01) << 7) | (v[23] << 2) | (v[24] >> 3)
	id[15] = ((v[24] & 0x07) << 5) | v[25]
	return id, nil
}

// Time extracts the millisecond timestamp embedded in the ULID.
func (u ULID) Time() time.Time {
	ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 |
		int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return time.UnixMilli(ms).UTC()
}

// Compare returns -1, 0, or +1 when u is less than, equal to, or greater than v.
// Comparison is byte-by-byte on the raw [16]byte, which equals time-then-random ordering.
func (u ULID) Compare(v ULID) int {
	for i := 0; i < 16; i++ {
		if u[i] < v[i] {
			return -1
		}
		if u[i] > v[i] {
			return 1
		}
	}
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
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c == 'O' || c == 'o':
		return 0, true
	case c == 'I' || c == 'i' || c == 'L' || c == 'l':
		return 1, true
	case c >= 'A' && c <= 'Z':
		// Map uppercase letters: ABCDEFGHJKMNPQRSTVWXYZ -> 10..31
		// The alphabet skips I(18), L(21), O(24), U(29) (0-indexed from 'A')
		offset := c - 'A'
		lookup := [26]byte{10, 11, 12, 13, 14, 15, 16, 17, 0, 18, 19, 0, 20, 21, 0, 22, 23, 24, 25, 26, 0, 27, 28, 29, 30, 31}
		val := lookup[offset]
		if val == 0 && c != 'A' {
			return 0, false // I, L, O, U are invalid (already handled above for O/I/L)
		}
		return val, true
	case c >= 'a' && c <= 'z':
		return decodeBase32Byte(c - 32)
	default:
		return 0, false
	}
}

// randomBytes fills b with cryptographically secure random bytes.
// Declared as a package-level variable so tests can substitute a deterministic source.
var randomBytes = func(b []byte) error {
	_, err := rand.Read(b)
	return err
}

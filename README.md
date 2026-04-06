# uid-generator-go

![CI](https://github.com/zeayush/uid-generator-go/actions/workflows/go-ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)

Two production-grade unique ID generators implemented from scratch in Go — a **Snowflake** 64-bit integer generator and a **ULID** 128-bit lexicographically sortable identifier.

> **Learning exercise** — function bodies are left as `TODO` stubs for you to implement. Tests are pre-written and pass once your implementations are correct. Once complete, this is a production-ready library.

---

## What are these?

**Snowflake** IDs were invented by Twitter in 2010 to replace database auto-increment in distributed systems. A single 64-bit integer embeds a millisecond timestamp, a machine ID, and a per-millisecond sequence counter — so IDs generated across thousands of machines are unique without any coordination or central authority.

**ULID** (Universally Unique Lexicographically Sortable Identifier) solves a different problem: UUID is globally unique but sorts randomly in indexes, causing B-tree fragmentation and poor write locality. ULID packs a 48-bit timestamp into the high bits so string-sorted order equals time-sorted order.

---

## How it Works

### Snowflake — bit packing

```
  63        22        12        0
  +----------+---------+--------+
  |  41-bit  | 10-bit  | 12-bit |
  | timestamp| machine |sequence|
  +----------+---------+--------+

  epoch: 2020-01-01 00:00:00 UTC  (configurable)
```

- **timestamp** — milliseconds since the custom epoch. 41 bits = ~69 years before rollover (valid past 2089).
- **machine ID** — unique per generator instance (0-1023). Resolved from env var, IP hash, or hostname hash.
- **sequence** — wraps at 4095 per millisecond; generator parks (busy-waits) until the next ms tick when exhausted.

Assembly and decomposition:

```
id = (timestampMS << 22) | (machineID << 12) | sequence
```

### ULID — timestamp prefix + entropy suffix

```
  01ARZ3NDEKTSV4RRFFQ69G5FAV
  +----------++-------------+
   10 chars    16 chars
   48-bit ms   80-bit random
```

128 raw bits are encoded as 26 Crockford Base32 characters (omits I L O U to prevent transcription errors). Two ULIDs from different milliseconds sort correctly as plain strings — no index tricks required.

---

## Quick Start

#### Snowflake

```go
import "uid-generator-go/uid"

mid, _ := uid.ResolveMachineID()            // env -> IP -> hostname
sf, _ := uid.NewSnowflake(mid, time.Time{}) // zero epoch -> DefaultEpoch (2020-01-01)

id, err := sf.NextID()                      // int64 -- monotonically increasing
ts, machine, seq := uid.DecomposeID(id)     // inspect each field
```

#### ULID

```go
u, err := uid.NewULID()     // uses time.Now()
fmt.Println(u)              // "01HXK7P9ZQABCDEFGHJKMNPQ" -- 26 chars

u2, _ := uid.ParseULID("01HXK7P9ZQABCDEFGHJKMNPQ")
t := u2.Time()              // time.Time, millisecond precision
cmp := u.Compare(u2)        // -1 / 0 / +1 -- same contract as bytes.Compare
```

#### Machine ID

```go
// Automatic resolution (recommended)
mid, err := uid.ResolveMachineID()

// Explicit sources
mid, err = uid.MachineIDFromEnv()      // reads UID_MACHINE_ID env var
mid, err = uid.MachineIDFromIP()       // FNV-1a hash of primary IPv4
mid, err = uid.MachineIDFromHostname() // FNV-1a hash of os.Hostname()
```

---

## API

### Snowflake

```go
var DefaultEpoch time.Time                           // 2020-01-01 00:00:00 UTC
const MaxMachineID int64 = 1023
const MaxSequence  int64 = 4095

func NewSnowflake(machineID int64, epoch time.Time) (*Snowflake, error)
func (s *Snowflake) NextID() (int64, error)
func DecomposeID(id int64) (timestampMS, machineID, sequence int64)
```

`NextID` is safe for concurrent use. `NewSnowflake` returns an error for `machineID` outside `[0, 1023]`.

### ULID

```go
func NewULID() (ULID, error)
func NewULIDWithTime(t time.Time) (ULID, error)
func ParseULID(s string) (ULID, error)

func (u ULID) String() string          // 26-char Crockford Base32
func (u ULID) Time() time.Time         // millisecond precision
func (u ULID) Compare(other ULID) int  // -1 / 0 / +1
```

`ParseULID` accepts upper- and lower-case input and maps I->1, L->1, O->0 per the Crockford spec.

### Machine ID

```go
const MachineIDEnvVar = "UID_MACHINE_ID"

func MachineIDFromEnv() (int64, error)
func MachineIDFromHostname() (int64, error)
func MachineIDFromIP() (int64, error)
func ResolveMachineID() (int64, error)   // tries sources in priority order
```

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| Custom epoch (2020-01-01) | Shifts the 41-bit counter forward -- IDs stay valid past 2089 vs 2039 with a Unix epoch |
| sync.Mutex on NextID | Simpler than atomics; fast enough for any realistic single-generator throughput |
| Busy-wait on sequence exhaustion | Avoids time.Sleep jitter -- the spin completes in < 1 ms in practice |
| crypto/rand for ULID entropy | OS entropy eliminates all birthday-collision risk at realistic throughputs |
| FNV-1a 32-bit for machine ID hash | No external dependencies, deterministic, hardware-friendly |
| Crockford Base32 for ULID encoding | 5 bits/character is maximally efficient; omitting I L O U prevents hand-transcription errors |
| time.Time epoch, not int64 | Readable API -- callers pass time.Date(2024, 1, 1, ...) instead of a raw ms offset |

---

## Benchmarks

```bash
go test -bench=. -benchmem -benchtime=5s ./bench/
```

Target performance (single-threaded, post-implementation):

| Benchmark | Target |
|---|---|
| BenchmarkSnowflake_SingleThread | >= 4 000 000 IDs/sec |
| BenchmarkSnowflake_Parallel | scales with goroutines (mutex-bound) |
| BenchmarkULID_SingleThread | >= 1 000 000 IDs/sec |
| BenchmarkULID_Parallel | scales with goroutines |

The Snowflake ceiling of 4 096 000 IDs/sec is architectural (12-bit sequence field) -- not a code quality issue. The ULID ceiling is bounded by crypto/rand throughput.

---

## Tests

```bash
go test -v -race ./...
```

**Snowflake** (10 tests): invalid machine ID, default epoch, custom epoch, 10k unique IDs, 5k monotonic IDs, 10k concurrent unique IDs, machine ID embedded in output, decompose round-trip, sequence overflow parking, custom epoch timestamp field.

**ULID** (9 tests): 26-char length, 1k unique IDs, sort order, parse round-trip, invalid length, invalid char, time round-trip, Compare ordering, case-insensitive parse.

**Machine ID** (5 tests): valid env var, unset env var error, out-of-range error, hostname in-range, fallthrough to hostname when env unset.

---

## Project Structure

```
uid-generator-go/
+-- uid/
|   +-- snowflake.go        <- Snowflake generator  (your implementation)
|   +-- snowflake_test.go   <- Snowflake tests      (pre-written)
|   +-- ulid.go             <- ULID generator       (your implementation)
|   +-- ulid_test.go        <- ULID tests           (pre-written)
|   +-- machineid.go        <- Machine ID sources   (your implementation)
|   +-- machineid_test.go   <- Machine ID tests     (pre-written)
+-- bench/
|   +-- bench_test.go       <- throughput benchmarks
+-- .github/
|   +-- workflows/
|       +-- go-ci.yml       <- CI: test + race + bench smoke
+-- go.mod
+-- README.md
```

---

## Implementation Guide

Work through the files in this order -- each step builds on the previous one.

### Step 1 -- uid/machineid.go

Start here because Snowflake depends on a valid machine ID.

1. `MachineIDFromEnv` -- parse with `strconv.ParseInt`, validate `[0, MaxMachineID]`.
2. `MachineIDFromHostname` -- `int64(h.Sum32()) & MaxMachineID` using `hash/fnv.New32a()`.
3. `MachineIDFromIP` -- iterate `net.InterfaceAddrs()`, hash the first non-loopback `net.IP.To4()`.
4. `ResolveMachineID` -- call each source in priority order, return the first non-error result.

### Step 2 -- uid/snowflake.go

1. `NewSnowflake` -- validate `machineID`, apply `DefaultEpoch` when `epoch.IsZero()`.
2. `currentMS` -- `time.Since(s.epoch).Milliseconds()`.
3. `waitNextMS` -- spin calling `currentMS()` until the result exceeds `last`.
4. `NextID` -- implement the four-step assembly (clock drift -> same-ms -> new-ms -> shift-or compose).
5. `DecomposeID` -- reverse the shifts with right-shifts and bit masks.

### Step 3 -- uid/ulid.go

1. `NewULIDWithTime` -- validate 48-bit range, big-endian encode ms into bytes 0-5, fill bytes 6-15 with `crypto/rand`.
2. `ULID.String` -- pack 128 bits into 26 x 5-bit Crockford characters (bit table is in the source comment).
3. `ParseULID` -- validate length, map each character through `decodeBase32Byte`, reassemble the byte array.
4. `ULID.Time` -- reconstruct ms from bytes 0-5 (reverse of the encoding step).
5. `ULID.Compare` -- `bytes.Compare` works directly on the `[16]byte` array.
6. `decodeBase32Byte` -- Crockford table with aliases I->1, L->1, O->0.

---

## Machine ID Resolution Order

| Priority | Source | Mechanism |
|---|---|---|
| 1 | UID_MACHINE_ID env var | Explicit integer override -- use this in production |
| 2 | Primary non-loopback IPv4 | FNV-1a hash of the raw 4-byte IP address |
| 3 | os.Hostname() | FNV-1a hash of the hostname string |

Set `UID_MACHINE_ID=<n>` to guarantee the same machine ID across restarts and redeployments.

---

## References

- [Announcing Snowflake -- Twitter Engineering Blog (2010)](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake)
- [ULID Specification -- github.com/ulid/spec](https://github.com/ulid/spec)
- [Crockford Base32 -- crockford.com](https://www.crockford.com/base32.html)
- [FNV Hash -- Fowler, Noll, Vo](http://www.isthe.com/chongo/tech/comp/fnv/)
- [oklog/ulid -- reference ULID implementation in Go](https://github.com/oklog/ulid)
- [sony/sonyflake -- production Snowflake variant in Go](https://github.com/sony/sonyflake)
- [System Design Interview Vol. 1, Ch. 7 -- Design a Unique ID Generator in Distributed Systems](https://www.amazon.com/System-Design-Interview-insiders-Second/dp/B08CMF2CQF)

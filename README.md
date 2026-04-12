# uid-generator-go

[![CI](https://github.com/zeayush/uid-generator-go/actions/workflows/go-ci.yml/badge.svg)](https://github.com/zeayush/uid-generator-go/actions/workflows/go-ci.yml)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

Unique ID generation in Go — API-consumer port of the Rust implementation,
including Snowflake-style 64-bit IDs and ULID.

Part of a distributed systems portfolio implementing every system from **Alex
Xu's System Design Interview (Vol. 1 & 2)**. This covers **Chapter 7 —
Design a Unique ID Generator in Distributed Systems**.

---

## What It Provides

- **Snowflake 64-bit IDs**: 41-bit timestamp, 10-bit machine ID, 12-bit sequence
- **ULID**: 48-bit timestamp + 80-bit randomness, lexicographically sortable
- **Custom epoch**: configurable start date
- **Machine ID resolution**: env var, IP hash, or hostname hash
- **Clock drift protection**: backward clock movement parks until safe time
- **Sequence exhaustion handling**: waits for next millisecond after 4096 IDs/ms

---

## Quick Start

```sh
go get github.com/zeayush/uid-generator-go
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/zeayush/uid-generator-go/uid"
)

func main() {
	machineID, _ := uid.ResolveMachineID()
	sf, _ := uid.NewSnowflake(machineID, time.Time{})

	id, _ := sf.NextID()
	ts, mid, seq := uid.DecomposeID(id, uid.DefaultEpoch)

	u, _ := uid.NewULID()
	parsed, _ := uid.ParseULID(u.String())

	fmt.Println(id, ts, mid, seq, parsed)
}
```

---

## API

```go
var DefaultEpoch time.Time
const MaxMachineID int64 = 1023
const MaxSequence int64 = 4095

func NewSnowflake(machineID int64, epoch time.Time) (*Snowflake, error)
func (s *Snowflake) NextID() (int64, error)
func DecomposeID(id int64, epoch time.Time) (time.Time, int64, int64)

func NewULID() (ULID, error)
func NewULIDWithTime(t time.Time) (ULID, error)
func ParseULID(s string) (ULID, error)
func (u ULID) String() string
func (u ULID) Time() time.Time
func (u ULID) Compare(other ULID) int

func MachineIDFromEnv() (int64, error)
func MachineIDFromIP() (int64, error)
func MachineIDFromHostname() (int64, error)
func ResolveMachineID() (int64, error)
```

---

## Benchmarks

```sh
go test -bench=. -benchmem ./bench
```

Current throughput benchmark output on Apple M1:

| Benchmark | Result |
|---|---|
| `BenchmarkSnowflake_SingleThread` | ~4.09M IDs/sec |
| `BenchmarkULID_SingleThread` | ~4.65M IDs/sec |

Snowflake has a hard architectural ceiling of $4096 \times 1000 = 4{,}096{,}000$
IDs/sec from its 12-bit sequence field.

---

## Tests

```sh
go test -v -race ./...
```

Coverage includes uniqueness, monotonicity, concurrent generation, sequence
overflow handling, ULID parse and ordering behavior, and machine ID resolution
fallback order.

---

## Project Criteria

| Item | Status |
|---|---|
| Duration | Scoped to a 2-week build window |
| Language rationale | Rust for hot paths, Go port for consumer APIs |
| Goal | OSS library for pkg.go.dev and internal service reuse |
| Stack | Go 1.22, benchmark package, GitHub Actions, Rust parity |
| Deliverable shape | Publish-ready module + docs + benchmark evidence |
| Feature set | Snowflake + ULID + custom epoch + machine ID fallback + drift/overflow handling |

---

## References

- [Announcing Snowflake -- Twitter Engineering Blog (2010)](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake)
- [ULID Specification -- github.com/ulid/spec](https://github.com/ulid/spec)
- [Crockford Base32 -- crockford.com](https://www.crockford.com/base32.html)
- [FNV Hash -- Fowler, Noll, Vo](http://www.isthe.com/chongo/tech/comp/fnv/)
- [oklog/ulid -- reference ULID implementation in Go](https://github.com/oklog/ulid)
- [sony/sonyflake -- production Snowflake variant in Go](https://github.com/sony/sonyflake)
- [System Design Interview Vol. 1, Ch. 7 -- Design a Unique ID Generator in Distributed Systems](https://www.amazon.com/System-Design-Interview-insiders-Second/dp/B08CMF2CQF)

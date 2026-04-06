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

## Features

| Feature | Snowflake | ULID |
|---|---|---|
| Size | 64-bit integer | 128-bit (26-char string) |
| Timestamp precision | millisecond | millisecond |
| Lexicographic sort | ✓ (as int64) | ✓ (as string) |
| Uniqueness guarantee | machine ID + sequence | cryptographic randomness |
| Throughput target | ≥ 4 000 000 / sec | ≥ 1 000 000 / sec |
| Custom epoch | ✓ | — |
| Clock drift protection | ✓ | — |

---

## ID Formats

### Snowflake — 64-bit integer

```
 63        22        12        0
 ┌──────────┬─────────┬────────┐
 │  41-bit  │ 10-bit  │ 12-bit │
 │ timestamp│machine  │sequence│
 └──────────┴─────────┴────────┘
```

- **timestamp** — milliseconds since the configured epoch (default: 2020-01-01). 41 bits = ~69 years of IDs before overflow.
- **machine ID** — unique ID for this generator instance (0–1023). Set via env var, IP hash, or hostname hash.
- **sequence** — counter that resets every millisecond. 12 bits = 4096 values/ms → 4 096 000 IDs/sec before parking.

### ULID — 26-character string

```
 01ARZ3NDEKTSV4RRFFQ69G5FAV
 │          │               │
 10 chars   16 chars
 48-bit ms  80-bit random
```

- **timestamp** — 48-bit millisecond Unix timestamp (≈ year 10889 safe).
- **random** — 80 bits from `crypto/rand` — collision probability ≈ 1 in 10²⁴ per millisecond.
- Encoded in [Crockford Base32](https://www.crockford.com/base32.html) (26 chars, case-insensitive, no ambiguous chars).

---

## Project Structure

```
uid-generator-go/
├── uid/
│   ├── snowflake.go        ← Snowflake generator  (YOUR IMPLEMENTATION)
│   ├── snowflake_test.go   ← Snowflake tests       (pre-written)
│   ├── ulid.go             ← ULID generator        (YOUR IMPLEMENTATION)
│   ├── ulid_test.go        ← ULID tests            (pre-written)
│   ├── machineid.go        ← Machine ID sources    (YOUR IMPLEMENTATION)
│   └── machineid_test.go   ← Machine ID tests      (pre-written)
├── bench/
│   └── bench_test.go       ← Throughput benchmarks
├── .github/
│   └── workflows/
│       └── go-ci.yml
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.22+

### Run the tests (all will fail until you implement the stubs)

```bash
go test ./...
```

### Run tests with the race detector

```bash
go test -race ./...
```

### Run benchmarks

```bash
# Full suite
go test -bench=. -benchmem -benchtime=5s ./...

# Snowflake only
go test -bench=BenchmarkSnowflake_SingleThread -benchtime=5s ./bench/

# ULID only
go test -bench=BenchmarkULID_SingleThread -benchtime=5s ./bench/
```

---

## Implementation Guide

Work through the files in this order:

### Step 1 — `uid/snowflake.go`

1. `NewSnowflake` — validate `machineID`, apply `DefaultEpoch` if zero.
2. `currentMS` — return `time.Since(s.epoch).Milliseconds()`.
3. `waitNextMS` — spin until the clock advances past `last`.
4. `NextID` — assemble the 64-bit ID from timestamp, machine ID, and sequence. Handle clock drift and sequence overflow.
5. `DecomposeID` — extract each field with bit masks and shifts.

### Step 2 — `uid/ulid.go`

1. `NewULIDWithTime` — encode ms into bytes 0–5; fill bytes 6–15 with `randomBytes`.
2. `ULID.String` — encode 128 bits as 26 Crockford Base32 characters (see bit layout in the doc comment).
3. `ParseULID` — reverse the `String()` encoding.
4. `ULID.Time` — reconstruct ms from bytes 0–5.
5. `ULID.Compare` — byte-by-byte comparison.
6. `decodeBase32Byte` — map Crockford chars (including I/L/O aliases) to 5-bit values.

### Step 3 — `uid/machineid.go`

1. `MachineIDFromEnv` — parse the env var, validate range.
2. `MachineIDFromHostname` — `h.Sum32() & MaxMachineID`.
3. `MachineIDFromIP` — iterate `net.InterfaceAddrs`, hash the first non-loopback IPv4.
4. `ResolveMachineID` — try sources in order; return first success.

---

## Useful References

- [Snowflake original announcement (Twitter)](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake)
- [ULID specification](https://github.com/ulid/spec)
- [Crockford Base32](https://www.crockford.com/base32.html)
- [FNV hash](http://www.isthe.com/chongo/tech/comp/fnv/)

---

## Machine ID Resolution Order

| Priority | Source | How |
|---|---|---|
| 1 | `UID_MACHINE_ID` env var | Explicit integer string |
| 2 | Primary IPv4 address | FNV-1a hash of IP bytes |
| 3 | Hostname | FNV-1a hash of hostname string |

Set `UID_MACHINE_ID=<n>` for deterministic IDs across restarts.

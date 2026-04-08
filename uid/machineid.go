package uid

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strconv"
)

// MachineIDEnvVar is the environment variable checked first when resolving the machine ID.
// Set it to a decimal integer in [0, MaxMachineID], e.g.:
//
//	export UID_MACHINE_ID=7
const MachineIDEnvVar = "UID_MACHINE_ID"

// MachineIDFromEnv reads the machine ID from the UID_MACHINE_ID environment variable.
// Returns an error when the variable is unset, not parseable, or outside [0, MaxMachineID].
func MachineIDFromEnv() (int64, error) {
	val := os.Getenv(MachineIDEnvVar)
	if val == "" {
		return 0, fmt.Errorf("%s is not set", MachineIDEnvVar)
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %w", val, err)
	}
	if n < 0 || n > MaxMachineID {
		return 0, fmt.Errorf("value %d out of range [0, %d]", n, MaxMachineID)
	}
	return n, nil
}

// MachineIDFromHostname hashes the system hostname with FNV-1a 32-bit and
// maps the result into [0, MaxMachineID].
func MachineIDFromHostname() (int64, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return 0, err
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(hostname))

	return int64(h.Sum32()) & MaxMachineID, nil
}

// MachineIDFromIP hashes the first non-loopback IPv4 address with FNV-1a 32-bit
// and maps the result into [0, MaxMachineID].
// Returns an error when no suitable address can be found.
func MachineIDFromIP() (int64, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil || ip.IsLoopback() {
			continue
		}
		h := fnv.New32a()
		h.Write(ip)
		return int64(h.Sum32()) & MaxMachineID, nil
	}
	return 0, errors.New("no suitable non-loopback IPv4 address found")
}

// ResolveMachineID tries sources in priority order and returns the first success:
//
//  1. Environment variable  (MachineIDFromEnv)
//  2. IP address hash       (MachineIDFromIP)
//  3. Hostname hash         (MachineIDFromHostname)
func ResolveMachineID() (int64, error) {
	if id, err := MachineIDFromEnv(); err == nil {
		return id, nil
	}
	if id, err := MachineIDFromIP(); err == nil {
		return id, nil
	}
	return MachineIDFromHostname()
}

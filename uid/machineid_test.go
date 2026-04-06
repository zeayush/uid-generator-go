package uid

import (
	"os"
	"testing"
)

func TestMachineIDFromEnv_Valid(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "42")
	defer os.Unsetenv(MachineIDEnvVar)

	id, err := MachineIDFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestMachineIDFromEnv_Zero(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "0")
	defer os.Unsetenv(MachineIDEnvVar)

	id, err := MachineIDFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 0 {
		t.Errorf("expected 0, got %d", id)
	}
}

func TestMachineIDFromEnv_Max(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "1023")
	defer os.Unsetenv(MachineIDEnvVar)

	id, err := MachineIDFromEnv()
	if err != nil {
		t.Fatalf("unexpected error for MaxMachineID: %v", err)
	}
	if id != MaxMachineID {
		t.Errorf("expected %d, got %d", MaxMachineID, id)
	}
}

func TestMachineIDFromEnv_Unset(t *testing.T) {
	os.Unsetenv(MachineIDEnvVar)
	_, err := MachineIDFromEnv()
	if err == nil {
		t.Error("expected error when env var not set, got nil")
	}
}

func TestMachineIDFromEnv_OutOfRange(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "9999")
	defer os.Unsetenv(MachineIDEnvVar)

	_, err := MachineIDFromEnv()
	if err == nil {
		t.Error("expected error for 9999 (> MaxMachineID 1023), got nil")
	}
}

func TestMachineIDFromEnv_Negative(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "-1")
	defer os.Unsetenv(MachineIDEnvVar)

	_, err := MachineIDFromEnv()
	if err == nil {
		t.Error("expected error for negative value, got nil")
	}
}

func TestMachineIDFromEnv_NotANumber(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "not-a-number")
	defer os.Unsetenv(MachineIDEnvVar)

	_, err := MachineIDFromEnv()
	if err == nil {
		t.Error("expected error for non-numeric value, got nil")
	}
}

func TestMachineIDFromHostname_Range(t *testing.T) {
	id, err := MachineIDFromHostname()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id < 0 || id > MaxMachineID {
		t.Errorf("id %d outside valid range [0, %d]", id, MaxMachineID)
	}
}

func TestMachineIDFromHostname_Deterministic(t *testing.T) {
	id1, err := MachineIDFromHostname()
	if err != nil {
		t.Skipf("hostname unavailable: %v", err)
	}
	id2, _ := MachineIDFromHostname()
	if id1 != id2 {
		t.Errorf("hostname machineID is not deterministic: %d != %d", id1, id2)
	}
}

func TestMachineIDFromIP_Range(t *testing.T) {
	id, err := MachineIDFromIP()
	if err != nil {
		t.Skipf("no suitable IP in this environment (common in restricted CI): %v", err)
	}
	if id < 0 || id > MaxMachineID {
		t.Errorf("id %d outside valid range [0, %d]", id, MaxMachineID)
	}
}

func TestResolveMachineID_ReturnsValidIDWithoutEnv(t *testing.T) {
	os.Unsetenv(MachineIDEnvVar)
	id, err := ResolveMachineID()
	if err != nil {
		t.Fatalf("ResolveMachineID failed: %v", err)
	}
	if id < 0 || id > MaxMachineID {
		t.Errorf("resolved id %d outside valid range [0, %d]", id, MaxMachineID)
	}
}

func TestResolveMachineID_EnvTakesPriority(t *testing.T) {
	os.Setenv(MachineIDEnvVar, "7")
	defer os.Unsetenv(MachineIDEnvVar)

	id, err := ResolveMachineID()
	if err != nil {
		t.Fatalf("ResolveMachineID failed: %v", err)
	}
	if id != 7 {
		t.Errorf("expected env-var value 7, got %d", id)
	}
}

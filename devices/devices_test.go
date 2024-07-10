package devices

import (
	"testing"
)

func checkDevices(t *testing.T, d map[string]Device) {
	t.Helper()
	for k, v := range d {
		t.Logf("Checking %v", k)
		if v.ID == "" {
			t.Errorf("%s: Empty ID", v.ID)
		}
		if k != v.ID {
			t.Errorf("%s: ID mismatch: %v", k, v.ID)
		}
		if v.BastionID == "" {
			t.Logf("%s: Warning, no BastionID present", k)
		}
		if v.WitnessPubkey == "" {
			t.Errorf("%s: no witness pubkey present", k)
		}
	}
}

func TestCIFiles(t *testing.T) {
	checkDevices(t, CI)
}

func TestProdFiles(t *testing.T) {
	checkDevices(t, Prod)
}

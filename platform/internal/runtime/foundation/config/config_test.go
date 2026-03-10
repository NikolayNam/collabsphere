package config

import "testing"

func TestPlatformBootstrapAccountUUIDs(t *testing.T) {
	auth := Auth{PlatformBootstrapIDs: "11111111-1111-1111-1111-111111111111, 11111111-1111-1111-1111-111111111111 22222222-2222-2222-2222-222222222222"}
	ids, err := auth.PlatformBootstrapAccountUUIDs()
	if err != nil {
		t.Fatalf("PlatformBootstrapAccountUUIDs() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("PlatformBootstrapAccountUUIDs() len = %d, want 2", len(ids))
	}
}

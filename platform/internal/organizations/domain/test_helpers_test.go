package domain

import (
	"testing"

	"github.com/google/uuid"
)

func newTestUUID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("uuid.Parse(%q): %v", value, err)
	}
	return id
}

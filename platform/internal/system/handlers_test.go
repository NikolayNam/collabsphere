package system

import (
	"context"
	"errors"
	"testing"
)

func TestReadyHandlerReturnsReadyWhenCheckerPasses(t *testing.T) {
	handler := readyHandler(ReadyFunc(func(ctx context.Context) error {
		return nil
	}))

	resp, err := handler(context.Background(), &struct{}{})
	if err != nil {
		t.Fatalf("readyHandler() error = %v", err)
	}
	if resp.Body.Status != "ready" {
		t.Fatalf("ready status = %q, want ready", resp.Body.Status)
	}
}

func TestReadyHandlerFailsWhenCheckerFails(t *testing.T) {
	handler := readyHandler(ReadyFunc(func(ctx context.Context) error {
		return errors.New("db ping failed")
	}))

	if _, err := handler(context.Background(), &struct{}{}); err == nil {
		t.Fatal("readyHandler() error = nil, want failure")
	}
}

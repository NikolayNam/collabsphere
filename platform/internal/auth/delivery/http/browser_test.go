package http

import (
	"context"
	"testing"
)

func TestResolveBrowserReturnToUsesPublicBaseURLForRelativePath(t *testing.T) {
	h := &Handler{browser: BrowserFlowConfig{
		DefaultReturnURL: "/auth/callback",
		PublicBaseURL:    "http://api.localhost:8080",
	}}

	got, err := h.resolveBrowserReturnTo("")
	if err != nil {
		t.Fatalf("resolveBrowserReturnTo returned error: %v", err)
	}
	if got != "http://api.localhost:8080/auth/callback" {
		t.Fatalf("resolveBrowserReturnTo = %q, want %q", got, "http://api.localhost:8080/auth/callback")
	}
}

func TestResolveBrowserReturnToAllowsAbsolutePublicBaseOrigin(t *testing.T) {
	h := &Handler{browser: BrowserFlowConfig{
		AllowedRedirectOrigins: []string{"http://localhost:3001"},
		PublicBaseURL:          "http://api.localhost:8080",
	}}

	got, err := h.resolveBrowserReturnTo("http://api.localhost:8080/auth/callback")
	if err != nil {
		t.Fatalf("resolveBrowserReturnTo returned error: %v", err)
	}
	if got != "http://api.localhost:8080/auth/callback" {
		t.Fatalf("resolveBrowserReturnTo = %q, want %q", got, "http://api.localhost:8080/auth/callback")
	}
}

func TestLookupCallbackReturnToNormalizesDefaultReturnURL(t *testing.T) {
	h := &Handler{browser: BrowserFlowConfig{
		DefaultReturnURL: "/auth/callback",
		PublicBaseURL:    "http://api.localhost:8080",
	}}

	got := h.lookupCallbackReturnTo(context.Background(), "")
	if got != "http://api.localhost:8080/auth/callback" {
		t.Fatalf("lookupCallbackReturnTo = %q, want %q", got, "http://api.localhost:8080/auth/callback")
	}
}

package zitadeladmin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientForceVerifyUserEmailAlreadyVerified(t *testing.T) {
	var resendCalls int
	var verifyCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer pat-token" {
			t.Fatalf("Authorization header = %q", got)
		}
		switch r.URL.Path {
		case "/v2/users/user-1":
			writeJSON(t, w, map[string]any{
				"user": map[string]any{
					"human": map[string]any{
						"email": map[string]any{
							"email":      "user@example.com",
							"isVerified": true,
						},
					},
				},
			})
		case "/v2/users/user-1/email/resend":
			resendCalls++
			writeJSON(t, w, map[string]any{"verificationCode": "unused"})
		case "/v2/users/user-1/email/verify":
			verifyCalls++
			writeJSON(t, w, map[string]any{"details": map[string]any{}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := &Client{baseURL: srv.URL, token: "pat-token", httpClient: srv.Client()}
	res, err := client.ForceVerifyUserEmail(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ForceVerifyUserEmail() error = %v", err)
	}
	if res == nil {
		t.Fatal("ForceVerifyUserEmail() result is nil")
	}
	if !res.AlreadyVerified {
		t.Fatal("expected AlreadyVerified = true")
	}
	if resendCalls != 0 {
		t.Fatalf("resendCalls = %d, want 0", resendCalls)
	}
	if verifyCalls != 0 {
		t.Fatalf("verifyCalls = %d, want 0", verifyCalls)
	}
}

func TestClientForceVerifyUserEmailResendsAndVerifies(t *testing.T) {
	var resendCalls int
	var sendCalls int
	var verifyCalls int
	var verifyBody struct {
		VerificationCode string `json:"verificationCode"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer pat-token" {
			t.Fatalf("Authorization header = %q", got)
		}
		switch r.URL.Path {
		case "/v2/users/user-2":
			writeJSON(t, w, map[string]any{
				"user": map[string]any{
					"human": map[string]any{
						"email": map[string]any{
							"email":      "user2@example.com",
							"isVerified": false,
						},
					},
				},
			})
		case "/v2/users/user-2/email/resend":
			resendCalls++
			body := readBody(t, r)
			if !strings.Contains(body, "returnCode") {
				t.Fatalf("resend body %q does not contain returnCode", body)
			}
			writeJSON(t, w, map[string]any{"verificationCode": "abc-123"})
		case "/v2/users/user-2/email/send":
			sendCalls++
			writeJSON(t, w, map[string]any{"verificationCode": "unexpected"})
		case "/v2/users/user-2/email/verify":
			verifyCalls++
			if err := json.NewDecoder(r.Body).Decode(&verifyBody); err != nil {
				t.Fatalf("decode verify body: %v", err)
			}
			writeJSON(t, w, map[string]any{"details": map[string]any{}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := &Client{baseURL: srv.URL, token: "pat-token", httpClient: srv.Client()}
	res, err := client.ForceVerifyUserEmail(context.Background(), "user-2")
	if err != nil {
		t.Fatalf("ForceVerifyUserEmail() error = %v", err)
	}
	if res == nil {
		t.Fatal("ForceVerifyUserEmail() result is nil")
	}
	if res.AlreadyVerified {
		t.Fatal("expected AlreadyVerified = false")
	}
	if resendCalls != 1 {
		t.Fatalf("resendCalls = %d, want 1", resendCalls)
	}
	if sendCalls != 0 {
		t.Fatalf("sendCalls = %d, want 0", sendCalls)
	}
	if verifyCalls != 1 {
		t.Fatalf("verifyCalls = %d, want 1", verifyCalls)
	}
	if verifyBody.VerificationCode != "abc-123" {
		t.Fatalf("verificationCode = %q, want %q", verifyBody.VerificationCode, "abc-123")
	}
}

func TestClientForceVerifyUserEmailFallsBackToSendWhenResendHasNoCode(t *testing.T) {
	var resendCalls int
	var sendCalls int
	var verifyCalls int
	var verifyBody struct {
		VerificationCode string `json:"verificationCode"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer pat-token" {
			t.Fatalf("Authorization header = %q", got)
		}
		switch r.URL.Path {
		case "/v2/users/user-3":
			writeJSON(t, w, map[string]any{
				"user": map[string]any{
					"human": map[string]any{
						"email": map[string]any{
							"email":      "user3@example.com",
							"isVerified": false,
						},
					},
				},
			})
		case "/v2/users/user-3/email/resend":
			resendCalls++
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(t, w, map[string]any{
				"message": "Code is empty",
				"code":    "EMAIL-5w5ilin4yt",
			})
		case "/v2/users/user-3/email/send":
			sendCalls++
			body := readBody(t, r)
			if !strings.Contains(body, "returnCode") {
				t.Fatalf("send body %q does not contain returnCode", body)
			}
			writeJSON(t, w, map[string]any{"verificationCode": "fresh-code"})
		case "/v2/users/user-3/email/verify":
			verifyCalls++
			if err := json.NewDecoder(r.Body).Decode(&verifyBody); err != nil {
				t.Fatalf("decode verify body: %v", err)
			}
			writeJSON(t, w, map[string]any{"details": map[string]any{}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := &Client{baseURL: srv.URL, token: "pat-token", httpClient: srv.Client()}
	res, err := client.ForceVerifyUserEmail(context.Background(), "user-3")
	if err != nil {
		t.Fatalf("ForceVerifyUserEmail() error = %v", err)
	}
	if res == nil {
		t.Fatal("ForceVerifyUserEmail() result is nil")
	}
	if res.AlreadyVerified {
		t.Fatal("expected AlreadyVerified = false")
	}
	if resendCalls != 1 {
		t.Fatalf("resendCalls = %d, want 1", resendCalls)
	}
	if sendCalls != 1 {
		t.Fatalf("sendCalls = %d, want 1", sendCalls)
	}
	if verifyCalls != 1 {
		t.Fatalf("verifyCalls = %d, want 1", verifyCalls)
	}
	if verifyBody.VerificationCode != "fresh-code" {
		t.Fatalf("verificationCode = %q, want %q", verifyBody.VerificationCode, "fresh-code")
	}
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

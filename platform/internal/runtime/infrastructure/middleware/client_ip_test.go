package middleware

import "testing"

func TestExtractClientIPTrustedProxyUsesForwardedHeaders(t *testing.T) {
	ip := ExtractClientIP("203.0.113.10, 10.0.0.1", "198.51.100.10", "172.16.0.5:1234", ClientIPOptions{TrustProxyHeaders: true})
	if ip == nil || *ip != "203.0.113.10" {
		t.Fatalf("ExtractClientIP() = %v, want forwarded ip", ip)
	}
}

func TestExtractClientIPWithoutTrustedProxyFallsBackToRemoteAddr(t *testing.T) {
	ip := ExtractClientIP("203.0.113.10", "198.51.100.10", "172.16.0.5:1234", ClientIPOptions{TrustProxyHeaders: false})
	if ip == nil || *ip != "172.16.0.5" {
		t.Fatalf("ExtractClientIP() = %v, want remote host", ip)
	}
}

func TestExtractClientIPHandlesEmptyInput(t *testing.T) {
	ip := ExtractClientIP("", "", "", ClientIPOptions{TrustProxyHeaders: false})
	if ip != nil {
		t.Fatalf("ExtractClientIP() = %v, want nil", *ip)
	}
}

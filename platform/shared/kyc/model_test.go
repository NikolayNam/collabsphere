package kyc

import "testing"

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"draft", true},
		{"submitted", true},
		{"IN_REVIEW", true},
		{" needs_info ", true},
		{"approved", true},
		{"rejected", true},
		{"unknown", false},
	}
	for _, tc := range tests {
		_, ok := ParseStatus(tc.input)
		if ok != tc.ok {
			t.Fatalf("ParseStatus(%q) ok=%v, want %v", tc.input, ok, tc.ok)
		}
	}
}

func TestParseDecision(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"approve", true},
		{"reject", true},
		{"request_info", true},
		{" REQUEST_INFO ", true},
		{"invalid", false},
	}
	for _, tc := range tests {
		_, ok := ParseDecision(tc.input)
		if ok != tc.ok {
			t.Fatalf("ParseDecision(%q) ok=%v, want %v", tc.input, ok, tc.ok)
		}
	}
}

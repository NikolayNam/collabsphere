package validation

import (
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pw         string
		wantErrSub string
	}{
		{
			name:       "empty",
			pw:         "",
			wantErrSub: "Password is required",
		},
		{
			name:       "spaces only",
			pw:         "   ",
			wantErrSub: "Password is required",
		},
		{
			name:       "leading spaces",
			pw:         " secret1",
			wantErrSub: "must not start or end",
		},
		{
			name:       "trailing spaces",
			pw:         "secret1 ",
			wantErrSub: "must not start or end",
		},
		{
			name:       "too short",
			pw:         "12345",
			wantErrSub: "at least 8",
		},
		{
			name:       "ok ascii",
			pw:         "secret1123",
			wantErrSub: "",
		},
		{
			name:       "ok with inner spaces",
			pw:         "se cret1231",
			wantErrSub: "",
		},
		{
			name:       "too long for bcrypt bytes",
			pw:         strings.Repeat("a", 73),
			wantErrSub: "too long",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePassword(tt.pw)

			if tt.wantErrSub == "" {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrSub)
			}

			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErrSub, err)
			}
		})
	}
}

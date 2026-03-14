package application

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

func pkceChallengeS256(verifier string) (string, error) {
	verifier = strings.TrimSpace(verifier)
	if verifier == "" {
		return "", errors.New("pkce verifier is empty")
	}
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

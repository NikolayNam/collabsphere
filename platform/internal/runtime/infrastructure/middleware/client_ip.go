package middleware

import (
	"net"
	"strings"
)

type ClientIPOptions struct {
	TrustProxyHeaders bool
}

func ExtractClientIP(forwardedFor, realIP, remoteAddr string, options ClientIPOptions) *string {
	if options.TrustProxyHeaders {
		if value := firstHeaderValue(forwardedFor); value != "" {
			return &value
		}
		if value := strings.TrimSpace(realIP); value != "" {
			return &value
		}
	}
	if host := remoteAddrHost(remoteAddr); host != "" {
		return &host
	}
	return nil
}

func firstHeaderValue(value string) string {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}

func remoteAddrHost(remoteAddr string) string {
	value := strings.TrimSpace(remoteAddr)
	if value == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(value)
	if err != nil || strings.TrimSpace(host) == "" {
		return value
	}
	return strings.TrimSpace(host)
}

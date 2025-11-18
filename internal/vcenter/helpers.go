//go:build windows

package vcenter

import (
	"net/url"
	"strings"
)

// normalizeServerHost removes schema (http/https), port and path so only FQDN/IP remains
func normalizeServerHost(h string) string {
	hs := strings.TrimSpace(h)
	if hs == "" {
		return hs
	}
	if strings.HasPrefix(strings.ToLower(hs), "http://") || strings.HasPrefix(strings.ToLower(hs), "https://") {
		if u, err := url.Parse(hs); err == nil {
			hostPart := u.Host
			// Remove port
			if idx := strings.Index(hostPart, ":"); idx != -1 {
				hostPart = hostPart[:idx]
			}
			return hostPart
		}
	}
	// Remove path if user wrote e.g. vcenter.local/sdk
	if slash := strings.IndexRune(hs, '/'); slash != -1 {
		hs = hs[:slash]
	}
	// Remove port if no schema
	if idx := strings.Index(hs, ":"); idx != -1 {
		hs = hs[:idx]
	}
	return hs
}

package vcenter

import (
	"net/url"
	"strings"
)

// normalizeServerHost tar bort schema (http/https), port och path så endast FQDN/IP återstår
func normalizeServerHost(h string) string {
	hs := strings.TrimSpace(h)
	if hs == "" {
		return hs
	}
	if strings.HasPrefix(strings.ToLower(hs), "http://") || strings.HasPrefix(strings.ToLower(hs), "https://") {
		if u, err := url.Parse(hs); err == nil {
			hostPart := u.Host
			// Ta bort port
			if idx := strings.Index(hostPart, ":"); idx != -1 {
				hostPart = hostPart[:idx]
			}
			return hostPart
		}
	}
	// Ta bort path om användaren skrivit t.ex. vcenter.local/sdk
	if slash := strings.IndexRune(hs, '/'); slash != -1 {
		hs = hs[:slash]
	}
	// Ta bort port om utan schema
	if idx := strings.Index(hs, ":"); idx != -1 {
		hs = hs[:idx]
	}
	return hs
}

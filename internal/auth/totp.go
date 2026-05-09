// Package auth provides token resolution helpers for project accounts:
// TOTP code generation, OAuth2 token flows, and the unified resolver
// that returns the appropriate Authorization header value.
package auth

import (
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTP returns the current 6-digit TOTP code for the given
// base32 secret. Spaces in the secret are tolerated. An empty secret
// returns the empty string with no error.
func GenerateTOTP(secret string) (string, error) {
	clean := strings.ReplaceAll(secret, " ", "")
	if clean == "" {
		return "", nil
	}
	return totp.GenerateCode(strings.ToUpper(clean), time.Now().UTC())
}

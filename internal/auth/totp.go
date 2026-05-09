package auth

import (
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

func GenerateTOTP(secret string) (string, error) {
	clean := strings.ReplaceAll(secret, " ", "")
	if clean == "" {
		return "", nil
	}
	return totp.GenerateCode(strings.ToUpper(clean), time.Now().UTC())
}

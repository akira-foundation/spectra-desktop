package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateTOTP_EmptySecret(t *testing.T) {
	code, err := GenerateTOTP("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "" {
		t.Fatalf("expected empty code, got %q", code)
	}
}

func TestGenerateTOTP_StripsSpacesAndUppercases(t *testing.T) {
	const secret = "jbswy3dpehpk3pxp"
	code, err := GenerateTOTP("jbsw y3dp ehpk 3pxp")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	want, err := totp.GenerateCode("JBSWY3DPEHPK3PXP", time.Now().UTC())
	if err != nil {
		t.Fatalf("reference: %v", err)
	}
	if code != want {
		t.Fatalf("code = %q, want %q (secret %s)", code, want, secret)
	}
	if len(code) != 6 {
		t.Fatalf("code length = %d", len(code))
	}
}

func TestGenerateTOTP_InvalidBase32(t *testing.T) {
	_, err := GenerateTOTP("!!!not-base32!!!")
	if err == nil {
		t.Fatal("expected error for invalid base32")
	}
}

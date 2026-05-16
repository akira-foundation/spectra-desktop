package billing

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var (
	BillingURL    = ""
	BillingSecret = ""
	LicensePubKey = ""
	ProductSlug   = "spectra"
)

func init() {
	loadDotEnv()
	if BillingURL == "" {
		BillingURL = firstNonEmpty(os.Getenv("AKIRA_BILLING_URL"), "https://billing.akira-io.com")
	}
	if BillingSecret == "" {
		BillingSecret = os.Getenv("AKIRA_BILLING_SECRET")
	}
	if LicensePubKey == "" {
		LicensePubKey = os.Getenv("AKIRA_LICENSE_PUBKEY")
	}
}

func loadDotEnv() {
	for _, candidate := range dotEnvCandidates() {
		if applyDotEnvFile(candidate) {
			return
		}
	}
}

func dotEnvCandidates() []string {
	cwd, _ := os.Getwd()
	exe, _ := os.Executable()
	exeDir := ""
	if exe != "" {
		exeDir = filepath.Dir(exe)
	}
	return []string{
		filepath.Join(cwd, ".env.local"),
		filepath.Join(cwd, ".env"),
		filepath.Join(exeDir, ".env.local"),
		filepath.Join(exeDir, ".env"),
	}
}

func applyDotEnvFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])
		value = strings.Trim(value, `"'`)
		if key == "" || value == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return true
}

func IsConfigured() bool {
	return BillingSecret != "" && LicensePubKey != ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

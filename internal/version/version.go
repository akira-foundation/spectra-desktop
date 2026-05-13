package version

import "strings"

// Version is the application version.
// Overridden at build time via -ldflags "-X spectra-desktop/internal/version.Version=x.y.z".
var Version = "dev"

// IsDev reports whether the binary is running without a release version baked
// in (i.e. `wails dev` or any build without -ldflags). Used to scope on-disk
// paths so dev never shares state with production.
func IsDev() bool {
	v := strings.TrimSpace(Version)
	return v == "" || v == "dev"
}

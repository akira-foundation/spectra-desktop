package version

import "testing"

func TestVersion_DefaultValue(t *testing.T) {
	if Version != "dev" {
		t.Fatalf("expected default Version %q, got %q", "dev", Version)
	}
}

func TestIsDev_Matrix(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"DefaultDev", "dev", true},
		{"Empty", "", true},
		{"WhitespaceOnly", "   ", true},
		{"DevWithSurroundingSpaces", "  dev  ", true},
		{"TabAndNewlineWrappedDev", "\tdev\n", true},
		{"SemverRelease", "1.0.0", false},
		{"SemverPrerelease", "1.0.0-beta.1", false},
		{"SemverWithBuildMeta", "1.0.0+build.5", false},
		{"PrefixedV", "v1.2.3", false},
		{"UppercaseDev", "DEV", false},
		{"MixedCaseDev", "Dev", false},
		{"DevPrefixedToken", "dev-1", false},
		{"NumericZero", "0", false},
		{"NonSemverArbitrary", "snapshot", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := Version
			Version = tc.value
			t.Cleanup(func() { Version = original })
			if got := IsDev(); got != tc.want {
				t.Fatalf("IsDev() for %q = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestIsDev_RestoresAfterMutation(t *testing.T) {
	original := Version
	Version = "9.9.9"
	if IsDev() {
		t.Fatalf("expected IsDev() false for release version")
	}
	Version = original
	if !IsDev() {
		t.Fatalf("expected IsDev() true after restoring default")
	}
}

package updater

import "testing"

func TestCompareSemver_Matrix(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"equal_with_v_prefix", "v1.2.3", "1.2.3", 0},
		{"equal_with_whitespace", "  1.2.3  ", "1.2.3", 0},
		{"major_greater", "2.0.0", "1.9.9", 1},
		{"major_less", "1.0.0", "2.0.0", -1},
		{"minor_greater", "1.3.0", "1.2.9", 1},
		{"minor_less", "1.2.0", "1.3.0", -1},
		{"patch_greater", "1.2.4", "1.2.3", 1},
		{"patch_less", "1.2.3", "1.2.4", -1},
		{"prerelease_lower_than_release", "1.0.0-beta.1", "1.0.0", -1},
		{"release_higher_than_prerelease", "1.0.0", "1.0.0-beta.1", 1},
		{"prerelease_numeric_order", "1.0.0-beta.1", "1.0.0-beta.2", -1},
		{"prerelease_numeric_order_rev", "1.0.0-beta.2", "1.0.0-beta.1", 1},
		{"prerelease_alpha_lt_beta", "1.0.0-alpha", "1.0.0-beta", -1},
		{"prerelease_more_ids_wins", "1.0.0-beta.1.1", "1.0.0-beta.1", 1},
		{"prerelease_equal", "1.0.0-rc.1", "1.0.0-rc.1", 0},
		{"numeric_lt_alphanumeric", "1.0.0-1", "1.0.0-alpha", -1},
		{"alphanumeric_gt_numeric", "1.0.0-alpha", "1.0.0-1", 1},
		{"build_metadata_ignored", "1.2.3+build1", "1.2.3+build2", 0},
		{"invalid_treated_as_zero", "garbage", "0.0.0", 0},
		{"empty_treated_as_zero", "", "0.0.0", 0},
		{"partial_version", "1.2", "1.2.0", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareSemver(tc.a, tc.b); got != tc.want {
				t.Fatalf("compareSemver(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestFormatVersion_AlwaysPrefixed(t *testing.T) {
	cases := map[string]string{
		"1.2.3":         "v1.2.3",
		"v1.2.3":        "v1.2.3",
		"0.0.1-beta.1":  "v0.0.1-beta.1",
		"v0.0.1-beta.1": "v0.0.1-beta.1",
	}
	for in, want := range cases {
		if got := formatVersion(in); got != want {
			t.Fatalf("formatVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitSemver_ExtractsCoreAndPre(t *testing.T) {
	core, pre := splitSemver("v2.3.4-rc.1+meta")
	if core != [3]int{2, 3, 4} {
		t.Fatalf("core = %v, want [2 3 4]", core)
	}
	if pre != "rc.1" {
		t.Fatalf("pre = %q, want %q", pre, "rc.1")
	}
}

func TestSplitSemver_BuildMetadataOnly(t *testing.T) {
	core, pre := splitSemver("1.0.0+build.42")
	if core != [3]int{1, 0, 0} {
		t.Fatalf("core = %v, want [1 0 0]", core)
	}
	if pre != "" {
		t.Fatalf("pre = %q, want empty", pre)
	}
}

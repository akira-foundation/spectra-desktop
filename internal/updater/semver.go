package updater

import (
	"fmt"
	"strconv"
	"strings"
)

// compareSemver returns 1 if a > b, -1 if a < b, 0 if equal.
// Implements semver 2.0 precedence: numeric core compared first, then
// pre-release identifiers per spec (a build with pre-release is lower
// than one without; pre-release identifiers compared numerically when
// both numeric, otherwise lexically).
func compareSemver(a, b string) int {
	coreA, preA := splitSemver(a)
	coreB, preB := splitSemver(b)

	for i := 0; i < 3; i++ {
		if coreA[i] > coreB[i] {
			return 1
		}
		if coreA[i] < coreB[i] {
			return -1
		}
	}

	// Pre-release rules (semver 2.0 §11.4):
	//   - missing pre-release > any pre-release
	//   - pre-release identifiers compared one by one
	if preA == "" && preB == "" {
		return 0
	}
	if preA == "" {
		return 1
	}
	if preB == "" {
		return -1
	}
	return comparePreRelease(preA, preB)
}

func splitSemver(v string) ([3]int, string) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	pre := ""
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		if v[idx] == '-' {
			pre = v[idx+1:]
			if plus := strings.Index(pre, "+"); plus >= 0 {
				pre = pre[:plus]
			}
		}
		v = v[:idx]
	}
	parts := strings.SplitN(v, ".", 3)
	out := [3]int{0, 0, 0}
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(parts[i])
		out[i] = n
	}
	return out, pre
}

func comparePreRelease(a, b string) int {
	idsA := strings.Split(a, ".")
	idsB := strings.Split(b, ".")
	for i := 0; i < len(idsA) && i < len(idsB); i++ {
		c := comparePreReleaseIdent(idsA[i], idsB[i])
		if c != 0 {
			return c
		}
	}
	if len(idsA) > len(idsB) {
		return 1
	}
	if len(idsA) < len(idsB) {
		return -1
	}
	return 0
}

func comparePreReleaseIdent(a, b string) int {
	na, errA := strconv.Atoi(a)
	nb, errB := strconv.Atoi(b)
	switch {
	case errA == nil && errB == nil:
		if na > nb {
			return 1
		}
		if na < nb {
			return -1
		}
		return 0
	case errA == nil:
		// numeric identifiers have lower precedence than alphanumeric
		return -1
	case errB == nil:
		return 1
	default:
		if a > b {
			return 1
		}
		if a < b {
			return -1
		}
		return 0
	}
}

func formatVersion(v string) string {
	return fmt.Sprintf("v%s", strings.TrimPrefix(v, "v"))
}

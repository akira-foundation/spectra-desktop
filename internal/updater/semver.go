package updater

import (
	"fmt"
	"strconv"
	"strings"
)

// compareSemver returns 1 if a > b, -1 if a < b, 0 if equal.
// Accepts versions with optional leading "v" and ignores pre-release/build suffixes.
func compareSemver(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] > pb[i] {
			return 1
		}
		if pa[i] < pb[i] {
			return -1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.SplitN(v, ".", 3)
	out := [3]int{0, 0, 0}
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(parts[i])
		out[i] = n
	}
	return out
}

func formatVersion(v string) string {
	return fmt.Sprintf("v%s", strings.TrimPrefix(v, "v"))
}

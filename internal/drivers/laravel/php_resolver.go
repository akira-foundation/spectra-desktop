package laravel

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/akira-io/desktopkit/shell"
)

var (
	phpResolveOnce sync.Once
	phpResolved    string
	phpResolveErr  error

	phpOverrideMu sync.RWMutex
	phpOverride   string
)

func SetPHPBinaryOverride(path string) {
	phpOverrideMu.Lock()
	phpOverride = strings.TrimSpace(path)
	phpOverrideMu.Unlock()
	phpResolveOnce = sync.Once{}
	phpResolved = ""
	phpResolveErr = nil
}

func currentPHPOverride() string {
	phpOverrideMu.RLock()
	defer phpOverrideMu.RUnlock()
	return phpOverride
}

func ResolvePHPBinaryPath() (string, error) {
	return resolvePHPBinary()
}

func resolvePHPBinary() (string, error) {
	if override := currentPHPOverride(); override != "" {
		if isExecutableFile(override) {
			return override, nil
		}
	}
	phpResolveOnce.Do(func() {
		phpResolved, phpResolveErr = findPHPBinary()
	})
	return phpResolved, phpResolveErr
}

func findPHPBinary() (string, error) {
	resolved, err := shell.NewCandidates().
		WithName("php").
		Resolve()
	if err == nil {
		return resolved.AbsolutePath(), nil
	}
	if path := lookupPHPViaLoginShell(); path != "" {
		return path, nil
	}
	resolved, err = shell.NewCandidates().
		WithCandidates(systemPHPCandidates()).
		Resolve()
	if err == nil {
		return resolved.AbsolutePath(), nil
	}
	return "", ErrPHPNotFound
}

func systemPHPCandidates() []string {
	return []string{
		"/opt/homebrew/bin/php",
		"/usr/local/bin/php",
		"/opt/local/bin/php",
		"/usr/bin/php",
	}
}

func lookupPHPViaLoginShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	cmd := exec.Command(shell, "-l", "-c", "command -v php")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	candidate := strings.TrimSpace(string(out))
	if candidate == "" || !isExecutableFile(candidate) {
		return ""
	}
	return candidate
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
}

func enrichedEnv() []string {
	env := os.Environ()
	pathValue := ""
	pathIdx := -1
	for i, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			pathValue = strings.TrimPrefix(kv, "PATH=")
			pathIdx = i
			break
		}
	}
	shellPath := loginShellPath()
	merged := mergeUniquePathSegments(pathValue, splitPath(shellPath))
	entry := "PATH=" + merged
	if pathIdx >= 0 {
		env[pathIdx] = entry
	} else {
		env = append(env, entry)
	}
	return env
}

func loginShellPath() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	out, err := exec.Command(shell, "-l", "-c", "echo $PATH").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func splitPath(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, string(os.PathListSeparator))
}

func mergeUniquePathSegments(existing string, extras []string) string {
	seen := map[string]struct{}{}
	parts := []string{}
	add := func(seg string) {
		if seg == "" {
			return
		}
		_, ok := seen[seg]
		if ok {
			return
		}
		seen[seg] = struct{}{}
		parts = append(parts, seg)
	}
	for _, seg := range splitPath(existing) {
		add(seg)
	}
	for _, seg := range extras {
		add(filepath.Clean(seg))
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

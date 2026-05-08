package laravel

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrNoFormRequest      = errors.New("no FormRequest detected")
	ErrControllerNotFound = errors.New("controller file not found")
	ErrRulesNotFound      = errors.New("rules() method not found")
)

var (
	formRequestParamRegex = regexp.MustCompile(`(?:\\)?([A-Za-z_][A-Za-z0-9_]*Request)\s+\$\w+`)
	useImportRegex        = regexp.MustCompile(`use\s+([A-Za-z0-9_\\]+);`)
	rulesMethodRegex      = regexp.MustCompile(`(?:public|protected)\s+function\s+rules\s*\([^)]*\)`)
)

func inferFromFormRequest(projectPath, handler, methodName string) (*RequestSchema, error) {
	controllerFile, err := resolveControllerFile(projectPath, handler)
	if err != nil {
		return nil, err
	}
	src, err := readFile(controllerFile)
	if err != nil {
		return nil, err
	}
	signature, ok := findMethodSignature(src, methodName)
	if !ok {
		return nil, ErrNoFormRequest
	}
	requestClass, ok := findFormRequestParam(signature)
	if !ok {
		return nil, ErrNoFormRequest
	}
	if requestClass == "Request" || requestClass == "FormRequest" {
		return nil, ErrNoFormRequest
	}

	formRequestFile, err := resolveFormRequestFile(projectPath, src, requestClass)
	if err != nil {
		return nil, err
	}
	formSrc, err := readFile(formRequestFile)
	if err != nil {
		return nil, err
	}
	body, ok := extractRulesArray(formSrc)
	if !ok {
		return nil, ErrRulesNotFound
	}
	fields := parseRulePairs(body)
	if len(fields) == 0 {
		return nil, ErrRulesNotFound
	}
	fields = expandConfirmedFields(fields)
	return &RequestSchema{
		Source:     SchemaSourceFormRequest,
		Confidence: ConfidenceHigh,
		Fields:     fields,
	}, nil
}

func resolveControllerFile(projectPath, handler string) (string, error) {
	className := strings.SplitN(handler, "@", 2)[0]
	className = strings.TrimSpace(className)
	if className == "" {
		return "", ErrControllerNotFound
	}
	return classToProjectPath(projectPath, className)
}

func classToProjectPath(projectPath, fullyQualified string) (string, error) {
	clean := strings.TrimLeft(strings.ReplaceAll(fullyQualified, "/", "\\"), "\\")
	parts := strings.Split(clean, "\\")
	if len(parts) < 2 {
		return "", ErrControllerNotFound
	}
	if strings.EqualFold(parts[0], "App") {
		parts = parts[1:]
	}
	rel := filepath.Join(parts...)
	candidate := filepath.Join(projectPath, "app", rel+".php")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", ErrControllerNotFound
}

func findMethodSignature(src, methodName string) (string, bool) {
	pattern := regexp.MustCompile(`(?:public|protected|private)\s+(?:static\s+)?function\s+` + regexp.QuoteMeta(methodName) + `\s*\(([^)]*)\)`)
	m := pattern.FindStringSubmatch(src)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func findFormRequestParam(signature string) (string, bool) {
	matches := formRequestParamRegex.FindAllStringSubmatch(signature, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			cls := strings.TrimSpace(m[1])
			if cls != "" {
				return cls, true
			}
		}
	}
	return "", false
}

func resolveFormRequestFile(projectPath, controllerSrc, className string) (string, error) {
	imports := useImportRegex.FindAllStringSubmatch(controllerSrc, -1)
	for _, m := range imports {
		if len(m) < 2 {
			continue
		}
		full := strings.TrimSpace(m[1])
		segments := strings.Split(full, "\\")
		if len(segments) == 0 {
			continue
		}
		last := segments[len(segments)-1]
		if last == className {
			path, err := classToProjectPath(projectPath, full)
			if err == nil {
				return path, nil
			}
		}
	}
	defaultPath := filepath.Join(projectPath, "app", "Http", "Requests", className+".php")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}
	return "", ErrControllerNotFound
}

func extractRulesArray(src string) (string, bool) {
	loc := rulesMethodRegex.FindStringIndex(src)
	if loc == nil {
		return "", false
	}
	bodyStart := strings.IndexByte(src[loc[1]:], '{')
	if bodyStart < 0 {
		return "", false
	}
	body, _, ok := extractBraceBlock(src, loc[1]+bodyStart+1)
	if !ok {
		return "", false
	}
	returnIdx := strings.Index(body, "return")
	if returnIdx < 0 {
		return "", false
	}
	bracketIdx := strings.IndexByte(body[returnIdx:], '[')
	if bracketIdx < 0 {
		return "", false
	}
	arr, _, ok := extractArrayLiteral(body, returnIdx+bracketIdx+1)
	if !ok {
		return "", false
	}
	return arr, true
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

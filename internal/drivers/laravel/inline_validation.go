package laravel

import (
	"errors"
	"regexp"
	"strings"
)

var ErrNoInlineValidation = errors.New("no inline validation detected")

var (
	requestValidateRegex = regexp.MustCompile(`->\s*validate\s*\(\s*\[`)
	validatorMakeRegex   = regexp.MustCompile(`Validator\s*::\s*make\s*\([^,]+,\s*\[`)
)

func inferFromInlineValidation(projectPath, handler, methodName string) (*RequestSchema, error) {
	controllerFile, err := resolveControllerFile(projectPath, handler)
	if err != nil {
		return nil, err
	}
	src, err := readFile(controllerFile)
	if err != nil {
		return nil, err
	}
	body, ok := extractMethodBody(src, methodName)
	if !ok {
		return nil, ErrNoInlineValidation
	}
	rulesArr, ok := findInlineRulesArray(body)
	if !ok {
		return nil, ErrNoInlineValidation
	}
	fields := parseRulePairs(rulesArr)
	if len(fields) == 0 {
		return nil, ErrNoInlineValidation
	}
	fields = expandConfirmedFields(fields)
	return &RequestSchema{
		Source:     SchemaSourceInline,
		Confidence: ConfidenceMedium,
		Fields:     fields,
	}, nil
}

func extractMethodBody(src, methodName string) (string, bool) {
	pattern := regexp.MustCompile(`(?:public|protected|private)\s+(?:static\s+)?function\s+` + regexp.QuoteMeta(methodName) + `\s*\([^)]*\)\s*(?::\s*[^\{]+)?`)
	loc := pattern.FindStringIndex(src)
	if loc == nil {
		return "", false
	}
	braceIdx := strings.IndexByte(src[loc[1]:], '{')
	if braceIdx < 0 {
		return "", false
	}
	body, _, ok := extractBraceBlock(src, loc[1]+braceIdx+1)
	return body, ok
}

func findInlineRulesArray(body string) (string, bool) {
	if loc := requestValidateRegex.FindStringIndex(body); loc != nil {
		arr, _, ok := extractArrayLiteral(body, loc[1])
		if ok {
			return arr, true
		}
	}
	if loc := validatorMakeRegex.FindStringIndex(body); loc != nil {
		arr, _, ok := extractArrayLiteral(body, loc[1])
		if ok {
			return arr, true
		}
	}
	return "", false
}

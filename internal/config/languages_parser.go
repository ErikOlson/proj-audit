package config

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed languages.yaml
var embeddedLanguages []byte

func defaultLanguages() map[string]LanguageConfig {
	langs, err := ParseLanguagesYAML(embeddedLanguages)
	if err != nil {
		panic(fmt.Sprintf("invalid embedded languages.yaml: %v", err))
	}
	return langs
}

func LoadLanguagesFile(path string) (map[string]LanguageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read languages file: %w", err)
	}
	return ParseLanguagesYAML(data)
}

func ParseLanguagesYAML(data []byte) (map[string]LanguageConfig, error) {
	languages := make(map[string]LanguageConfig)
	var currentLang string
	var currentField string

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := countLeadingSpaces(line)
		content := strings.TrimSpace(line)
		switch indent {
		case 0:
			if !strings.HasSuffix(content, ":") {
				return nil, fmt.Errorf("line %d: expected language name ending with ':'", lineNo)
			}
			name := strings.TrimSuffix(content, ":")
			name = strings.Trim(name, `"'`)
			currentLang = name
			currentField = ""
			if _, exists := languages[name]; !exists {
				languages[name] = LanguageConfig{}
			}
		case 2:
			if currentLang == "" {
				return nil, fmt.Errorf("line %d: field without language context", lineNo)
			}
			if !strings.HasSuffix(content, ":") {
				return nil, fmt.Errorf("line %d: expected ':' after field name", lineNo)
			}
			field := strings.TrimSuffix(content, ":")
			if field != "extensions" && field != "skipDirs" {
				return nil, fmt.Errorf("line %d: unknown field %q", lineNo, field)
			}
			currentField = field
		case 4:
			if currentLang == "" || currentField == "" {
				return nil, fmt.Errorf("line %d: list item without context", lineNo)
			}
			if !strings.HasPrefix(content, "- ") {
				return nil, fmt.Errorf("line %d: expected '- value'", lineNo)
			}
			value := strings.TrimSpace(strings.TrimPrefix(content, "- "))
			lang := languages[currentLang]
			if currentField == "extensions" {
				lang.Extensions = append(lang.Extensions, value)
			} else {
				lang.SkipDirs = append(lang.SkipDirs, value)
			}
			languages[currentLang] = lang
		default:
			return nil, fmt.Errorf("line %d: unsupported indentation level %d", lineNo, indent)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return languages, nil
}

func MergeLanguageMaps(base, overrides map[string]LanguageConfig) map[string]LanguageConfig {
	if base == nil && overrides == nil {
		return nil
	}
	result := make(map[string]LanguageConfig)
	for name, lang := range base {
		result[name] = lang
	}
	for name, lang := range overrides {
		if existing, ok := result[name]; ok {
			result[name] = LanguageConfig{
				Extensions: appendUnique(existing.Extensions, lang.Extensions),
				SkipDirs:   appendUnique(existing.SkipDirs, lang.SkipDirs),
			}
		} else {
			result[name] = lang
		}
	}
	return result
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
			continue
		}
		if ch == '\t' {
			count += 2
			continue
		}
		break
	}
	return count
}

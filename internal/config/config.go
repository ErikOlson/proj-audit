package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LanguageConfig struct {
	Extensions []string `json:"extensions"`
	SkipDirs   []string `json:"skipDirs"`
}

type Config struct {
	Root          string                    `json:"root"`
	MaxDepth      int                       `json:"maxDepth"`
	Format        string                    `json:"format"`
	IgnoreDirs    []string                  `json:"ignoreDirs"`
	IncludeHidden bool                      `json:"includeHidden"`
	LanguagesFile string                    `json:"languagesFile"`
	Languages     map[string]LanguageConfig `json:"languages"`
	Analyzers     map[string]bool           `json:"analyzers"`
	Scoring       *ScoringConfig            `json:"scoring"`
}

func DefaultConfig() Config {
	return Config{
		Root:       ".",
		MaxDepth:   0,
		Format:     "tree",
		IgnoreDirs: defaultIgnoreDirs(),
		Languages:  defaultLanguages(),
		Analyzers:  defaultAnalyzerToggles(),
		Scoring:    DefaultScoringConfig(),
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", filepath.Base(path), err)
	}
	return cfg, nil
}

func Merge(base Config, overrides Config) Config {
	merged := base

	if overrides.Root != "" {
		merged.Root = overrides.Root
	}
	if overrides.MaxDepth != 0 {
		merged.MaxDepth = overrides.MaxDepth
	}
	if overrides.Format != "" {
		merged.Format = overrides.Format
	}
	if overrides.IgnoreDirs != nil {
		merged.IgnoreDirs = appendUnique(merged.IgnoreDirs, overrides.IgnoreDirs)
	}
	if overrides.IncludeHidden {
		merged.IncludeHidden = true
	}
	if overrides.LanguagesFile != "" {
		merged.LanguagesFile = overrides.LanguagesFile
	}
	if len(overrides.Languages) > 0 {
		merged.Languages = MergeLanguageMaps(merged.Languages, overrides.Languages)
	}
	if overrides.Analyzers != nil {
		if merged.Analyzers == nil {
			merged.Analyzers = make(map[string]bool)
		}
		for name, enabled := range overrides.Analyzers {
			merged.Analyzers[strings.ToLower(name)] = enabled
		}
	}
	if overrides.Scoring != nil {
		merged.Scoring = overrides.Scoring
	}
	return merged
}

func (c Config) AllIgnoreDirs() []string {
	set := make(map[string]struct{})
	for _, dir := range c.IgnoreDirs {
		clean := strings.TrimSpace(dir)
		if clean == "" {
			continue
		}
		set[clean] = struct{}{}
	}
	for _, lang := range c.Languages {
		for _, dir := range lang.SkipDirs {
			clean := strings.TrimSpace(dir)
			if clean == "" {
				continue
			}
			set[clean] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for dir := range set {
		out = append(out, dir)
	}
	return out
}

func (c Config) ExtensionMapping() map[string]string {
	mapping := make(map[string]string)
	for language, langConfig := range c.Languages {
		for _, ext := range langConfig.Extensions {
			clean := strings.TrimSpace(ext)
			if clean == "" {
				continue
			}
			if !strings.HasPrefix(clean, ".") {
				clean = "." + clean
			}
			mapping[strings.ToLower(clean)] = language
		}
	}
	return mapping
}

func (c Config) ResolveLanguages() (map[string]LanguageConfig, error) {
	langs := defaultLanguages()
	if c.LanguagesFile != "" {
		fileLangs, err := LoadLanguagesFile(c.LanguagesFile)
		if err != nil {
			return nil, err
		}
		langs = MergeLanguageMaps(langs, fileLangs)
	}
	if len(c.Languages) > 0 {
		langs = MergeLanguageMaps(langs, c.Languages)
	}
	return langs, nil
}

func (c Config) EffectiveAnalyzers() map[string]bool {
	toggles := defaultAnalyzerToggles()
	for name, enabled := range c.Analyzers {
		if name == "" {
			continue
		}
		toggles[strings.ToLower(name)] = enabled
	}
	return toggles
}

func defaultAnalyzerToggles() map[string]bool {
	var toggles map[string]bool
	if err := decodeYAML(defaultAnalyzersYAML, &toggles); err != nil {
		panic(fmt.Sprintf("invalid default analyzers yaml: %v", err))
	}
	return toggles
}

type ScoringConfig struct {
	Effort     EffortConfig   `json:"effort"`
	Polish     PolishConfig   `json:"polish"`
	Recency    []AgeThreshold `json:"recency"`
	Categories CategoryConfig `json:"categories"`
}

type EffortConfig struct {
	Commit []RangeThreshold `json:"commit"`
	Active []RangeThreshold `json:"active"`
}

type PolishConfig struct {
	Readme int `json:"readme"`
	Tests  int `json:"tests"`
	CI     int `json:"ci"`
	Docker int `json:"docker"`
}

type RangeThreshold struct {
	Min    int `json:"min"`
	Points int `json:"points"`
}

type AgeThreshold struct {
	MaxDays int `json:"maxDays"`
	Points  int `json:"points"`
}

type CategoryConfig struct {
	Experiment CategoryRule `json:"experiment"`
	Prototype  CategoryRule `json:"prototype"`
	Archived   CategoryRule `json:"archived"`
	Product    CategoryRule `json:"product"`
}

type CategoryRule struct {
	CommitMax  *int `json:"commitMax"`
	EffortMax  *int `json:"effortMax"`
	EffortMin  *int `json:"effortMin"`
	PolishMax  *int `json:"polishMax"`
	PolishMin  *int `json:"polishMin"`
	RecencyMax *int `json:"recencyMax"`
	RecencyMin *int `json:"recencyMin"`
}

func defaultLanguages() map[string]LanguageConfig {
	var langs map[string]LanguageConfig
	if err := decodeYAML(defaultLanguagesYAML, &langs); err != nil {
		panic(fmt.Sprintf("invalid default languages yaml: %v", err))
	}
	return langs
}

func LoadLanguagesFile(path string) (map[string]LanguageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read languages file: %w", err)
	}
	var langs map[string]LanguageConfig
	if err := decodeYAML(data, &langs); err != nil {
		return nil, err
	}
	return langs, nil
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

func appendUnique(base []string, more []string) []string {
	if len(more) == 0 {
		return base
	}
	set := make(map[string]struct{}, len(base)+len(more))
	var result []string
	for _, item := range base {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := set[item]; exists {
			continue
		}
		set[item] = struct{}{}
		result = append(result, item)
	}
	for _, item := range more {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := set[item]; exists {
			continue
		}
		set[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func defaultIgnoreDirs() []string {
	return []string{
		".git",
		"node_modules",
		"vendor",
		"bin",
		"dist",
		"build",
		"out",
		"target",
		".gocache",
		".cache",
		"venv",
		".venv",
		"__pycache__",
		".cargo",
	}
}

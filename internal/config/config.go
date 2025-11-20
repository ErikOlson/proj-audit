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
	Languages     map[string]LanguageConfig `json:"languages"`
}

func DefaultConfig() Config {
	return Config{
		Root:       ".",
		MaxDepth:   0,
		Format:     "tree",
		IgnoreDirs: defaultIgnoreDirs(),
		Languages:  defaultLanguages(),
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
	if len(overrides.Languages) > 0 {
		if merged.Languages == nil {
			merged.Languages = make(map[string]LanguageConfig)
		}
		for name, lang := range overrides.Languages {
			merged.Languages[name] = lang
		}
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

func defaultLanguages() map[string]LanguageConfig {
	return map[string]LanguageConfig{
		"Go": {
			Extensions: []string{".go"},
			SkipDirs:   []string{"vendor", "bin"},
		},
		"Rust": {
			Extensions: []string{".rs"},
			SkipDirs:   []string{"target", ".cargo"},
		},
		"Python": {
			Extensions: []string{".py"},
			SkipDirs:   []string{"__pycache__", ".venv", "venv"},
		},
		"JavaScript": {
			Extensions: []string{".js"},
			SkipDirs:   []string{"node_modules", "dist"},
		},
		"TypeScript": {
			Extensions: []string{".ts"},
			SkipDirs:   []string{"node_modules", "dist"},
		},
		"Java": {
			Extensions: []string{".java"},
			SkipDirs:   []string{"build", "out"},
		},
		"C#": {
			Extensions: []string{".cs"},
			SkipDirs:   []string{"bin", "obj"},
		},
		"C/C++": {
			Extensions: []string{".c", ".h", ".cpp", ".cc", ".cxx", ".hpp", ".hh"},
			SkipDirs:   []string{"build"},
		},
	}
}

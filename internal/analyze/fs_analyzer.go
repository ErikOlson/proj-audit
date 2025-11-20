package analyze

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type FsAnalyzer struct {
	ignoreDirs    map[string]struct{}
	includeHidden bool
}

func NewFsAnalyzer(ignoreDirs []string, includeHidden bool) *FsAnalyzer {
	return &FsAnalyzer{
		ignoreDirs:    makeIgnoreSet(ignoreDirs),
		includeHidden: includeHidden,
	}
}

func (f *FsAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	var metrics model.ProjectMetrics

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if p != path && f.shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			// Detect CI configurations eagerly to avoid additional passes.
			if d.Name() == ".github" {
				if info, err := os.Stat(filepath.Join(p, "workflows")); err == nil && info.IsDir() {
					metrics.HasCI = true
				}
			}
			return nil
		}

		metrics.Files++

		if lines, err := countLines(p); err == nil {
			metrics.LinesOfCode += lines
		}

		name := d.Name()
		lowerName := strings.ToLower(name)

		if filepath.Dir(p) == path && strings.HasPrefix(lowerName, "readme") {
			metrics.HasREADME = true
		}

		if hasTestIndicator(lowerName, p) {
			metrics.HasTests = true
		}

		if name == "Dockerfile" || lowerName == "docker-compose.yml" {
			metrics.HasDocker = true
		}

		if lowerName == ".gitlab-ci.yml" {
			metrics.HasCI = true
		}

		if info, err := d.Info(); err == nil {
			modTime := info.ModTime()
			if modTime.After(metrics.LastTouched) {
				metrics.LastTouched = modTime
			}
		}

		return nil
	})
	if err != nil {
		return model.ProjectMetrics{}, fmt.Errorf("fs analyzer walk: %w", err)
	}

	return metrics, nil
}

func (f *FsAnalyzer) shouldSkipDir(name string) bool {
	if name == "" {
		return false
	}
	if !f.includeHidden && strings.HasPrefix(name, ".") && name != ".git" && name != ".github" {
		return true
	}
	_, skip := f.ignoreDirs[name]
	return skip
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	reader.Buffer(buf, 1024*1024)

	count := 0
	for reader.Scan() {
		count++
	}
	if err := reader.Err(); err != nil && err != io.EOF {
		return count, err
	}
	return count, nil
}

func hasTestIndicator(lowerName, fullPath string) bool {
	if strings.HasSuffix(lowerName, "_test.go") {
		return true
	}

	jsTestSuffixes := []string{".spec.js", ".test.js", ".spec.ts", ".test.ts"}
	for _, suffix := range jsTestSuffixes {
		if strings.HasSuffix(lowerName, suffix) {
			return true
		}
	}

	if strings.Contains(fullPath, string(os.PathSeparator)+"__tests__"+string(os.PathSeparator)) {
		return true
	}

	return false
}

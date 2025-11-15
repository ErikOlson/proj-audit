package analyze

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type GitAnalyzer struct{}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

func (g *GitAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.ProjectMetrics{}, nil
		}
		return model.ProjectMetrics{}, fmt.Errorf("git analyzer stat: %w", err)
	}
	if !info.IsDir() {
		return model.ProjectMetrics{}, nil
	}

	metrics := model.ProjectMetrics{
		HasGit: true,
	}

	if count, err := gitInt(path, "rev-list", "--count", "HEAD"); err == nil {
		metrics.CommitCount = count
	}

	firstCommit, errFirst := gitTime(path, "log", "--format=%ct", "--reverse", "--max-count=1")
	lastCommit, errLast := gitTime(path, "log", "-1", "--format=%ct")

	if errFirst == nil && errLast == nil && !lastCommit.IsZero() && !firstCommit.IsZero() && !lastCommit.Before(firstCommit) {
		metrics.ActiveDays = int(lastCommit.Sub(firstCommit).Hours() / 24)
	}

	if errLast == nil {
		metrics.LastTouched = lastCommit
	}

	return metrics, nil
}

func gitInt(path string, args ...string) (int, error) {
	out, err := gitOutput(path, args...)
	if err != nil {
		return 0, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return 0, fmt.Errorf("empty git output")
	}
	value, err := strconv.Atoi(out)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func gitTime(path string, args ...string) (time.Time, error) {
	out, err := gitOutput(path, args...)
	if err != nil {
		return time.Time{}, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return time.Time{}, fmt.Errorf("empty git timestamp")
	}
	secs, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(secs, 0), nil
}

func gitOutput(path string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", path}, args...)
	cmd := exec.Command("git", cmdArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Return combined stderr for easier debugging, but keep error non-fatal to callers.
		return "", fmt.Errorf("git %v: %w: %s", args, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

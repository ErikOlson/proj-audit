package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type MarkdownRenderer struct{}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func (r *MarkdownRenderer) Render(root *model.Node, w io.Writer) error {
	if root == nil {
		return nil
	}

	if _, err := fmt.Fprintln(w, "# Project Audit Report"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if err := r.renderTable(root, w); err != nil {
		return err
	}

	treeBuf := &strings.Builder{}
	if err := NewTreeRenderer().Render(root, treeBuf); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "```"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, treeBuf.String()); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "```"); err != nil {
		return err
	}

	return nil
}

func (r *MarkdownRenderer) renderTable(root *model.Node, w io.Writer) error {
	projects := flattenProjects(root)

	if _, err := fmt.Fprintln(w, "| Name | Path | Category | Score | Languages | Commits | Last Touched |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "|------|------|----------|-------|-----------|---------|--------------|"); err != nil {
		return err
	}

	for _, project := range projects {
		category := project.Category
		if category == "" {
			category = "Uncategorized"
		}
		langs := strings.Join(project.Metrics.Languages, ", ")
		if langs == "" {
			langs = "-"
		}
		commits := project.Metrics.CommitCount
		last := "-"
		if !project.Metrics.LastTouched.IsZero() {
			last = project.Metrics.LastTouched.Format("2006-01-02")
		}

		row := fmt.Sprintf("| %s | %s | %s | %d | %s | %d | %s |",
			project.Name,
			project.Path,
			category,
			project.Scores.Overall,
			langs,
			commits,
			last,
		)
		if _, err := fmt.Fprintln(w, row); err != nil {
			return err
		}
	}

	if len(projects) == 0 {
		if _, err := fmt.Fprintln(w, "| _no projects detected_ | - | - | - | - | - | - |"); err != nil {
			return err
		}
	}

	return nil
}

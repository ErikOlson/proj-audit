package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type TreeRenderer struct{}

func NewTreeRenderer() *TreeRenderer {
	return &TreeRenderer{}
}

func (r *TreeRenderer) Render(root *model.Node, w io.Writer) error {
	if root == nil {
		return nil
	}

	if _, err := fmt.Fprintln(w, formatNodeLine(root.Name, root.Project)); err != nil {
		return err
	}

	for i, child := range root.Children {
		if err := r.renderNode(child, "", i == len(root.Children)-1, w); err != nil {
			return err
		}
	}
	return nil
}

func (r *TreeRenderer) renderNode(node *model.Node, prefix string, isLast bool, w io.Writer) error {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	line := formatNodeLine(fmt.Sprintf("%s%s%s", prefix, connector, node.Name), node.Project)
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}

	for i, child := range node.Children {
		if err := r.renderNode(child, childPrefix, i == len(node.Children)-1, w); err != nil {
			return err
		}
	}
	return nil
}

func formatNodeLine(base string, project *model.Project) string {
	if project == nil {
		return base
	}
	return base + " " + formatProjectSummary(project)
}

func formatProjectSummary(project *model.Project) string {
	if project == nil {
		return ""
	}

	category := project.Category
	if category == "" {
		category = "Uncategorized"
	}
	overall := project.Scores.Overall

	parts := []string{fmt.Sprintf("%s: %d", category, overall)}
	if languages := strings.Join(project.Metrics.Languages, ","); languages != "" {
		parts = append(parts, languages)
	}

	if project.Metrics.HasGit {
		parts = append(parts, fmt.Sprintf("%d commits", project.Metrics.CommitCount))
	} else if project.Metrics.CommitCount > 0 {
		parts = append(parts, fmt.Sprintf("%d commits", project.Metrics.CommitCount))
	}

	if !project.Metrics.LastTouched.IsZero() {
		parts = append(parts, fmt.Sprintf("last: %s", project.Metrics.LastTouched.Format("2006-01")))
	}

	return "[" + strings.Join(parts, " | ") + "]"
}

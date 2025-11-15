package render

import "github.com/ErikOlson/proj-audit/internal/model"

func flattenProjects(root *model.Node) []*model.Project {
	var projects []*model.Project
	collectProjects(root, &projects)
	return projects
}

func collectProjects(node *model.Node, projects *[]*model.Project) {
	if node == nil {
		return
	}
	if node.Project != nil {
		*projects = append(*projects, node.Project)
	}
	for _, child := range node.Children {
		collectProjects(child, projects)
	}
}

package render

import (
	"encoding/json"
	"io"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type JSONRenderer struct{}

func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func (r *JSONRenderer) Render(root *model.Node, w io.Writer) error {
	if root == nil {
		return nil
	}

	payload := struct {
		Root     *model.Node      `json:"root"`
		Projects []*model.Project `json:"projects"`
	}{
		Root:     root,
		Projects: flattenProjects(root),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

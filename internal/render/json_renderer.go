package render

import (
	"io"

	"github.com/your-username/proj-audit/internal/model"
)

type JSONRenderer struct{}

func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func (r *JSONRenderer) Render(root *model.Node, w io.Writer) error {
	// TODO: implement json rendering
	return nil
}

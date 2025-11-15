package render

import (
	"io"

	"github.com/your-username/proj-audit/internal/model"
)

type TreeRenderer struct{}

func NewTreeRenderer() *TreeRenderer {
	return &TreeRenderer{}
}

func (r *TreeRenderer) Render(root *model.Node, w io.Writer) error {
	// TODO: implement tree rendering
	return nil
}

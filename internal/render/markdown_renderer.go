package render

import (
	"io"

	"github.com/your-username/proj-audit/internal/model"
)

type MarkdownRenderer struct{}

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

func (r *MarkdownRenderer) Render(root *model.Node, w io.Writer) error {
	// TODO: implement markdown rendering
	return nil
}

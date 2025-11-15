package render

import (
	"io"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type Renderer interface {
	Render(root *model.Node, w io.Writer) error
}

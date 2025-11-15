package render

import (
	"io"

	"github.com/your-username/proj-audit/internal/model"
)

type Renderer interface {
	Render(root *model.Node, w io.Writer) error
}

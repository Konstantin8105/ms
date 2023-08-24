package window

import (
	"github.com/Konstantin8105/ds"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var _ ds.Window = (*Empty)(nil)

type Empty struct{}

func (e *Empty) SetMouseButtonCallback(
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
	xcursor, ycursor float64,
) {
}
func (e *Empty) SetCharCallback(r rune) {}
func (e *Empty) SetScrollCallback(
	xcursor, ycursor float64,
	xoffset, yoffset float64,
) {
}
func (e *Empty) SetCursorPosCallback(
	xpos float64,
	ypos float64,
) {
}
func (e *Empty) SetKeyCallback(
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
}
func (e *Empty) Draw(x, y, w, h int32) {}

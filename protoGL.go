//go:build ignore

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

func init() {
	runtime.LockOSThread()
}

var font *gltext.Font

const fontSize = int32(12)

func main() {
	if err := glfw.Init(); err != nil {
		err = fmt.Errorf("failed to initialize glfw: %v", err)
		return
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(800, 600, "3D model", nil, nil)
	if err != nil {
		return
	}
	window.MakeContextCurrent()

	// 	op.window.SetMouseButtonCallback(op.mouseButton)
	// 	op.window.SetScrollCallback(op.scroll)
	// 	op.window.SetCursorPosCallback(op.cursorPos)
	// 	op.window.SetKeyCallback(op.key)

	if err = gl.Init(); err != nil {
		return
	}

	glfw.SwapInterval(1) // Enable vsync

	// DrawText text on the screen
	DrawText := func(str string, x, y int32) {
		gl.Color4ub(0, 0, 0, 255)
		gl.LoadIdentity()
		font.Printf(float32(x), float32(y), str)
	}

	// create new Font from given filename (.ttf expected)
	fd, err := os.Open("/home/konstantin/.fonts/Go-Mono-Bold.ttf") // fontfile
	if err != nil {
		return
	}
	font, err = gltext.LoadTruetype(fd, fontSize, 32, 127, gltext.LeftToRight)
	if err != nil {
		return
	}
	err = fd.Close()
	if err != nil {
		return
	}
	// font is prepared

	// 	op.fps.Init()

	gl.Disable(gl.LIGHTING)

	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()

	// vl demo
	root, action := vl.Demo()
	_ = action

	for !window.ShouldClose() {
		// windows
		w, h := window.GetSize()
		gl.Viewport(0, 0, int32(w), int32(h))

		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)

		gl.Enable(gl.DEPTH_TEST)
		// gl.Enable(gl.BLEND) // Transparency
		// gl.Enable(gl.LINE_SMOOTH)

		// minimal screen notes
		// DrawText("Text UNICODE", 0, 0*fontSize)

		var widthSymbol uint = uint(w) / uint(fontSize)
		_ = root.Render(widthSymbol,
			func(row, col uint, st tcell.Style, r rune) {
				//	if row < 0 || uint(height) < row {
				//		return
				//	}
				//	if col < 0 || uint(width) < col {
				//		return
				//	}
				DrawText(string(r), int32(col)*fontSize, int32(row)*fontSize)
				// screen.SetCell(int(col), int(row), st, r)
			})

		// end
		window.MakeContextCurrent()
		window.SwapBuffers()
	}
}

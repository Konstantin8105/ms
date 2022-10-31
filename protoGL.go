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

const fontSize = int32(16)

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

	// 	op.window.SetScrollCallback(op.scroll)
	// 	op.window.SetCursorPosCallback(op.cursorPos)
	// 	op.window.SetKeyCallback(op.key)

	if err = gl.Init(); err != nil {
		return
	}

	glfw.SwapInterval(1) // Enable vsync

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

	// DrawText text on the screen
	DrawText := func(cell vl.Cell, x, y int32) {
		gl.Color4ub(0, 0, 0, 255)
		gl.LoadIdentity()
		font.Printf(float32(x), float32(y), string(cell.R))
	}

	// 	op.fps.Init()

	gl.Disable(gl.LIGHTING)

	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()

	// vl demo
	root, action := vl.Demo()
	_ = action

	screen := vl.Screen{
		Root: root,
	}
	var cells [][]vl.Cell

	var widthSymbol uint
	var heightSymbol uint
	var w, h int

	window.SetMouseButtonCallback(func(
		w *glfw.Window,
		button glfw.MouseButton,
		action glfw.Action,
		mods glfw.ModifierKey,
	) {
		// convert button
		var bm tcell.ButtonMask
		switch button {
		case glfw.MouseButton1:
			bm = tcell.Button1
		case glfw.MouseButton2:
			bm = tcell.Button2
		case glfw.MouseButton3:
			bm = tcell.Button3
		default:
			return
		}
		// calculate position
		x, y := w.GetCursorPos()
		xs := int(x / float64(fontSize))
		ys := int(y / float64(fontSize))
		// create event
		switch action {
		case glfw.Press:
			screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
		case glfw.Release:

		default:
			// case glfw.Repeat:
			// do nothing
		}
	})

	for !window.ShouldClose() {
		// windows
		w, h = window.GetSize()
		gl.Viewport(0, 0, int32(w), int32(h))

		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)

		widthSymbol = uint(w) / uint(fontSize)
		heightSymbol = uint(h) / uint(fontSize)
		screen.SetHeight(heightSymbol)
		screen.GetContents(widthSymbol, &cells)
		for r := 0; r < len(cells); r++ {
			if len(cells[r]) == 0 {
				continue
			}
			for c := 0; c < len(cells[r]); c++ {
				DrawText(cells[r][c], int32(c)*fontSize, int32(r)*fontSize)
			}
		}

		// end
		window.MakeContextCurrent()
		window.SwapBuffers()
	}
}

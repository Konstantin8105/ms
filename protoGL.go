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
	const fontSize = int32(16)
	font, err := gltext.LoadTruetype(fd, fontSize, 32, 127, gltext.LeftToRight)
	if err != nil {
		return
	}
	err = fd.Close()
	if err != nil {
		return
	}
	// font is prepared

	color := func(c tcell.Color) (R, G, B float32) {
		switch c {
		case tcell.ColorBlack:
			R, G, B = 0, 0, 0
		case tcell.ColorWhite:
			R, G, B = 1, 1, 1
		case tcell.ColorYellow:
			R, G, B = 1, 1, 0
		case tcell.ColorViolet:
			R, G, B = 0.5, 0, 1.0
		case tcell.ColorMaroon:
			R, G, B = 0.5, 0, 0
		default:
			panic(c)
		}
		return
	}

	// DrawText text on the screen
	DrawText := func(cell vl.Cell, x, y int) {

		// We need to offset each string by the height of the
		// font. To ensure they don't overlap each other.
		gw, gh := font.GlyphBounds()

		x *= gw
		y *= gh

		// prepare colors
		fg, bg, attr := cell.S.Decompose()
		_ = fg
		_ = bg
		_ = attr

		// rectangle
		var vp [4]int32
		gl.GetIntegerv(gl.VIEWPORT, &vp[0])

		gl.PushAttrib(gl.TRANSFORM_BIT)
		gl.MatrixMode(gl.PROJECTION)
		gl.PushMatrix()
		gl.LoadIdentity()
		gl.Ortho(float64(vp[0]), float64(vp[2]), float64(vp[1]), float64(vp[3]), 0, 1)
		gl.PopAttrib()

		gl.PushAttrib(gl.LIST_BIT | gl.CURRENT_BIT | gl.ENABLE_BIT | gl.TRANSFORM_BIT)
		{
			gl.MatrixMode(gl.MODELVIEW)
			gl.Disable(gl.LIGHTING)
			gl.Disable(gl.DEPTH_TEST)
			gl.Enable(gl.BLEND)
			gl.Enable(gl.TEXTURE_2D)

			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
			//gl.BindTexture(gl.TEXTURE_2D, f.texture)
			//gl.ListBase(f.listbase)

			var mv [16]float32
			gl.GetFloatv(gl.MODELVIEW_MATRIX, &mv[0])
			gl.PushMatrix()
			{
				gl.LoadIdentity()

				//mgw := float32(gw) // f.maxGlyphWidth)
				mgh := float32(gh) // f.maxGlyphHeight)

				//switch f.config.Dir {
				//case LeftToRight, TopToBottom:
				gl.Translatef(float32(x), float32(vp[3])-float32(y)-mgh, 0)
				//case RightToLeft:
				//	gl.Translatef(x-mgw, float32(vp[3])-y-mgh, 0)
				//}

				//fmt.Println(cell.S)
				r, g, b := color(bg)
				gl.Color4f(r, g, b, 1)

				gl.Rectf(0, 0, float32(gw), float32(gh))

				gl.MultMatrixf(&mv[0])
				//gl.CallLists(int32(len(indices)), gl.UNSIGNED_INT, unsafe.Pointer(&indices[0]))
			}
			gl.PopMatrix()
		}
		gl.PopAttrib()

		gl.PushAttrib(gl.TRANSFORM_BIT)
		gl.MatrixMode(gl.PROJECTION)
		gl.PopMatrix()
		gl.PopAttrib()

		gl.LoadIdentity()

		str := string(cell.R)

		r, g, b := color(fg)
		gl.Color4f(r, g, b, 1)
		err = font.Printf(float32(x), float32(y), str)
		if err != nil {
			panic(err)
		}
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

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		screen.Event(tcell.NewEventKey(tcell.KeyRune, rune('R'), tcell.ModNone)) // TODO add for arrow
	})

	window.SetScrollCallback(func(w *glfw.Window, xoffset, yoffset float64) {
		x, y := w.GetCursorPos()
		gw, gh := font.GlyphBounds()
		xs := int(x / float64(gw))
		ys := int(y / float64(gh))

		var bm tcell.ButtonMask
		if yoffset < 0 {
			bm = tcell.WheelDown
		}
		if 0 < yoffset {
			bm = tcell.WheelUp
		}
		if xoffset < 0 {
			bm = tcell.WheelLeft
		}
		if 0 < xoffset {
			bm = tcell.WheelRight
		}
		screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
	})

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
		gw, gh := font.GlyphBounds()
		xs := int(x / float64(gw))
		ys := int(y / float64(gh))
		_ = gh
		_ = gw
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

		gl.Color4f(0.9, 0.9, 0.8, 0.9)
		gl.Rectf(-0.5, -0.5, 0.5, 0.5)

		gw, gh := font.GlyphBounds()
		widthSymbol = uint(w) / uint(gw)
		heightSymbol = uint(h) / uint(gh)
		screen.SetHeight(heightSymbol)
		screen.GetContents(widthSymbol, &cells)
		for r := 0; r < len(cells); r++ {
			if len(cells[r]) == 0 {
				continue
			}
			for c := 0; c < len(cells[r]); c++ {
				DrawText(cells[r][c], c, r)
			}
		}

		// end
		window.MakeContextCurrent()
		window.SwapBuffers()
	}
}

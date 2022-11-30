//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

const (
	runeStart = rune(byte(32))
	runeEnd   = rune('■')
)

func init() {
	runtime.LockOSThread()
}

var WindowRatio float64 = 0.5

func main() {
	// initialize
	var root vl.Widget
	var action chan func()

	// vl demo
	root, action = vl.Demo()

	// unicode table
	//	{
	//		var t vl.Text
	//		var str string
	//		for i := runeStart; i < runeEnd; i++ {
	//			str += " " + string(rune(i))
	//		}
	//		t.SetText(str)
	//		var sc vl.Scroll
	//		sc.Root = &t
	//		root = &sc
	//	}

	// run vl widget in OpenGL
	err := Run(root, action)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
}

func Run(root vl.Widget, action chan func()) (err error) {
	//mutex
	var mutex sync.Mutex

	if err = glfw.Init(); err != nil {
		err = fmt.Errorf("failed to initialize glfw: %v", err)
		return
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	var window *glfw.Window
	window, err = glfw.CreateWindow(800, 600, "3D model", nil, nil)
	if err != nil {
		return
	}
	window.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		return
	}

	glfw.SwapInterval(1) // Enable vsync

	// create new Font from given filename (.ttf expected)
	var font *gltext.Font
	{
		var fd *os.File
		fd, err = os.Open("ProggyClean.ttf") // fontfile
		if err != nil {
			return
		}
		const fontSize = int32(16)
		font, err = gltext.LoadTruetype(fd, fontSize, runeStart, runeEnd, gltext.LeftToRight)
		if err != nil {
			return
		}
		err = fd.Close()
		if err != nil {
			return
		}
	}
	gw, gh := font.GlyphBounds()
	// gw -= 3
	// gh -= 5
	// font is prepared

	color := func(c tcell.Color) (R, G, B float32) {
		switch c {
		case tcell.ColorWhite:
			R, G, B = 1, 1, 1
		case tcell.ColorBlack:
			R, G, B = 0, 0, 0
		case tcell.ColorRed:
			R, G, B = 1, 0.3, 0.3
		case tcell.ColorYellow:
			R, G, B = 1, 1, 0
		case tcell.ColorViolet:
			R, G, B = 0.75, 0.90, 0.90 //0.5, 0, 1.0
		case tcell.ColorMaroon:
			R, G, B = 1, 0.5, 0 // 0.5, 0, 0
		default:
			ri, gi, bi := c.RGB()
			return float32(ri), float32(gi), float32(bi)
		}
		return
	}

	// DrawText text on the screen
	DrawText := func(cell vl.Cell, x, y int) {

		// We need to offset each string by the height of the
		// font. To ensure they don't overlap each other.
		x *= gw
		y *= gh

		// prepare colors
		fg, bg, attr := cell.S.Decompose()
		_ = fg
		_ = bg
		_ = attr

		if bg != tcell.ColorWhite {
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
		}

		str := string(cell.R)

		if str != " " {
			r, g, b := color(fg)
			gl.Color4f(r, g, b, 1)
			err = font.Printf(float32(x), float32(y), str)
			if err != nil {
				panic(err)
			}
		}
	}

	gl.Disable(gl.LIGHTING)

	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()

	screen := vl.Screen{
		Root: root,
	}
	var cells [][]vl.Cell

	var widthSymbol uint
	var heightSymbol uint
	var w, h int

	window.SetCharCallback(func(w *glfw.Window, r rune) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action

		// rune limit
		if !(runeStart <= r && r <= runeEnd) {
			return
		}
		screen.Event(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	})

	window.SetScrollCallback(func(w *glfw.Window, xoffset, yoffset float64) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action

		x, y := w.GetCursorPos()
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
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action

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
		xs := int(x / float64(gw))
		ys := int(y / float64(gh))
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
		x := int(float64(w) * WindowRatio)

		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		r, g, b := color(tcell.ColorWhite)
		gl.ClearColor(r, g, b, 1)

		// Opengl
		gl.Viewport(int32(x), 0, int32(w-x), int32(h))
		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()

		gl.Begin(gl.QUADS)
		gl.Color3d(0.8, 0.1, 0.1)
		{
			gl.Vertex2d(-0.99, -0.99)
			gl.Vertex2d(-0.99, +0.99)
			gl.Vertex2d(+0.99, +0.99)
			gl.Vertex2d(+0.99, -0.99)
		}
		gl.End()
		{
			// screen coordinates
			// openGlScreenCoordinate(op.window)
			gl.Disable(gl.DEPTH_TEST)
			gl.Disable(gl.TEXTURE_2D)

			//w, h := window.GetSize()
			//gl.Viewport(0, 0, int32(w), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(0, float64(x), 0, float64(h), float64(-100.0), float64(100.0))

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
		}
		{
			// draw axe coordinates
			// op.drawAxes()
			//w, h := op.window.GetSize()

			s := math.Min(math.Min(50.0, float64(h)/8.0), float64(x))
			b := 5.0 // distance from window border

			centerX := float64(x) - b - s/2.0
			centerY := b + s/2.0
			gl.Begin(gl.QUADS)
			gl.Color3d(0.8, 0.8, 0.8)
			{
				gl.Vertex2d(centerX-s/2, centerY-s/2)
				gl.Vertex2d(centerX+s/2, centerY-s/2)
				gl.Vertex2d(centerX+s/2, centerY+s/2)
				gl.Vertex2d(centerX-s/2, centerY+s/2)
			}
			gl.End()
			gl.LineWidth(1)
			gl.Begin(gl.LINES)
			gl.Color3d(0.1, 0.1, 0.1)
			{
				gl.Vertex2d(centerX-s/2, centerY-s/2)
				gl.Vertex2d(centerX+s/2, centerY-s/2)
				gl.Vertex2d(centerX+s/2, centerY-s/2)
				gl.Vertex2d(centerX+s/2, centerY+s/2)
				gl.Vertex2d(centerX+s/2, centerY+s/2)
				gl.Vertex2d(centerX-s/2, centerY+s/2)
				gl.Vertex2d(centerX-s/2, centerY+s/2)
				gl.Vertex2d(centerX-s/2, centerY-s/2)
			}
			gl.End()

			gl.Translated(centerX, centerY, 0)
			betta := 30.0
			alpha := 10.0
			gl.Rotated(betta, 1.0, 0.0, 0.0)
			gl.Rotated(alpha, 0.0, 1.0, 0.0)
			gl.LineWidth(1)
			gl.Begin(gl.LINES)
			{
				factor := 2.5
				A := s / factor * 1.0 / 4.0
				// X - red
				gl.Color3d(1, 0, 0)
				{
					gl.Vertex3d(0, 0, 0)
					gl.Vertex3d(A*2.0, 0, 0)
					gl.Vertex3d(A*3.0, -A*0.5, 0)
					gl.Vertex3d(A*4.0, +A*0.5, 0)
					gl.Vertex3d(A*4.0, -A*0.5, 0)
					gl.Vertex3d(A*3.0, +A*0.5, 0)
				}
				// Y - green
				gl.Color3d(0, 1, 0)
				{
					gl.Vertex3d(0, 0, 0)
					gl.Vertex3d(0, A*2.0, 0)
					gl.Vertex3d(0, A*3.0, 0)
					gl.Vertex3d(0, A*3.5, 0)
					gl.Vertex3d(0, A*3.5, 0)
					gl.Vertex3d(-A*0.5, A*4.0, 0)
					gl.Vertex3d(0, A*3.5, 0)
					gl.Vertex3d(+A*0.5, A*4.0, 0)
				}
				// Z - blue
				gl.Color3d(0, 0, 1)
				{
					gl.Vertex3d(0, 0, 0)
					gl.Vertex3d(0, 0, A*2.0)
					gl.Vertex3d(0, +0.5*A, A*3.0)
					gl.Vertex3d(0, +0.5*A, A*4.0)
					gl.Vertex3d(0, -0.5*A, A*3.0)
					gl.Vertex3d(0, -0.5*A, A*4.0)
					gl.Vertex3d(0, +0.5*A, A*3.0)
					gl.Vertex3d(0, -0.5*A, A*4.0)
				}
			}
			gl.End()
		}

		// gui
		gl.Viewport(0, 0, int32(x), int32(h))
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		widthSymbol = uint(float64(w) / float64(gw) * WindowRatio)
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
	return
}

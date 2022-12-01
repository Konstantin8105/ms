//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

const (
	runeStart = rune(byte(32))
	runeEnd   = rune(byte(127)) // int32('■'))
)

func init() {
	runtime.LockOSThread()
}

var (
	betta = 30.0
	alpha = 10.0
)

var WindowRatio float64 = 0.5

func main() {
	// initialize
	var root vl.Widget
	var action chan func()

	// vl demo
	root, action = vl.Demo()

	// unicode table
	// {
	// 	var t vl.Text
	// 	var str string
	// 	for i := runeStart; i < runeEnd; i++ {
	// 		str += " " + string(rune(i))
	// 	}
	// 	t.SetText(str)
	// 	var sc vl.Scroll
	// 	sc.Root = &t
	// 	root = &sc
	// }

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
	// TODO : add distance between glyph gw++
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

	var widthSymbol uint
	var heightSymbol uint
	var w, h int

	// DrawText text on the screen
	DrawText := func(cell vl.Cell, x, y int) {
		if x < 0 || y < 0 {
			return
		}

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
			// draw background rectangle
			r, g, b := color(bg)
			gl.Color4f(r, g, b, 1)
			gl.Begin(gl.QUADS)
			{
				gl.Vertex2d(float64(x), float64(h-y))
				gl.Vertex2d(float64(x+gw), float64(h-y))
				gl.Vertex2d(float64(x+gw), float64(h-y-gh))
				gl.Vertex2d(float64(x), float64(h-y-gh))
			}
			gl.End()
		}

		str := string(cell.R)
		r, g, b := color(fg)
		gl.Color4f(r, g, b, 1)

		//	str = strings.ToUpper(str)

		//	gl.Begin(gl.LINES)
		//	id := 35
		//	for a := range symbol[id] {
		//		for b := range symbol[id][a] {
		//			d := symbol[id][a][b]
		//			const fontSize = 16
		//			dx := float64(x)+ float64(fontSize) * float64(d[0])/400
		//			dy := float64(y)+ float64(fontSize) * float64(d[1])/400
		//			gl.Vertex2d(dx,dy )
		//		}
		//	}
		//	gl.End()
		//	return

		switch str {
		// TODO: performance
		// case "+", "|", "-", "=", ">", "<", "[", "]","(",")":
		// 	gl.Begin(gl.LINES)
		// 	{
		// 		gl.Vertex2d(float64(x), float64(h-y-gh/2))
		// 		gl.Vertex2d(float64(x+gw), float64(h-y-gh/2))
		// 		gl.Vertex2d(float64(x+gw/2), float64(h-y))
		// 		gl.Vertex2d(float64(x+gw/2), float64(h-y-gh))
		// 	}
		// 	gl.End()
		case " ":
		// do nothing
		default:
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

	var fps uint64
	start := time.Now()

	for !window.ShouldClose() {
		alpha += 0.2
		betta += 0.05

		{
			// FPS
			if diff := time.Now().Sub(start); 1 < diff.Seconds() {
				fmt.Printf("FPS(%d) ", fps)
				fps = 0
				start = time.Now()
			}
			fps++
		}

		// windows
		w, h = window.GetSize()
		x := int(float64(w) * WindowRatio)

		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		r, g, b := color(tcell.ColorWhite)
		gl.ClearColor(r, g, b, 1)

		if w < 10 || h < 10 {
			// TODO: fix resizing window
			// PROBLEM with text rendering
			continue
		}

		// Opengl

		// 3D model
		{
			gl.Viewport(int32(x), 0, int32(w-x), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()

			var ratio float64
			//if w-x < h {
			ratio = float64(w-x) / float64(h)
			// gl.Ortho(
			// 	(-op.camera.R-op.camera.moveX)+cx, (op.camera.R-op.camera.moveX)+cx,
			// 	(-op.camera.R-op.camera.moveY)/ratio+cy, (op.camera.R-op.camera.moveY)/ratio+cy,
			// 	(-op.camera.R-cz)*Zzoom, (op.camera.R+cz)*Zzoom)
			//} else {
			//	ratio = float64(h) / float64(w)
			// gl.Ortho(
			// 	(-op.camera.R-op.camera.moveX)/ratio+cx, (op.camera.R-op.camera.moveX)/ratio+cx,
			// 	(-op.camera.R-op.camera.moveY)+cy, (op.camera.R-op.camera.moveY)+cy,
			// 	(-op.camera.R-cz)*Zzoom, (op.camera.R+cz)*Zzoom)
			//}

			gl.Ortho(-50*ratio, 50*ratio, -50, 50, float64(-100.0), float64(100.0))

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			gl.Translated(0, 0, 0)
			gl.Rotated(betta, 1.0, 0.0, 0.0)
			gl.Rotated(alpha, 0.0, 1.0, 0.0)
			// cube
			size := 10.0
			gl.Color3d(0.1, 0.7, 0.1)
			gl.Begin(gl.LINES)
			{
				gl.Vertex3d(-size, -size, -size)
				gl.Vertex3d(+size, -size, -size)

				gl.Vertex3d(-size, -size, -size)
				gl.Vertex3d(-size, +size, -size)

				gl.Vertex3d(-size, -size, -size)
				gl.Vertex3d(-size, -size, +size)

				gl.Vertex3d(+size, +size, +size)
				gl.Vertex3d(-size, +size, +size)

				gl.Vertex3d(+size, +size, +size)
				gl.Vertex3d(+size, -size, +size)

				gl.Vertex3d(+size, +size, +size)
				gl.Vertex3d(+size, +size, -size)
			}
			gl.End()

			gl.PointSize(5)
			gl.Color3d(0.2, 0.8, 0.5)
			gl.Begin(gl.POINTS)
			{
				gl.Vertex3d(-size, -size, -size)
				gl.Vertex3d(+size, -size, -size)
				gl.Vertex3d(-size, +size, -size)
				gl.Vertex3d(-size, -size, +size)
				gl.Vertex3d(+size, +size, -size)
				gl.Vertex3d(+size, -size, +size)
				gl.Vertex3d(-size, +size, +size)
				gl.Vertex3d(+size, +size, +size)
			}
			gl.End()

			DrawSpiral()

		}
		// Axes
		{
			gl.Viewport(int32(x), 0, int32(w-x), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(0, float64(x), 0, float64(h), float64(-100.0), float64(100.0))

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			// draw axe coordinates
			// op.drawAxes()
			//w, h := op.window.GetSize()

			b := 5.0 // distance from window border
			s := math.Min(math.Min(float64(x)-b, float64(h)-b), 50.0)

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
		{
			// separator
			gl.Viewport(int32(x), 0, int32(w-x), int32(h))
			// gl.MatrixMode(gl.PROJECTION)
			// gl.LoadIdentity()
			gl.Ortho(0, float64(x), 0, float64(h), float64(-1000000.0), float64(1000000.0))

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			// separator
			gl.Color3d(0.9, 0.1, 0.1)
			gl.LineStipple(1, 0xAAAA)
			gl.Enable(gl.LINE_STIPPLE)
			gl.Begin(gl.LINES)
			{
				gl.Vertex2d(0, 0)
				gl.Vertex2d(0, float64(h))
			}
			gl.End()
			gl.Disable(gl.LINE_STIPPLE)
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

func DrawSpiral() {
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 800
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)
	for i := 0; i < int(levels); i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0

		bc0 := float32(Ri * math.Sin(angle))
		bc1 := float32(float64(i) * dy)
		bc2 := float32(Ri * math.Cos(angle))

		c := 0.1 + float32(i)/(float32(levels)*1.2)
		gl.Begin(gl.POINTS)
		gl.Color3f(c, 0.4, 0.1)
		gl.Vertex3f(bc0, bc1, bc2)
		gl.End()

		fc0 := float32(Ro * math.Sin(angle))
		fc1 := float32(float64(i) * dy)
		fc2 := float32(Ro * math.Cos(angle))

		gl.Begin(gl.POINTS)
		gl.Color3f(0.1, c, 0.4)
		gl.Vertex3f(fc0, fc1, fc2)
		gl.End()

		gl.Begin(gl.LINES)
		gl.Color3f(0.4, 0.1, c)
		gl.Vertex3f(bc0, bc1, bc2)
		gl.Vertex3f(fc0, fc1, fc2)
		gl.End()
	}
}

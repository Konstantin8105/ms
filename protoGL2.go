//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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

var WindowRatio float64 = 0.4

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

	f := new(Font)
	v := NewVl(root, f)

	// run vl widget in OpenGL
	err := Run(v, action, func() {
		f.Init()
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
}

type Vl struct {
	screen vl.Screen
	cells  [][]vl.Cell
	font   *Font
}

func NewVl(root vl.Widget, f *Font) (v *Vl) {
	v = new(Vl)
	v.screen = vl.Screen{
		Root: root,
	}
	v.font = f
	return
}

func (v *Vl) Draw(w, h int) {
	widthSymbol := uint(float64(w) / float64(v.font.gw))
	heightSymbol := uint(h) / uint(v.font.gh)
	v.screen.SetHeight(heightSymbol)
	v.screen.GetContents(widthSymbol, &v.cells)
	for r := 0; r < len(v.cells); r++ {
		if len(v.cells[r]) == 0 {
			continue
		}
		for c := 0; c < len(v.cells[r]); c++ {
			v.font.DrawText(v.cells[r][c], c, r, h)
		}
	}
}

type Font struct {
	font   *glsymbol.Font
	gw, gh int
}

func (f *Font) Init() (err error) {
	f.font, err = glsymbol.DefaultFont()
	if err != nil {
		panic(err)
	}
	// f.gw = int(f.font.MaxGlyphWidth)
	w, h := f.font.GlyphBounds()
	f.gw, f.gh = int(w), int(h)

	// fmt.Println(f.gw, f.gh)
	// 	// create new Font from given filename (.ttf expected)
	// 	fd, err := os.Open("ProggyClean.ttf") // fontfile
	// 	if err != nil {
	// 		return
	// 	}
	// 	const fontSize = int32(16)
	// 	ft, err := gltext.LoadTruetype(fd, fontSize, runeStart, runeEnd, gltext.LeftToRight)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	f.font = ft
	// 	err = fd.Close()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	f.gw, f.gh = f.font.GlyphBounds()
	// 	f.gw++ // add distance between glyph
	return
}

// DrawText text on the screen
func (f Font) DrawText(cell vl.Cell, x, y, h int) {
	if x < 0 || y < 0 {
		return
	}

	x *= f.gw
	y *= f.gh

	// prepare colors
	fg, bg, attr := cell.S.Decompose()
	_ = fg
	_ = bg
	_ = attr

	if bg != tcell.ColorWhite {
		r, g, b := color(bg)
		gl.Color4f(r, g, b, 1)
		gl.Rectf(float32(x), float32(h-y-f.gh), float32(x+f.gw), float32(h-y))
	}

	if cell.R == ' ' {
		return
	}
	r, g, b := color(fg)
	gl.Color4f(r, g, b, 1)
	i := int(byte(cell.R)) - int(f.font.Config.Low)
	gl.RasterPos2i(int32(x), int32(h-y-f.gh))
	gl.Bitmap(
		f.font.Config.Glyphs[i].Width, f.font.Config.Glyphs[i].Height,
		0.0, 0.0,
		0.0, 0.0,
		(*uint8)(gl.Ptr(&f.font.Config.Glyphs[i].BitmapData[0])),
	)
	// return checkGLError()
}

func color(c tcell.Color) (R, G, B float32) {
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

func Run(v *Vl, action chan func(), init func()) (err error) {
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

	if init != nil {
		init()
	}

	// 	var widthSymbol uint
	// 	var heightSymbol uint
	var w, h int

	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()

	window.SetCharCallback(func(w *glfw.Window, r rune) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action

		// rune limit
		if !(runeStart <= r && r <= runeEnd) {
			return
		}
		v.screen.Event(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	})

	window.SetScrollCallback(func(w *glfw.Window, xoffset, yoffset float64) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action

		x, y := w.GetCursorPos()
		xs := int(x / float64(v.font.gw))
		ys := int(y / float64(v.font.gh))

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
		v.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
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
		xs := int(x / float64(v.font.gw))
		ys := int(y / float64(v.font.gh))
		// create event
		switch action {
		case glfw.Press:
			v.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
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
		betta += 0.25

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
			ratio = float64(w-x) / float64(h)
			ymax := 0.2 * 800
			gl.Ortho(-500*ratio, 500*ratio, -500, 500, float64(-ymax), float64(ymax))

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			drawRight()
		}
		{
			// gui
			gl.Viewport(0, 0, int32(x), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(0, float64(x), 0, float64(h), -1.0, 1.0)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			v.Draw(x, h)
		}
		{
			// separator
			gl.Viewport(0, 0, int32(x), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			gl.Ortho(0, float64(x), 0, float64(h), -1.0, 1.0)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			gl.Color3f(0.7, 0.7, 0.7)
			gl.Begin(gl.LINES)
			gl.Vertex3f(float32(x), 0, 0)
			gl.Vertex3f(float32(x), float32(h), 0)
			gl.End()
		}
		// end
		window.MakeContextCurrent()
		window.SwapBuffers()
	}
	return
}

func drawRight() {
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

	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 8000
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

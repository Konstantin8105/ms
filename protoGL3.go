//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"

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

func main() {
	// initialize
	var root vl.Widget

	// vl demo
	root, _ = vl.Demo()

	v := NewVl(root)

	// run vl widget in OpenGL
	if err := Run(v); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
}

//===========================================================================//

type Font struct {
	font *glsymbol.Font
}

func (f Font) GetSymbolSize() (gw, gh int) {
	return int(f.font.MaxGlyphWidth), int(f.font.MaxGlyphHeight)
}

// DrawText text on the screen
func (f Font) DrawText(cell vl.Cell, x, y, h int) {
	if x < 0 || y < 0 {
		return
	}

	gw, gh := f.GetSymbolSize()

	x *= gw
	y *= gh

	// prepare colors
	fg, bg, attr := cell.S.Decompose()
	_ = fg
	_ = bg
	_ = attr

	if bg != tcell.ColorWhite {
		r, g, b := color(bg)
		gl.Color4f(r, g, b, 1)
		gl.Rectf(float32(x), float32(h-y-gh), float32(x+gw), float32(h-y))
	}

	if cell.R == ' ' {
		return
	}
	r, g, b := color(fg)
	gl.Color4f(r, g, b, 1)
	i := int(byte(cell.R)) - int(f.font.Config.Low)
	gl.RasterPos2i(int32(x), int32(h-y-gh))
	gl.Bitmap(
		f.font.Config.Glyphs[i].Width, f.font.Config.Glyphs[i].Height,
		0.0, 0.0,
		0.0, 0.0,
		(*uint8)(gl.Ptr(&f.font.Config.Glyphs[i].BitmapData[0])),
	)
	// return checkGLError()
}

//===========================================================================//

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

//===========================================================================//

func Run(v *Vl) (err error) {
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

	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()

	var font Font
	font.font, err = glsymbol.DefaultFont()
	if err != nil {
		return
	}

	// window ratio
	const windowRatio float64 = 0.4

	// dimensions
	var w, h, split int

	// windows prepared
	windows := [2]Window{new(Vl), new(Opengl)}
	windows[0] = v
	var focus uint
	for i := range windows {
		windows[i].SetFont(&font)
	}

	// windows input data
	window.SetCharCallback(func(w *glfw.Window, r rune) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action
		if len(windows) <= int(focus) {
			focus = 0
		}
		windows[focus].CharCallback(w, r)
	})
	window.SetScrollCallback(func(w *glfw.Window, xoffset, yoffset float64) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		// calculate position
		x, _ := w.GetCursorPos()
		// create event
		if int(x) < split {
			focus = 0
		} else {
			focus = 1
		}
		//action
		if len(windows) <= int(focus) {
			focus = 0
		}
		windows[focus].ScrollCallback(w, xoffset, yoffset)
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
		// calculate position
		x, _ := w.GetCursorPos()
		// create event
		if action == glfw.Press {
			if int(x) < split {
				focus = 0
			} else {
				focus = 1
			}
		}
		//action
		if len(windows) <= int(focus) {
			focus = 0
		}
		windows[focus].MouseButtonCallback(w, button, action, mods)
	})
	window.SetKeyCallback(func(
		w *glfw.Window,
		key glfw.Key,
		scancode int,
		action glfw.Action,
		mods glfw.ModifierKey) {
		//mutex
		mutex.Lock()
		defer mutex.Unlock()
		//action
		if len(windows) <= int(focus) {
			focus = 0
		}
		windows[focus].KeyCallback(w, key, scancode, action, mods)
	})

	// draw
	for !window.ShouldClose() {
		// windows
		w, h = window.GetSize()
		split = int(float64(w) * windowRatio)

		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		r, g, b := color(tcell.ColorWhite)
		gl.ClearColor(r, g, b, 1)

		if w < 10 || h < 10 {
			// no need to do
			// problem with text rendering
			continue
		}

		{
			// gui
			gl.Viewport(0, 0, int32(split), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()

			windows[0].Draw(split, h)
		}
		{
			// 3D model
			gl.Viewport(int32(split), 0, int32(w-split), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()

			windows[1].Draw(split, h)
		}
		{
			// separator
			gl.Viewport(0, 0, int32(split), int32(h))
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()

			gl.Ortho(0, float64(split), 0, float64(h), -1.0, 1.0)
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
			gl.Color3f(0.7, 0.7, 0.7)
			gl.Begin(gl.LINES)
			gl.Vertex3f(float32(split), 0, 0)
			gl.Vertex3f(float32(split), float32(h), 0)
			gl.End()
		}
		// end
		window.MakeContextCurrent()
		window.SwapBuffers()
	}
	return
}

//===========================================================================//

var _ Window = new(Vl)

type Vl struct {
	font *Font

	screen vl.Screen
	cells  [][]vl.Cell
}

func (vl *Vl) SetFont(f *Font) {
	vl.font = f
}

func NewVl(root vl.Widget) (v *Vl) {
	v = new(Vl)
	v.screen = vl.Screen{
		Root: root,
	}
	return
}

func (v *Vl) Draw(w, h int) {
	gl.Ortho(0, float64(w), 0, float64(h), -1.0, 1.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gw, gh := v.font.GetSymbolSize()

	widthSymbol := uint(float64(w) / float64(gw))
	heightSymbol := uint(h) / uint(gh)
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

func (vl *Vl) CharCallback(w *glfw.Window, r rune) {
	fmt.Printf("%p char %v\n", vl, r)
	// rune limit
	if !((runeStart <= r && r <= runeEnd) || r == rune('\n')) {
		return
	}
	vl.screen.Event(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
}
func (vl *Vl) ScrollCallback(w *glfw.Window, xoffset, yoffset float64) {
	fmt.Printf("%p scroll %v %v\n", vl, xoffset, yoffset)

	gw, gh := vl.font.GetSymbolSize()

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
	vl.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
}
func (vl *Vl) MouseButtonCallback(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	fmt.Printf("%p mouse %v %v %v\n", vl, button, action, mods)

	gw, gh := vl.font.GetSymbolSize()

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
		vl.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
	case glfw.Release:

	default:
		// case glfw.Repeat:
		// do nothing
	}
}
func (vl *Vl) KeyCallback(
	w *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	fmt.Printf("%p key %v %v %v %v\n", vl, key, scancode, action, mods)
	if action != glfw.Press {
		return
	}
	switch key {
	case glfw.KeyUp:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyUp, rune(' '), tcell.ModNone))
	case glfw.KeyDown:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyDown, rune(' '), tcell.ModNone))
	case glfw.KeyLeft:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyLeft, rune(' '), tcell.ModNone))
	case glfw.KeyRight:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyRight, rune(' '), tcell.ModNone))
	case glfw.KeyEnter:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyEnter, rune('\n'), tcell.ModNone))
	case glfw.KeyBackspace:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyBackspace, rune(' '), tcell.ModNone))
	case glfw.KeyDelete:
		vl.screen.Event(tcell.NewEventKey(tcell.KeyDelete, rune(' '), tcell.ModNone))
	default:
		// do nothing
	}
}

//===========================================================================//

var _ Window = new(Opengl)

type Opengl struct {
	font *Font

	betta float64
	alpha float64
}

func (op *Opengl) SetFont(f *Font) {
	op.font = f
}

func (op *Opengl) Draw(w, h int) {
	var ratio float64
	ratio = float64(w) / float64(h)
	ymax := 0.2 * 800
	gl.Ortho(-50*ratio, 50*ratio, -50, 50, float64(-ymax), float64(ymax))

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	// gl.Translated(0, 0, 0)
	gl.Rotated(op.betta, 1.0, 0.0, 0.0)
	gl.Rotated(op.alpha, 0.0, 1.0, 0.0)
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
		levels = 80
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

func (op *Opengl) CharCallback(w *glfw.Window, r rune) {
	fmt.Printf("%p char %v\n", op, r)
}
func (op *Opengl) ScrollCallback(w *glfw.Window, xoffset, yoffset float64) {
	fmt.Printf("%p scroll %v %v\n", op, xoffset, yoffset)
}
func (op *Opengl) MouseButtonCallback(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	fmt.Printf("%p mouse %v %v %v\n", op, button, action, mods)
}
func (op *Opengl) KeyCallback(
	w *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	fmt.Printf("%p key %v %v %v %v\n", op, key, scancode, action, mods)
	if action != glfw.Press {
		return
	}
	switch key {
	case glfw.KeyUp:
		op.alpha += 2
	case glfw.KeyDown:
		op.alpha -= 2
	case glfw.KeyLeft:
		op.betta += 2
	case glfw.KeyRight:
		op.betta -= 2
	default:
		// do nothing
	}
}

//===========================================================================//

type Window interface {
	SetFont(font *Font)
	Draw(w, h int)
	CharCallback(w *glfw.Window, r rune)
	ScrollCallback(w *glfw.Window, xoffset, yoffset float64)
	MouseButtonCallback(
		w *glfw.Window,
		button glfw.MouseButton,
		action glfw.Action,
		mods glfw.ModifierKey,
	)
	KeyCallback(
		w *glfw.Window,
		key glfw.Key,
		scancode int,
		action glfw.Action,
		mods glfw.ModifierKey,
	)
}

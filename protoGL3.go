//go:build ignore

package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/tf"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	syncPoint := make(chan func(), 10)

	var m Model
	m.Value = 10

	var u Undo
	u.actual = &m
	u.syncPoint = &syncPoint

	// run vl widget in OpenGL
	if err := Run(&u, &syncPoint); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
}

func InputUnsigned(prefix, postfix string, defValue uint) (
	w vl.Widget,
	gettext func() (_ uint, ok bool),
) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText(fmt.Sprintf("%d", defValue))
	in.Filter(tf.UnsignedInteger)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, func() (uint, bool) {
		text := in.GetText()
		value, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(value), true
	}
}

type Model struct {
	Value uint
}

func (m *Model) Change(value uint) {
	m.Value = value
}

func (m *Model) GetValue() uint {
	return m.Value
}

type Undo struct {
	list      *list.List
	syncPoint *chan func()
	log       []string
	actual    *Model
}

func (u *Undo) addToUndo() {
	b, err := json.Marshal(u.actual)
	if err != nil {
		u.AddLog(fmt.Sprintf("%v", err))
		return
	}
	if u.list == nil {
		u.list = list.New()
		u.addToUndo() // store
		return
	}
	u.AddLog("Add to undo")
	u.list.PushBack(b)
}

func (u *Undo) Change(value uint) {
	(*u.syncPoint) <- func() {
		u.addToUndo()
		// action
		u.actual.Change(value)
	}
}

func (u *Undo) GetValue() uint {
	return u.actual.GetValue()
}
func (u *Undo) AddLog(log string) {
	u.log = append(u.log, log)
}
func (u *Undo) GetLog() []string {
	return u.log
}

func (u *Undo) Undo() {
	(*u.syncPoint) <- func() {
		// u.addToUndo() // No need
		// action
		if u.list == nil {
			u.AddLog("undo list is empty")
			return
		}
		el := u.list.Back()
		if el == nil {
			u.AddLog("undo list back is empty")
			return
		}
		var last Model
		b := el.Value.([]byte)
		if err := json.Unmarshal(b, &last); err != nil {
			u.AddLog(fmt.Sprintf("json: %v", err))
			return
		}
		// undo model
		u.actual = &last
		// remove
		u.list.Remove(el)
	}
}

type Changable interface {
	Change(value uint)
	GetValue() uint
	AddLog(log string)
	GetLog() []string
	Undo()
}

var _ Changable = new(Undo)

//===========================================================================//

type Font struct {
	font *glsymbol.Font
}

func (f Font) GetRunes() (low, high rune) {
	return f.font.Config.Low, f.font.Config.High
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

func Run(ch Changable, syncPoint *chan func()) (err error) {
	//mutex
	var mutex sync.Mutex // TODO change to syncPoint

	if err = glfw.Init(); err != nil {
		err = fmt.Errorf("failed to initialize glfw: %v", err)
		return
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	var window *glfw.Window
	window, err = glfw.CreateWindow(600, 300, "3D model", nil, nil)
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

	vl := NewVl(func() vl.Widget {
		var list vl.List

		r, rgt := InputUnsigned("Amount levels", "", ch.GetValue())
		list.Add(r)

		var actual vl.Text
		list.Add(&actual)
		go func() {
			for {
				time.Sleep(time.Millisecond * 100)
				(*syncPoint) <- func() {
					actual.SetText(fmt.Sprintf("Size: %03d", ch.GetValue()))
				}
			}
		}()

		var b vl.Button
		b.SetText("Add")
		b.OnClick = func() {
			n, ok := rgt()
			if !ok {
				return
			}
			if n < 1 {
				return
			}
			ch.Change(n)
		}
		list.Add(&b)

		var back vl.Button
		back.SetText("Undo")
		back.OnClick = func() {
			ch.Undo()
		}
		list.Add(&back)

		var logs vl.Text
		list.Add(&logs)
		go func() {
			for {
				time.Sleep(time.Second)
				(*syncPoint) <- func() {
					logs.SetText(strings.Join(ch.GetLog(), "\n"))
				}
			}
		}()

		return &list
	}())

	windows := [2]Window{
		vl,
		new(Opengl),
	}
	for i := range windows {
		windows[i].SetModel(ch)
	}

	// dimensions
	var w, h, split int

	// windows prepared
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
		// sync
		select {
		case f := <-*syncPoint:
			f()
		default:
		}

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
			gl.LineWidth(1)
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
	ch   Changable

	screen vl.Screen
	cells  [][]vl.Cell
}

func NewVl(root vl.Widget) (v *Vl) {
	v = new(Vl)
	v.screen = vl.Screen{
		Root: &vl.Scroll{
			Root: root,
		},
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

func (v *Vl) SetFont(f *Font) {
	v.font = f
}
func (v *Vl) SetModel(ch Changable) {
	v.ch = ch
}

func (v *Vl) CharCallback(w *glfw.Window, r rune) {
	// fmt.Printf("%p char %v\n", vl, r)
	// rune limit
	runeStart, runeEnd := v.font.GetRunes()
	if !((runeStart <= r && r <= runeEnd) || r == rune('\n')) {
		return
	}
	v.screen.Event(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
}

func (v *Vl) ScrollCallback(w *glfw.Window, xoffset, yoffset float64) {
	// fmt.Printf("%p scroll %v %v\n", vl, xoffset, yoffset)

	gw, gh := v.font.GetSymbolSize()

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
	v.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
}

func (v *Vl) MouseButtonCallback(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	// fmt.Printf("%p mouse %v %v %v\n", vl, button, action, mods)

	gw, gh := v.font.GetSymbolSize()

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
		v.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
	case glfw.Release:

	default:
		// case glfw.Repeat:
		// do nothing
	}
}

func (v *Vl) KeyCallback(
	w *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	// fmt.Printf("%p key %v %v %v %v\n", v, key, scancode, action, mods)
	if action != glfw.Press {
		return
	}
	run := func(k tcell.Key, ch rune, mod tcell.ModMask) {
		v.screen.Event(tcell.NewEventKey(k, ch, mod))
	}
	switch key {
	case glfw.KeyUp:
		run(tcell.KeyUp, rune(' '), tcell.ModNone)
	case glfw.KeyDown:
		run(tcell.KeyDown, rune(' '), tcell.ModNone)
	case glfw.KeyLeft:
		run(tcell.KeyLeft, rune(' '), tcell.ModNone)
	case glfw.KeyRight:
		run(tcell.KeyRight, rune(' '), tcell.ModNone)
	case glfw.KeyEnter:
		run(tcell.KeyEnter, rune('\n'), tcell.ModNone)
	case glfw.KeyBackspace:
		run(tcell.KeyBackspace, rune(' '), tcell.ModNone)
	case glfw.KeyDelete:
		run(tcell.KeyDelete, rune(' '), tcell.ModNone)
	default:
		// do nothing
	}
}

//===========================================================================//

var _ Window = new(Opengl)

type Opengl struct {
	font *Font
	ch   Changable

	betta float64
	alpha float64
}

func (op *Opengl) Draw(w, h int) {
	ratio := float64(w) / float64(h)
	ymax := 0.2 * 800
	gl.Ortho(-50*ratio, 50*ratio, -50, 50, float64(-ymax), float64(ymax))

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	// gl.Translated(0, 0, 0)
	gl.Rotated(op.betta, 1.0, 0.0, 0.0)
	gl.Rotated(op.alpha, 0.0, 1.0, 0.0)
	// cube
	size := 10.0
	gl.LineWidth(2)
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
		levels = op.ch.GetValue()
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

func (op *Opengl) SetFont(f *Font) {
	op.font = f
}

func (op *Opengl) SetModel(ch Changable) {
	op.ch = ch
}

func (op *Opengl) CharCallback(w *glfw.Window, r rune) {
	// fmt.Printf("%p char %v\n", op, r)
}

func (op *Opengl) ScrollCallback(w *glfw.Window, xoffset, yoffset float64) {
	// fmt.Printf("%p scroll %v %v\n", op, xoffset, yoffset)
}

func (op *Opengl) MouseButtonCallback(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	// fmt.Printf("%p mouse %v %v %v\n", op, button, action, mods)
}

func (op *Opengl) KeyCallback(
	w *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	// fmt.Printf("%p key %v %v %v %v\n", op, key, scancode, action, mods)
	if action != glfw.Press {
		return
	}
	const angle = 5.0
	switch key {
	case glfw.KeyLeft:
		op.alpha += angle
	case glfw.KeyRight:
		op.alpha -= angle
	case glfw.KeyUp:
		op.betta += angle
	case glfw.KeyDown:
		op.betta -= angle
	default:
		// do nothing
	}
}

//===========================================================================//

// Window interface for important implementation
type Window interface {
	SetFont(font *Font)
	SetModel(ch Changable)
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

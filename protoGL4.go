package main

import (
	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type TuiWindow struct {
	font   *glsymbol.Font
	screen vl.Screen
	cells  [][]vl.Cell
}

func (t *TuiWindow) SetMouseButtonCallback(
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
	xcursor, ycursor float64,
) {
	gw := int(t.font.MaxGlyphWidth)
	gh := int(t.font.MaxGlyphHeight)

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
	xs := int(xcursor / float64(gw))
	ys := int(ycursor / float64(gh))
	// create event
	switch action {
	case glfw.Press:
		t.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
	case glfw.Release:

	default:
		// case glfw.Repeat:
		// do nothing
	}
	// })
}
func (t *TuiWindow) SetCharCallback(r rune) {
	// rune limit
	runeStart, runeEnd := t.font.Config.Low, t.font.Config.High
	if !((runeStart <= r && r <= runeEnd) || r == rune('\n')) {
		return
	}
	t.screen.Event(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
}
func (t *TuiWindow) SetKeyCallback(
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	if action != glfw.Press {
		return
	}
	run := func(k tcell.Key, ch rune, mod tcell.ModMask) {
		t.screen.Event(tcell.NewEventKey(k, ch, mod))
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
func (t *TuiWindow) SetScrollCallback(
	xcursor, ycursor float64,
	xoffset, yoffset float64,
) {
	//mutex
	// mutex.Lock()
	// defer mutex.Unlock()
	//action

	gw := int(t.font.MaxGlyphWidth)
	gh := int(t.font.MaxGlyphHeight)

	xs := int(xcursor / float64(gw))
	ys := int(ycursor / float64(gh))

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
	t.screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
}
func (t *TuiWindow) Draw(x, y, w, h int32) {
	//
	// 			gl.Viewport(0, 0, int32(x), int32(h))
	// 			gl.MatrixMode(gl.PROJECTION)
	// 			gl.LoadIdentity()
	// 			gl.Ortho(0, float64(x), 0, float64(h), -1.0, 1.0)
	// 			gl.MatrixMode(gl.MODELVIEW)
	// 			gl.LoadIdentity()
	//
	gl.Viewport(int32(x), int32(y), int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	gl.Ortho(0, float64(w), 0, float64(h), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	widthSymbol := uint(float64(w) / float64(t.font.MaxGlyphWidth))
	heightSymbol := uint(h) / uint(t.font.MaxGlyphHeight)
	t.screen.SetHeight(heightSymbol)
	t.screen.GetContents(widthSymbol, &t.cells)
	for r := 0; r < len(t.cells); r++ {
		if len(t.cells[r]) == 0 {
			continue
		}
		for c := 0; c < len(t.cells[r]); c++ {
			t.DrawText(t.cells[r][c], c, r, int(h))
		}
	}
}

// DrawText text on the screen
func (t *TuiWindow) DrawText(cell vl.Cell, x, y, h int) {
	if x < 0 || y < 0 {
		return
	}

	gw := int(t.font.MaxGlyphWidth)
	gh := int(t.font.MaxGlyphHeight)

	x *= int(gw)
	y *= int(gh)

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
	i := int(byte(cell.R)) - int(t.font.Config.Low)
	gl.RasterPos2i(int32(x), int32(h-y-gh))
	gl.Bitmap(
		t.font.Config.Glyphs[i].Width, t.font.Config.Glyphs[i].Height,
		0.0, 0.0,
		0.0, 0.0,
		(*uint8)(gl.Ptr(&t.font.Config.Glyphs[i].BitmapData[0])),
	)
}

func main() {
	var ws [2]ds.Window

	w0, a0 := vl.Demo()
	tw0 := TuiWindow{screen: vl.Screen{Root: &vl.Scroll{Root: w0}}}
	ws[0] = &tw0

	w1, a1 := vl.Demo()
	tw1 := TuiWindow{screen: vl.Screen{Root: &vl.Scroll{Root: w1}}}
	ws[1] = &tw1

	ch := make(chan func(), 1000)
	go func() {
		for a := range a0 {
			ch <- a
		}
	}()
	go func() {
		for a := range a1 {
			ch <- a
		}
	}()

	screen, err := ds.New("Demo", ws, &ch)
	if err != nil {
		panic(err)
	}

	// add fonts
	f, err := glsymbol.DefaultFont()
	if err != nil {
		panic(err)
	}
	tw0.font = f
	tw1.font = f

	screen.Run()
}

const (
	runeStart = rune(byte(32))
	runeEnd   = rune(byte(127)) // int32('■'))
)

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

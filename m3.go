package ms

import (
	"fmt"
	_ "image/png"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

func init() {
	runtime.LockOSThread()
}

var font *gltext.Font

const fontSize = int32(12)

type Opengl struct {
	window *glfw.Window

	model *Model

	// for 3d view
	state       viewState
	cursorLeft  viewState
	updateModel bool
	camera      struct {
		alpha, betta float64
		R            float64
		center       Coordinate
		moveX, moveY float64
	}

	// calculate data for FPS
	fps Fps
}

func (op *Opengl) Init() {
	op.updateModel = true
	if op.state != normal && op.state != colorEdgeElements {
		op.state = normal
	}
	op.cursorLeft = selectPoints
}

func NewOpengl() (op *Opengl, err error) {
	op = new(Opengl)
	op.Init()

	if err = glfw.Init(); err != nil {
		err = fmt.Errorf("failed to initialize glfw: %v", err)
		return
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	op.window, err = glfw.CreateWindow(800, 600, "3D model", nil, nil)
	if err != nil {
		return
	}
	op.window.MakeContextCurrent()

	op.window.SetMouseButtonCallback(op.mouseButton)
	op.window.SetScrollCallback(op.scroll)
	op.window.SetCursorPosCallback(op.cursorPos)
	op.window.SetKeyCallback(op.key)

	if err = gl.Init(); err != nil {
		return
	}

	glfw.SwapInterval(1) // Enable vsync

	// ???

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

	op.fps.Init()

	gl.Disable(gl.LIGHTING)

	return
}

func (op *Opengl) Run() {
	defer glfw.Terminate()
	for !op.window.ShouldClose() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)

		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.LINE_SMOOTH)

		op.cameraView()
		op.model3d(op.state)

		// check select rectangle
		if selectObjects.fromAdd && selectObjects.toAdd {
			op.selectByRectangle()
			selectObjects.fromAdd = false
			selectObjects.toAdd = false
			selectObjects.toUpdate = false
			continue
		}

		// screen coordinates
		openGlScreenCoordinate(op.window)
		// select rectangle
		drawSelectRectangle(op.window)
		// draw axe coordinates
		op.drawAxes()
		// minimal screen notes
		DrawText(fmt.Sprintf("FPS       : %6.2f", op.fps.Get()), 0, 0*fontSize)
		if op.model != nil {
			DrawText(fmt.Sprintf("Nodes     : %6d", len(op.model.Coords)), 0, 1*fontSize)
			DrawText(fmt.Sprintf("Elements  : %6d", len(op.model.Elements)), 0, 2*fontSize)
		}

		// TODO : REMOVE: gl.Disable(gl.DEPTH_TEST)
		// TODO : REMOVE: ui(window)

		op.window.MakeContextCurrent()
		op.window.SwapBuffers()

		op.fps.EndFrame()
	}
}

type Fps struct {
	framesCount int64
	framesTime  time.Time
	last        float32
}

func (f *Fps) Init() {
	f.framesTime = time.Now()
}

func (f *Fps) Get() float32 {
	ms := time.Now().Sub(f.framesTime).Milliseconds()
	if ms < 1000 && f.framesCount < 100 {
		return f.last
	}
	f.last = float32(f.framesCount) / float32(ms) * 1000.0
	f.framesCount = 0
	f.framesTime = time.Now()
	return f.last
}

func (f *Fps) EndFrame() {
	f.framesCount++
}

// Draw text on the screen
func DrawText(str string, x, y int32) {
	gl.Color4ub(0, 0, 0, 255)
	gl.LoadIdentity()
	font.Printf(float32(x), float32(y), str)
}

func angle_norm(a float64) float64 {
	if 360.0 < a {
		return a - 360.0
	}
	if a < -360.0 {
		return a + 360.0
	}
	return a
}

func (op *Opengl) cameraView() {
	// better angle value
	op.camera.alpha = angle_norm(op.camera.alpha)
	op.camera.betta = angle_norm(op.camera.betta)

	w, h := op.window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	ratio := 1.0
	{
		// for avoid 3D cutting back model
		const Zzoom float64 = 100.0
		// renaming
		cx := op.camera.center.X
		cy := op.camera.center.Y
		cz := op.camera.center.Z
		// scaling monitor 3d model on screen
		if w < h {
			ratio = float64(w) / float64(h)
			gl.Ortho(
				(-op.camera.R-op.camera.moveX)+cx, (op.camera.R-op.camera.moveX)+cx,
				(-op.camera.R-op.camera.moveY)/ratio+cy, (op.camera.R-op.camera.moveY)/ratio+cy,
				(-op.camera.R-cz)*Zzoom, (op.camera.R+cz)*Zzoom)
		} else {
			ratio = float64(h) / float64(w)
			gl.Ortho(
				(-op.camera.R-op.camera.moveX)/ratio+cx, (op.camera.R-op.camera.moveX)/ratio+cx,
				(-op.camera.R-op.camera.moveY)+cy, (op.camera.R-op.camera.moveY)+cy,
				(-op.camera.R-cz)*Zzoom, (op.camera.R+cz)*Zzoom)
		}
	}
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Translated(op.camera.center.X, op.camera.center.Y, op.camera.center.Z)
	gl.Rotated(op.camera.betta, 1.0, 0.0, 0.0)
	gl.Rotated(op.camera.alpha, 0.0, 1.0, 0.0)
	gl.Translated(-op.camera.center.X, -op.camera.center.Y, -op.camera.center.Z)

	// minimal R
	if op.camera.R < 0.1 {
		op.camera.R = 0.1
	}
}

func (op *Opengl) model3d(s viewState) {
	if op.model == nil {
		return
	}
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			Debug = append(Debug, fmt.Sprintf("%v\n%v", r, string(debug.Stack())))
	// 		}
	// 	}()

	gl.PushMatrix()
	defer func() {
		gl.PopMatrix()
	}()

	if op.updateModel {
		op.updateModel = false

		// Do not update angles
		// angle in global plate XOZ
		// camera.alpha = 0.0
		// angle in global plate XOY
		// camera.betta = 0.0

		// distance from center to camera
		op.camera.R = 1.0
		if len(op.model.Coords) == 0 {
			return
		}
		// renaming
		ps := op.model.Coords
		// calculate radius
		var (
			xmin = ps[0].X
			xmax = ps[0].X
			ymin = ps[0].Y
			ymax = ps[0].Y
			zmin = ps[0].Z
			zmax = ps[0].Z
		)
		for i := range ps {
			xmin = math.Min(xmin, ps[i].X)
			ymin = math.Min(ymin, ps[i].Y)
			zmin = math.Min(zmin, ps[i].Z)
			xmax = math.Max(xmax, ps[i].X)
			ymax = math.Max(ymax, ps[i].Y)
			zmax = math.Max(zmax, ps[i].Z)
		}
		op.camera.R = math.Max(xmax-xmin, op.camera.R)
		op.camera.R = math.Max(ymax-ymin, op.camera.R)
		op.camera.R = math.Max(zmax-zmin, op.camera.R)
		op.camera.center = Coordinate{
			X: (xmax + xmin) / 2.0,
			Y: (ymax + ymin) / 2.0,
			Z: (zmax + zmin) / 2.0,
		}
	}

	// TODO: if op.model.Coords[i].Hided {
	// TODO: 	continue
	// TODO: }

	// Point
	gl.PointSize(5)
	switch s {
	case normal, colorEdgeElements:
		gl.Begin(gl.POINTS)
		for i := range op.model.Coords {
			if op.model.Coords[i].Removed {
				continue
			}
			if op.model.Coords[i].hided {
				continue
			}
			if op.model.Coords[i].selected {
				gl.Color3ub(255, 1, 1)
			} else {
				gl.Color3ub(1, 1, 1)
			}
			gl.Vertex3d(op.model.Coords[i].X, op.model.Coords[i].Y, op.model.Coords[i].Z)
		}
		gl.End()
	case selectPoints:
		gl.Begin(gl.POINTS)
		for i := range op.model.Coords {
			if op.model.Coords[i].Removed {
				continue
			}
			if op.model.Coords[i].hided {
				continue
			}
			if op.model.Coords[i].selected {
				continue
			}
			convertToColor(i)
			gl.Vertex3d(op.model.Coords[i].X, op.model.Coords[i].Y, op.model.Coords[i].Z)
		}
		gl.End()
	case selectLines, selectTriangles: // do nothing
	default:
		panic(fmt.Errorf("not valid selection : %v", s))
	}
	// Elements
	gl.PointSize(2) // default points size
	gl.LineWidth(3) // default lines width
	for i, el := range op.model.Elements {
		if op.model.IsIgnore(uint(i)) {
			continue
		}
		// do not show selected elements in Select case
		if s != normal && s != colorEdgeElements && el.selected {
			continue
		}
		if el.hided { // hided element
			continue
		}
		if el.ElementType == ElRemove { // removed element
			continue
		}
		// color identification
		switch s {
		case normal, colorEdgeElements:
			switch el.ElementType {
			case Line2:
				if el.selected {
					gl.Color3ub(255, 1, 1)
				} else {
					gl.Color3ub(153, 153, 153)
				}
			case Triangle3:
				if el.selected {
					gl.Color3ub(255, 1, 1)
				} else {
					gl.Color3ub(153, 0, 153)
				}
			default:
				panic(fmt.Errorf("not valid element type: %v", el))
			}
		case selectPoints, selectLines, selectTriangles:
			convertToColor(i)
		default:
			panic(fmt.Errorf("not valid select element: %v", s))
		}
		// select points
		switch s {
		case selectLines, selectTriangles:
			gl.Begin(gl.POINTS)
			for _, p := range el.Indexes {
				gl.Vertex3d(op.model.Coords[p].X, op.model.Coords[p].Y, op.model.Coords[p].Z)
			}
			gl.End()
		}
		// draw lines in 3D
		switch el.ElementType {
		case Line2:
			if s == normal || s == selectLines || (s == colorEdgeElements && el.selected) {
				gl.Begin(gl.LINES)
				for _, k := range el.Indexes {
					c := op.model.Coords[k]
					gl.Vertex3d(c.X, c.Y, c.Z)
				}
				gl.End()
			}
			if s == colorEdgeElements && !el.selected {
				gl.Begin(gl.LINES)
				for i, k := range el.Indexes {
					edgeColor(i)
					c := op.model.Coords[k]
					gl.Vertex3d(c.X, c.Y, c.Z)
				}
				gl.End()
			}
		case Triangle3:
			if s == selectTriangles {
				gl.Begin(gl.LINES)
				for p := range el.Indexes {
					from, to := p, p+1
					if to == len(el.Indexes) {
						from = el.Indexes[from]
						to = el.Indexes[0]
					} else {
						from = el.Indexes[from]
						to = el.Indexes[to]
					}
					gl.Vertex3d(
						op.model.Coords[from].X,
						op.model.Coords[from].Y,
						op.model.Coords[from].Z)
					gl.Vertex3d(
						op.model.Coords[to].X,
						op.model.Coords[to].Y,
						op.model.Coords[to].Z)
				}
				gl.End()
			}
		default:
			panic(fmt.Errorf("not valid element: %v", el))
		}
		// draw triangles in 3D
		switch el.ElementType {
		case Line2: // do nothing
		case Triangle3:
			if s == normal || s == selectTriangles || (s == colorEdgeElements && el.selected) {
				gl.Begin(gl.TRIANGLES)
				for _, p := range el.Indexes {
					gl.Vertex3d(
						op.model.Coords[p].X,
						op.model.Coords[p].Y,
						op.model.Coords[p].Z)
				}
				gl.End()
			}
			if s == colorEdgeElements && !el.selected {
				gl.Begin(gl.TRIANGLES)
				for i, p := range el.Indexes {
					edgeColor(i)
					gl.Vertex3d(
						op.model.Coords[p].X,
						op.model.Coords[p].Y,
						op.model.Coords[p].Z)
				}
				gl.End()
			}
		default:
			panic(fmt.Errorf("not valid element: %v", el))
		}
	}
}

func openGlScreenCoordinate(window *glfw.Window) {
	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.TEXTURE_2D)

	w, h := window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(w), 0, float64(h), float64(-100.0), float64(100.0))

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func drawSelectRectangle(window *glfw.Window) {
	// draw select rectangle
	if !selectObjects.fromAdd || !selectObjects.toUpdate {
		return
	}
	_, h := window.GetSize()

	gl.LineWidth(1)
	gl.Begin(gl.LINES)
	gl.Color3d(1.0, 0.0, 0.0) // Red
	{
		x1 := float64(selectObjects.xFrom)
		y1 := float64(h) - float64(selectObjects.yFrom)
		x2 := float64(selectObjects.xTo)
		y2 := float64(h) - float64(selectObjects.yTo)
		gl.Vertex2d(x1, y1)
		gl.Vertex2d(x1, y2)

		gl.Vertex2d(x1, y2)
		gl.Vertex2d(x2, y2)

		gl.Vertex2d(x2, y2)
		gl.Vertex2d(x2, y1)

		gl.Vertex2d(x2, y1)
		gl.Vertex2d(x1, y1)
	}
	gl.End()
}

func (op *Opengl) drawAxes() {
	w, h := op.window.GetSize()

	s := math.Max(50.0, float64(h)/8.0)
	b := 5.0 // distance from window border

	center_x := float64(w) - b - s/2.0
	center_y := b + s/2.0
	gl.Begin(gl.QUADS)
	gl.Color3d(0.8, 0.8, 0.8)
	{
		gl.Vertex2d(center_x-s/2, center_y-s/2)
		gl.Vertex2d(center_x+s/2, center_y-s/2)
		gl.Vertex2d(center_x+s/2, center_y+s/2)
		gl.Vertex2d(center_x-s/2, center_y+s/2)
	}
	gl.End()
	gl.LineWidth(1)
	gl.Begin(gl.LINES)
	gl.Color3d(0.1, 0.1, 0.1)
	{
		gl.Vertex2d(center_x-s/2, center_y-s/2)
		gl.Vertex2d(center_x+s/2, center_y-s/2)
		gl.Vertex2d(center_x+s/2, center_y-s/2)
		gl.Vertex2d(center_x+s/2, center_y+s/2)
		gl.Vertex2d(center_x+s/2, center_y+s/2)
		gl.Vertex2d(center_x-s/2, center_y+s/2)
		gl.Vertex2d(center_x-s/2, center_y+s/2)
		gl.Vertex2d(center_x-s/2, center_y-s/2)
	}
	gl.End()

	gl.Translated(center_x, center_y, 0)
	gl.Rotated(op.camera.betta, 1.0, 0.0, 0.0)
	gl.Rotated(op.camera.alpha, 0.0, 1.0, 0.0)
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

func (op *Opengl) scroll(window *glfw.Window, xoffset, yoffset float64) {
	const factor = 0.05
	switch {
	case 0 <= yoffset:
		op.camera.R /= (1 + factor)
	case yoffset <= 0:
		op.camera.R *= (1 + factor)
	}
}

var selectObjects = struct {
	xFrom, yFrom int32
	fromAdd      bool

	xTo, yTo int32
	toUpdate bool
	toAdd    bool
}{}

type viewState = uint8

const (
	normal viewState = 1 << iota
	colorEdgeElements
	selectPoints
	selectLines
	selectTriangles
)

func edgeColor(pos int) {
	switch pos {
	case 0: // yellow
		gl.Color3ub(255, 255, 0)
		return
	case 1: // blue
		gl.Color3ub(0, 0, 255)
		return
	case 2: // green
		gl.Color3ub(0, 255, 0)
		return
	case 3: // purple
		gl.Color3ub(255, 0, 125)
		return
	}
	panic(fmt.Errorf("not valid pos: %d", pos))
}

// maximal amount colors is 245^3 = 14 706 125
const (
	convertOffset  = uint64(5)
	convertMaxUint = uint64(245) // max uint value
)

func convertToIndex(color []uint8) (index int) {
	for i := range color {
		if i == 3 {
			break
		}
		if uint64(color[i]) < convertOffset {
			return -2
		}
		if convertOffset+convertMaxUint < uint64(color[i]) {
			return -3
		}
	}
	color[0] -= uint8(convertOffset)
	color[1] -= uint8(convertOffset)
	color[2] -= uint8(convertOffset)
	u := convertMaxUint
	return int(uint64(color[0]) + u*uint64(color[1]) + u*u*uint64(color[2]))
}

func convertToColor(i int) {
	// func Color3ub(red, green, blue uint8)
	// `uint8` is the set of all unsigned 8-bit integers.
	// Range: 0 through 255.

	// offset `uint8` to 5 through 250.
	o := convertOffset
	u := convertMaxUint
	N := uint64(i)

	// Example:
	// func main() {
	// 	u := 245
	// 	N := 1111111
	// 	k0 := N % u
	// 	N = (N - k0) / u
	// 	k1 := N % u
	// 	N = (N - k1) / u
	// 	k2 := N % u
	// 	N = (N - k1) / u
	// 	k3 := N % u
	// 	fmt.Println(k0, k1, k2, k3, k0+u*k1+u*u*k2+u*u*u*k3)
	// }
	// Result:
	//	36 125 18 0 1111111
	var value [3]uint64
	value[0] = N % u
	N = (N - value[0]) / u
	value[1] = N % u
	N = (N - value[1]) / u
	value[2] = N % u
	for i := range value {
		value[i] += o
	}
	gl.Color3ub(
		uint8(value[0]),
		uint8(value[1]),
		uint8(value[2]),
	)
}

func (op *Opengl) selectByRectangle() {
	_, h := op.window.GetSize()

	selectObjects.yFrom = int32(h) - selectObjects.yFrom
	selectObjects.yTo = int32(h) - selectObjects.yTo

	//  glXGetConfig(dpy, vInfo, GLX_RED_SIZE, &attribs->redSize);
	// GLX_BUFFER_SIZE
	//
	//     Number of bits per color buffer. For RGBA visuals,
	// GLX_BUFFER_SIZE is the sum of GLX_RED_SIZE, GLX_GREEN_SIZE,
	// GLX_BLUE_SIZE, and GLX_ALPHA_SIZE.
	// For color index visuals, GLX_BUFFER_SIZE is the size of the color indexes.
	//
	// GLX_RED_SIZE
	//
	//     Number of bits of red stored in each color buffer.
	// Undefined if GLX_RGBA is False.
	//
	// glxinfo
	//     visual  x   bf lv rg d st  colorbuffer  sr ax dp st accumbuffer  ms  cav
	//   id dep cl sp  sz l  ci b ro  r  g  b  a F gb bf th cl  r  g  b  a ns b eat
	// ----------------------------------------------------------------------------
	// 0x081 24 tc  0  32  0 r  . .   8  8  8  8 .  .  0  0  0  0  0  0  0  0 0 None
	//
	// GLX_BUFFER_SIZE = 32
	// GLX_RED_SIZE    =  8
	// 8bits is 0...256

	if selectObjects.xTo < selectObjects.xFrom {
		// swap
		selectObjects.xTo, selectObjects.xFrom = selectObjects.xFrom, selectObjects.xTo
	}
	if selectObjects.yTo < selectObjects.yFrom {
		// swap
		selectObjects.yTo, selectObjects.yFrom = selectObjects.yFrom, selectObjects.yTo
	}

	var found bool

	for _, s := range []struct {
		st viewState
		sf func(index int) (found bool)
	}{
		{
			st: selectPoints, sf: func(index int) bool {
				if index < 0 || len(op.model.Coords) <= index {
					return false
				}
				op.model.Coords[index].selected = true
				return true
			},
		},
		{
			st: selectLines, sf: func(index int) bool {
				if index < 0 || len(op.model.Elements) <= index {
					return false
				}
				if op.model.Elements[index].ElementType != Line2 {
					return false
				}
				op.model.Elements[index].selected = true
				return true
			},
		},
		{
			st: selectTriangles, sf: func(index int) bool {
				if index < 0 || len(op.model.Elements) <= index {
					return false
				}
				if op.model.Elements[index].ElementType != Triangle3 {
					return false
				}
				op.model.Elements[index].selected = true
				return true
			},
		},
	} {
		if op.cursorLeft&s.st == 0 {
			continue
		}

		found = true
		for found { // TODO : infinite loop
			found = false
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.ClearColorxOES(0, 0, 0, 0) // ClearColor(1, 1, 1, 1)
			gl.Enable(gl.DEPTH_TEST)
			gl.Disable(gl.LINE_SMOOTH)
			op.cameraView()
			// color initialize
			op.model3d(s.st)

			// TODO : screen coordinates
			// TODO : openGlScreenCoordinate(window)
			// TODO : gl.Flush()

			// color selection
			var color []uint8 = make([]uint8, 4)
			for x := selectObjects.xFrom; x <= selectObjects.xTo; x++ {
				for y := selectObjects.yFrom; y <= selectObjects.yTo; y++ {
					// func ReadPixels(
					//	x int32, y int32,
					//	width int32, height int32,
					//	format uint32, xtype uint32, pixels unsafe.Pointer)
					gl.ReadPixels(x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&color[0]))
					index := convertToIndex(color)
					if s.sf(index) {
						found = true
					}
				}
			}
			// if any find selection, then try again
		}
	}
}

func (op *Opengl) mouseButton(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	switch button {
	case glfw.MouseButton1:
		switch action {
		case glfw.Press:
			x, y := w.GetCursorPos()
			selectObjects.xFrom = int32(x)
			selectObjects.yFrom = int32(y)
			selectObjects.fromAdd = true
		case glfw.Release:
			x, y := w.GetCursorPos()
			selectObjects.xTo = int32(x)
			selectObjects.yTo = int32(y)
			selectObjects.toAdd = true
		case glfw.Repeat:
			// do nothing
		}
	default:
		selectObjects.toUpdate = false
		selectObjects.fromAdd = false
		selectObjects.toAdd = false
	}
}

var (
	xlast float64
	ylast float64
)

func (op *Opengl) cursorPos(w *glfw.Window, xpos, ypos float64) {
	if selectObjects.fromAdd || selectObjects.toAdd {
		selectObjects.xTo = int32(xpos)
		selectObjects.yTo = int32(ypos)
		selectObjects.toUpdate = true
		return
	}

	const angle = 5.0
	if w.GetMouseButton(glfw.MouseButton3) == glfw.Press {
		switch {
		case xpos < xlast:
			op.camera.alpha -= angle
		case xlast < xpos:
			op.camera.alpha += angle
		}
		switch {
		case ypos < ylast:
			op.camera.betta -= angle
		case ylast < ypos:
			op.camera.betta += angle
		}
		xlast = xpos
		ylast = ypos
	}

	const factor = 0.01
	if w.GetMouseButton(glfw.MouseButton2) == glfw.Press {
		switch {
		case xpos < xlast:
			op.camera.moveX -= op.camera.R * factor
		case xlast < xpos:
			op.camera.moveX += op.camera.R * factor
		}
		switch {
		case ypos < ylast:
			op.camera.moveY += op.camera.R * factor
		case ylast < ypos:
			op.camera.moveY -= op.camera.R * factor
		}
		xlast = xpos
		ylast = ypos
	}
}

func (op *Opengl) key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		// deselect all
		op.Init()
		op.model.DeselectAll()
	}
}

func (op *Opengl) StandardView(view SView) {
	op.updateModel = true
	switch view {
	case StandardViewXOYpos:
		op.camera.alpha = 0.0
		op.camera.betta = 0.0
	case StandardViewYOZpos:
		op.camera.alpha = 90.0
		op.camera.betta = 0.0
	case StandardViewXOZpos:
		op.camera.alpha = 0.0
		op.camera.betta = 270.0
	case StandardViewXOYneg:
		op.camera.alpha = 180.0
		op.camera.betta = 0.0
	case StandardViewYOZneg:
		op.camera.alpha = 270.0
		op.camera.betta = 0.0
	case StandardViewXOZneg:
		op.camera.alpha = 0.0
		op.camera.betta = 90.0
	}
}

func (op *Opengl) ColorEdge(isColor bool) {
	if isColor {
		op.state = colorEdgeElements
	} else {
		op.state = normal
	}
}

func (op *Opengl) SelectLeftCursor(nodes, lines, tria bool) {
	op.cursorLeft = 0
	if nodes {
		op.cursorLeft |= selectPoints
	}
	if lines {
		op.cursorLeft |= selectLines
	}
	if tria {
		op.cursorLeft |= selectTriangles
	}
}

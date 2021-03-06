package ms

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/Konstantin8105/gog"
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

	mesh Mesh

	// for 3d view
	state       viewState
	cursorLeft  viewState
	updateModel bool
	camera      struct {
		alpha, betta float64
		R            float64
		center       gog.Point3d
		moveX, moveY float64
	}

	// calculate data for FPS
	fps Fps

	// mouses
	mouses   [3]Mouse  // left, middle, right
	mouseMid MouseRoll // middle scroll
}

func (op *Opengl) Init() {
	op.updateModel = true
	if op.state != normal && op.state != colorEdgeElements {
		op.state = normal
	}
	op.cursorLeft = selectPoints
}

func NewOpengl(m Mesh) (op *Opengl, err error) {
	op = new(Opengl)
	op.Init()
	op.mesh = m

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

	// mouse initialize
	op.MouseDefault()

	return
}

func (op *Opengl) MouseDefault() {
	op.mouses[0] = new(MouseSelect) // left button
	op.mouses[1] = new(MouseMove)   // right button
	op.mouses[2] = new(MouseRotate) // middle button
	op.mouseMid = new(MouseZoom)    // middle scroll
}

func (op *Opengl) Run() {
	defer func() {
		// 3D window is close
		glfw.Terminate()
	}()
	for !op.window.ShouldClose() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)

		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.BLEND) // Transparency
		gl.Enable(gl.LINE_SMOOTH)

		// Avoid panics if Model is changed.
		// Main problem of synchronization.
		func() {
			defer func() {
				if r := recover(); r != nil {
					// safety ignore panic
					<-time.After(100 * time.Millisecond)
					AddInfo("Opengl: safety ignore panic: %s", r)
				}
			}()

			op.cameraView()
			op.model3d(op.state, "run")

			// screen coordinates
			openGlScreenCoordinate(op.window)
			// draw axe coordinates
			op.drawAxes()
			// minimal screen notes
			DrawText(fmt.Sprintf("FPS       : %6.2f",
				op.fps.Get()), 0, 0*fontSize)
			if op.mesh != nil {
				DrawText(fmt.Sprintf("Nodes     : %6d",
					len(op.mesh.GetCoords())), 0, 1*fontSize)
				DrawText(fmt.Sprintf("Elements  : %6d",
					len(op.mesh.GetElements())), 0, 2*fontSize)
			}

			for i := range op.mouses {
				if op.mouses[i] == nil {
					continue
				}
				if op.mouses[i].ReadyAction() {
					op.mouses[i].Action(op)
				}
				if op.mouses[i].ReadyPreview() {
					op.mouses[i].Preview()
				}
			}
		}()

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
	ms := time.Since(f.framesTime).Milliseconds()
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

// DrawText text on the screen
func DrawText(str string, x, y int32) {
	gl.Color4ub(0, 0, 0, 255)
	gl.LoadIdentity()
	font.Printf(float32(x), float32(y), str)
}

func angle360(a float64) float64 {
	if 360.0 < a {
		return a - 360.0
	}
	if a < -360.0 {
		return a + 360.0
	}
	return a
}

// TODO: aadd to TUI and other
func (op *Opengl) ViewAll(centerCorrection bool) {
	// take only coordinates, because all element are limit be coordinates
	cos := op.mesh.GetCoords()
	// for empty coordinate no need to do anythink
	var initialized bool
	var xmin, xmax, ymin, ymax, zmin, zmax float64
	for i := range cos {
		if cos[i].hided {
			continue
		}
		if cos[i].Removed {
			continue
		}
		if !initialized {
			initialized = true
			xmin = cos[i].Point3d[0]
			xmax = cos[i].Point3d[0]
			ymin = cos[i].Point3d[1]
			ymax = cos[i].Point3d[1]
			zmin = cos[i].Point3d[2]
			zmax = cos[i].Point3d[2]
		}
		// find extemal values
		xmin = math.Min(xmin, cos[i].Point3d[0])
		ymin = math.Min(ymin, cos[i].Point3d[1])
		zmin = math.Min(zmin, cos[i].Point3d[2])
		xmax = math.Max(xmax, cos[i].Point3d[0])
		ymax = math.Max(ymax, cos[i].Point3d[1])
		zmax = math.Max(zmax, cos[i].Point3d[2])
	}
	if len(cos) == 0 || !initialized {
		op.camera.R = 1.0
		return
	}
	// update camera
	op.camera.R = math.Max(xmax-xmin, math.Max(ymax-ymin, zmax-zmin))
	if centerCorrection {
		op.camera.center = gog.Point3d{
			(xmax + xmin) / 2.0,
			(ymax + ymin) / 2.0,
			(zmax + zmin) / 2.0,
		}
		return
	}
	// without center correction
}

func (op *Opengl) UpdateModel() {
	op.updateModel = true
	// TODO  add logic
}

func (op *Opengl) cameraView() {
	// better angle value
	op.camera.alpha = angle360(op.camera.alpha)
	op.camera.betta = angle360(op.camera.betta)

	w, h := op.window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	var ratio float64
	{
		// for avoid 3D cutting back model
		const Zzoom float64 = 100.0
		// renaming
		cx := op.camera.center[0]
		cy := op.camera.center[1]
		cz := op.camera.center[2]
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

	gl.Translated(
		op.camera.center[0],
		op.camera.center[1],
		op.camera.center[2],
	)
	gl.Rotated(op.camera.betta, 1.0, 0.0, 0.0)
	gl.Rotated(op.camera.alpha, 0.0, 1.0, 0.0)
	gl.Translated(
		-op.camera.center[0],
		-op.camera.center[1],
		-op.camera.center[2],
	)

	// minimal R
	if op.camera.R < 0.1 {
		op.camera.R = 0.1
	}
}

func (op *Opengl) model3d(s viewState, parent string) {
	if op.mesh == nil {
		return
	}
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			Debug = append(Debug, fmt.Sprintf("%v\n%v", r, string(debug.Stack())))
	// 		}
	// 	}()

	//	AddInfo("model3d at begin : %v %s", s, parent)
	//	defer func() {
	//		AddInfo("model3d at end : %v %s", s, parent)
	//	}()

	gl.PushMatrix()
	defer func() {
		gl.PopMatrix()
	}()

	cos, els := op.mesh.GetCoords(), op.mesh.GetElements()

	if op.updateModel {
		op.updateModel = false

		// Do not update angles
		// angle in global plate XOZ
		// camera.alpha = 0.0
		// angle in global plate XOY
		// camera.betta = 0.0

		// distance from center to camera
		op.camera.R = 1.0
		if len(cos) == 0 {
			return
		}
		// renaming
		ps := cos
		// calculate radius
		var (
			xmin = ps[0].Point3d[0]
			xmax = ps[0].Point3d[0]
			ymin = ps[0].Point3d[1]
			ymax = ps[0].Point3d[1]
			zmin = ps[0].Point3d[2]
			zmax = ps[0].Point3d[2]
		)
		for i := range ps {
			xmin = math.Min(xmin, ps[i].Point3d[0])
			ymin = math.Min(ymin, ps[i].Point3d[1])
			zmin = math.Min(zmin, ps[i].Point3d[2])
			xmax = math.Max(xmax, ps[i].Point3d[0])
			ymax = math.Max(ymax, ps[i].Point3d[1])
			zmax = math.Max(zmax, ps[i].Point3d[2])
		}
		op.camera.R = math.Max(xmax-xmin, op.camera.R)
		op.camera.R = math.Max(ymax-ymin, op.camera.R)
		op.camera.R = math.Max(zmax-zmin, op.camera.R)
		op.camera.center = gog.Point3d{
			(xmax + xmin) / 2.0,
			(ymax + ymin) / 2.0,
			(zmax + zmin) / 2.0,
		}
	}

	// TODO: if cos[i].Hided {
	// TODO: 	continue
	// TODO: }

	// Point
	gl.PointSize(5)
	switch s {
	case normal, colorEdgeElements:
		gl.Begin(gl.POINTS)
		for i := range cos {
			if cos[i].Removed {
				continue
			}
			if cos[i].hided {
				continue
			}
			if cos[i].selected {
				gl.Color3ub(255, 1, 1)
			} else {
				gl.Color3ub(1, 1, 1)
			}
			gl.Vertex3d(cos[i].Point3d[0], cos[i].Point3d[1], cos[i].Point3d[2])
		}
		gl.End()
	case selectPoints:
		gl.Begin(gl.POINTS)
		for i := range cos {
			if cos[i].Removed {
				continue
			}
			if cos[i].hided {
				continue
			}
			if cos[i].selected {
				continue
			}
			convertToColor(i)
			gl.Vertex3d(cos[i].Point3d[0], cos[i].Point3d[1], cos[i].Point3d[2])
		}
		gl.End()
	case selectLines, selectTriangles: // do nothing
	default:
		panic(fmt.Errorf("not valid selection : %v", s))
	}
	// Elements
	gl.PointSize(2) // default points size
	gl.LineWidth(3) // default lines width
	for i, el := range els {
		if op.mesh.IsIgnore(uint(i)) {
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
					gl.Color3ub(255, 51, 51)
				} else {
					gl.Color3ub(153, 153, 153)
				}
			case Triangle3:
				if el.selected {
					// gl.Color3ub(255, 91, 91)
					gl.Color4ub(255, 91, 91, 200)
				} else {
					// gl.Color3ub(153, 0, 153)
					gl.Color4ub(153, 0, 153, 200)
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
		if (s == selectLines && el.ElementType == Line2) ||
			(s == selectTriangles && el.ElementType == Triangle3) {
			gl.Begin(gl.POINTS)
			for _, p := range el.Indexes {
				gl.Vertex3d(cos[p].Point3d[0], cos[p].Point3d[1], cos[p].Point3d[2])
			}
			gl.End()
		}
		// draw lines in 3D
		switch el.ElementType {
		case Line2:
			if s == normal || s == selectLines || (s == colorEdgeElements && el.selected) {
				gl.Begin(gl.LINES)
				for _, k := range el.Indexes {
					c := cos[k]
					gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
				}
				gl.End()
			}
			if s == colorEdgeElements && !el.selected {
				gl.Begin(gl.LINES)
				for i, k := range el.Indexes {
					edgeColor(i)
					c := cos[k]
					gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
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
						cos[from].Point3d[0],
						cos[from].Point3d[1],
						cos[from].Point3d[2])
					gl.Vertex3d(
						cos[to].Point3d[0],
						cos[to].Point3d[1],
						cos[to].Point3d[2])
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
						cos[p].Point3d[0],
						cos[p].Point3d[1],
						cos[p].Point3d[2])
				}
				gl.End()
			}
			if s == colorEdgeElements && !el.selected {
				gl.Begin(gl.TRIANGLES)
				for i, p := range el.Indexes {
					edgeColor(i)
					gl.Vertex3d(
						cos[p].Point3d[0],
						cos[p].Point3d[1],
						cos[p].Point3d[2])
				}
				gl.End()
			}
			if s == normal {
				gl.Color4ub(123, 0, 123, 200)
				gl.LineWidth(1)
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
						cos[from].Point3d[0],
						cos[from].Point3d[1],
						cos[from].Point3d[2])
					gl.Vertex3d(
						cos[to].Point3d[0],
						cos[to].Point3d[1],
						cos[to].Point3d[2])
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

func (op *Opengl) drawAxes() {
	w, h := op.window.GetSize()

	s := math.Max(50.0, float64(h)/8.0)
	b := 5.0 // distance from window border

	centerX := float64(w) - b - s/2.0
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
	op.mouseMid.Roll(int32(xoffset), int32(yoffset), op)
}

type viewState uint8

const (
	normal            viewState = 1 << iota // 1
	colorEdgeElements                       // 2
	selectPoints                            // 4
	selectLines                             // 8
	selectTriangles                         // 16
)

func (s viewState) String() string {
	switch s {
	case normal:
		return "normal"
	case colorEdgeElements:
		return "colorEdgeElements"
	case selectPoints:
		return "selectPoints"
	case selectLines:
		return "selectLines"
	case selectTriangles:
		return "selectTriangles"
	}
	return fmt.Sprintf("%d", s)
}

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

func (op *Opengl) mouseButton(
	w *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	var index int
	switch button {
	case glfw.MouseButton1:
		index = 0
	case glfw.MouseButton2:
		index = 1
	case glfw.MouseButton3:
		index = 2
	default:
		return
	}
	for i := range op.mouses {
		if op.mouses[i] == nil {
			continue
		}
		if i != index {
			op.mouses[i].Reset()
			continue
		}
		x, y := w.GetCursorPos()
		_, h := w.GetSize()
		y = float64(h) - y
		switch action {
		case glfw.Press:
			op.mouses[index].Press(int32(x), int32(y))
		case glfw.Release:
			op.mouses[index].Release(int32(x), int32(y))
		default:
			// case glfw.Repeat:
			// do nothing
		}
		return
	}
}

func (op *Opengl) cursorPos(w *glfw.Window, xpos, ypos float64) {
	_, h := w.GetSize()
	ypos = float64(h) - ypos
	for i := range op.mouses {
		if op.mouses[i] == nil {
			continue
		}
		op.mouses[i].Update(int32(xpos), int32(ypos))
	}
}

func (op *Opengl) key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		// deselect all
		op.Init()
		op.mesh.DeselectAll()
		for i := range op.mouses {
			if op.mouses[i] == nil {
				continue
			}
			op.mouses[i].Reset()
		}
		op.MouseDefault()
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

func (op *Opengl) SelectScreen(from, to [2]int32) {
	for i := range op.mouses {
		if op.mouses[i] == nil {
			continue
		}
		switch m := op.mouses[i].(type) {
		case *MouseSelect:
			m.Reset()
			m.Press(from[0], from[1])
			m.Release(to[0], to[1])
		}
	}
}

type LeftCursor uint8

const (
	AddLinesLC = iota
	AddTrianglesLC
	endLC
)

func (lc LeftCursor) AmountNodes() int {
	switch lc {
	case AddLinesLC:
		return 2
	case AddTrianglesLC:
		return 3
	}
	return -1
}

func (lc LeftCursor) String() string {
	switch lc {
	case AddLinesLC:
		return "Lines"
	case AddTrianglesLC:
		return "Triangles"
	}
	return "Undefined"
}

func (op *Opengl) AddLeftCursor(lc LeftCursor) {
	var ma MouseAdd
	ma.LC = lc
	op.mouses[0] = &ma
}

type MouseRoll interface {
	Roll(xoffset, yoffset int32, op *Opengl)
}

type MouseZoom struct {
	x, y int32
}

func (mz *MouseZoom) Roll(xoffset, yoffset int32, op *Opengl) {
	mz.x = xoffset
	mz.y = yoffset
	mz.AfterRoll(op)
}

func (mz *MouseZoom) AfterRoll(op *Opengl) {
	const factor = 0.05
	switch {
	case 0 <= mz.y:
		op.camera.R /= (1 + factor)
	case mz.y <= 0:
		op.camera.R *= (1 + factor)
	}
}

type Mouse interface {
	Press(x, y int32)
	Update(x, y int32)
	Release(x, y int32)

	ReadyPreview() bool
	Preview()

	ReadyAction() bool
	Action(op *Opengl)

	Reset()
}

type Mouse2P struct {
	from    [2]int32
	fromAdd bool

	to       [2]int32
	toUpdate bool
	toAdd    bool
}

func (m2 *Mouse2P) Press(x, y int32) {
	m2.from[0] = x
	m2.from[1] = y
	m2.fromAdd = true
}

func (m2 *Mouse2P) Update(x, y int32) {
	if !m2.fromAdd {
		return
	}
	m2.to[0] = x
	m2.to[1] = y
	m2.toUpdate = true
}

func (m2 *Mouse2P) Release(x, y int32) {
	if !m2.fromAdd {
		return
	}
	m2.to[0] = x
	m2.to[1] = y
	m2.toAdd = true
}

func (m2 *Mouse2P) Reset() {
	m2.fromAdd = false
	m2.toUpdate = false
	m2.toAdd = false
}

func (m2 *Mouse2P) ReadyPreview() bool {
	if !m2.fromAdd || !m2.toUpdate {
		return false
	}
	return true
}

func (m2 *Mouse2P) ReadyAction() bool {
	if !m2.fromAdd || !m2.toAdd {
		return false
	}
	return true
}

type MouseSelect struct {
	Mouse2P
}

func (ms *MouseSelect) Preview() {
	if !ms.ReadyPreview() {
		return
	}

	// draw select rectangle
	gl.LineWidth(1)
	gl.Begin(gl.LINES)
	gl.Color3d(1.0, 0.0, 0.0) // Red
	{
		var (
			x1 = ms.from[0]
			y1 = ms.from[1]
			x2 = ms.to[0]
			y2 = ms.to[1]
		)
		gl.Vertex2i(x1, y1)
		gl.Vertex2i(x1, y2)

		gl.Vertex2i(x1, y2)
		gl.Vertex2i(x2, y2)

		gl.Vertex2i(x2, y2)
		gl.Vertex2i(x2, y1)

		gl.Vertex2i(x2, y1)
		gl.Vertex2i(x1, y1)
	}
	gl.End()
}

func (ms *MouseSelect) Action(op *Opengl) {
	if !ms.ReadyAction() {
		return
	}
	defer ms.Reset()

	for c := 0; c < 2; c++ {
		if ms.to[c] < ms.from[c] {
			// swap
			ms.to[c], ms.from[c] = ms.from[c], ms.to[c]
		}
		if ms.from[c] < 0 {
			ms.from[c] = 0
		}
		if ms.to[c] < 0 {
			ms.to[c] = 0
		}
	}

	cos, els := op.mesh.GetCoords(), op.mesh.GetElements()

	for _, s := range []struct {
		st viewState
		sf func(index int) (found bool)
	}{
		{st: selectPoints, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(cos) <= index {
				AddInfo("selectPoints index outside: %d", index)
				return false
			}
			cos[index].selected = true
			return true
		}}, {st: selectLines, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(els) <= index {
				AddInfo("selectLines index outside: %d", index)
				return false
			}
			if els[index].ElementType != Line2 {
				AddInfo("selectLines index is not line: %d. %v",
					index, els[index])
				return false
			}
			els[index].selected = true
			return true
		}}, {st: selectTriangles, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(els) <= index {
				AddInfo("selectTriangles index outside: %d", index)
				return false
			}
			if els[index].ElementType != Triangle3 {
				AddInfo("selectTriangles index is not triangle: %d", index)
				return false
			}
			els[index].selected = true
			return true
		}},
	} {
		if op.cursorLeft&s.st == 0 {
			continue
		}

		// find selection
		found := true
		for found { // TODO : infinite loop
			found = false
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.ClearColorxOES(0, 0, 0, 0)
			// gl.ClearColor(1, 1, 1, 1)
			gl.Enable(gl.DEPTH_TEST)
			gl.Disable(gl.LINE_SMOOTH)
			gl.Disable(gl.BLEND) // Transparency
			op.cameraView()
			// color initialize

			op.model3d(s.st, "select")

			// TODO : screen coordinates
			// TODO : openGlScreenCoordinate(window)
			// TODO :

			gl.Flush()
			gl.Finish()

			// color selection
			sizeX := (ms.to[0] - ms.from[0] + 1)
			sizeY := (ms.to[1] - ms.from[1] + 1)
			size := sizeX * sizeY
			color := make([]uint8, 4*size)
			gl.ReadPixels(ms.from[0], ms.from[1], sizeX, sizeY,
				gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&color[0]))
			for i := 0; i < len(color)-1; i += 4 {
				index := convertToIndex(color[i : i+4])
				if s.sf(index) {
					found = true
				}
			}

			// if any find selection, then try again
		}
	}
}

type MouseRotate struct {
	Mouse2P
}

func (mr *MouseRotate) Preview() {}

func (mr *MouseRotate) Action(op *Opengl) {
	if !mr.ReadyAction() {
		return
	}
	defer mr.Reset()
	// action
	const angle = 15.0

	dx := mr.to[0] - mr.from[0]
	dy := mr.to[1] - mr.from[1]
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dy < dx {
		switch {
		case mr.to[0] < mr.from[0]:
			op.camera.alpha -= angle
		case mr.from[0] < mr.to[0]:
			op.camera.alpha += angle
		}
	} else {
		switch {
		case mr.to[1] < mr.from[1]:
			op.camera.betta += angle
		case mr.from[1] < mr.to[1]:
			op.camera.betta -= angle
		}
	}
}

type MouseMove struct {
	Mouse2P
}

func (mr *MouseMove) Preview() {}

func (mr *MouseMove) Action(op *Opengl) {
	if !mr.ReadyAction() {
		return
	}
	defer mr.Reset()
	// action
	const factor = 0.05
	switch {
	case mr.to[0] < mr.from[0]:
		op.camera.moveX -= op.camera.R * factor
	case mr.from[0] < mr.to[0]:
		op.camera.moveX += op.camera.R * factor
	}
	switch {
	case mr.to[1] < mr.from[1]:
		op.camera.moveY -= op.camera.R * factor
	case mr.from[1] < mr.to[1]:
		op.camera.moveY += op.camera.R * factor
	}
}

type MouseAdd struct {
	MouseSelect

	LC LeftCursor
	ps []uint
}

func (ma *MouseAdd) Action(op *Opengl) {
	if !ma.ReadyAction() {
		return
	}
	defer ma.MouseSelect.Reset()

	// store node
	op.mesh.DeselectAll()
	ma.MouseSelect.Action(op)
	ids := op.mesh.SelectNodes(Single)
	if len(ids) != 1 {
		return
	}
	// nodes is not same
	var same bool
	for i := range ma.ps {
		if ma.ps[i] == ids[0] {
			same = true
		}
	}
	if !same {
		ma.ps = append(ma.ps, ids[0])
	}
	// TODO : add preview indicated points
	op.mesh.DeselectAll()

	if len(ma.ps) != ma.LC.AmountNodes() {
		return
	}
	// action
	switch ma.LC {
	case AddLinesLC:
		op.mesh.AddLineByNodeNumber(
			ma.ps[0],
			ma.ps[1],
		)
	case AddTrianglesLC:
		op.mesh.AddTriangle3ByNodeNumber(
			ma.ps[0],
			ma.ps[1],
			ma.ps[2],
		)
	}
	ma.Reset()
}

func (ma *MouseAdd) Reset() {
	ma.ps = nil
	ma.MouseSelect.Reset()
}

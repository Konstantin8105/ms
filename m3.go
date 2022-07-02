package ms

import (
	"fmt"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gltext"
)

func init() {
	runtime.LockOSThread()
}

var screen vl.Screen

func init() {
	root, _, err := UserInterface()
	if err != nil {
		panic(err)
	}
	var l vl.ListH
	l.Add(root)
	l.Add(nil)
	screen.Root = &l
}

var (
	font     *Font
	fontSize int = 12
)

var fps Fps

func M3() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(800, 600, "3D model", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// window.SetMouseButtonCallback(mouseButtonCallback)
	window.SetScrollCallback(scrollCallback)
	window.SetCursorPosCallback(cursorPosCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	glfw.SwapInterval(1) // Enable vsync

	// ???

	file := "/home/konstantin/.fonts/Go-Mono-Bold.ttf"
	font, err = NewFont(file, fontSize)
	if err != nil {
		panic(err)
	}

	fps.Init()

	for !window.ShouldClose() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)

		gl.Disable(gl.DEPTH_TEST)
		gl.Disable(gl.LIGHTING)

		// ui(window)

		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.TEXTURE_2D)

		cameraView(window)
		model3d(window)
		axe(window)

		font.Draw(fmt.Sprintf("FPS  : %6.2f", fps.Get()), 0, 0*fontSize)
		font.Draw(fmt.Sprintf("Nodes: %6d", len(model.Points)), 0, 1*fontSize)
		font.Draw(fmt.Sprintf("Lines: %6d", len(model.Lines)), 0, 2*fontSize)

		window.MakeContextCurrent()
		window.SwapBuffers()

		fps.EndFrame()
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

// Font handle
type Font struct {
	Handle *gltext.Font
}

// NewFont create new Font from given filename (.ttf expected)
func NewFont(filename string, size int) (*Font, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	font, _ := gltext.LoadTruetype(fd, int32(size), 32, 127, gltext.LeftToRight)
	return &Font{Handle: font}, nil
}

// Draw text on the screen
func (f *Font) Draw(str string, x, y int) { // , c Color) {
	gl.Color4ub(0, 0, 0, 255)
	gl.LoadIdentity()
	f.Handle.Printf(float32(x), float32(y), str)
}

func (f *Font) Metrics(text string) (int, int) {
	return f.Handle.Metrics(text)
}

var cells [][]vl.Cell

func ui(window *glfw.Window) {
	// OpenGl implementation of vl.Drawer
	// Drawer = func(row, col uint, s tcell.Style, r rune)
	w, h := window.GetSize()

	gw, gh := font.Metrics(" ")

	runeW := uint(w / gw)
	runeH := uint(h / gh)

	// panic (fmt.Errorf("%v %v", runeW, runeH))

	screen.SetHeight(runeH)
	screen.GetContents(runeW, &cells)

	for r := range cells {
		for c := range cells[r] {
			cell := cells[r][c]

			gw, gh := font.Metrics(string(cell.R))

			// background
			fg, bg, _ := cell.S.Decompose()
			_ = fg // TODO

			switch bg {
			case tcell.ColorYellow:
				gl.Color3d(1, 1, 0)
			case tcell.ColorViolet:
				gl.Color3d(0.5, 0, 1)
			case tcell.ColorWhite:
				gl.Color3d(1, 1, 1)
			case tcell.ColorBlack:
				gl.Color3d(0, 0, 0)
			default:
				gl.Color3d(1, 0, 0)
			}
			gl.Begin(gl.QUADS)
			{
				// func Vertex2i(x int32, y int32)
				gl.Vertex2i(int32((c+0)*gw), int32(h-(r+0)*gh))
				gl.Vertex2i(int32((c+1)*gw), int32(h-(r+0)*gh))
				gl.Vertex2i(int32((c+1)*gw), int32(h-(r+1)*gh))
				gl.Vertex2i(int32((c+0)*gw), int32(h-(r+1)*gh))
			}
			gl.End()

			// text
			if cells[r][c].R == ' ' {
				continue
			}
			font.Draw(string(cell.R), gw*int(c), gh*int(r))
		}
	}
}

var camera = struct {
	alpha, betta float64
	R            float64
	center       Point
}{
	alpha:  0,
	betta:  0,
	R:      1,
	center: Point{0, 0, 0},
}

var (
	orientation = [3]float64{0, 0, 0} // Radians
	position    = [3]float64{0, 0, 0}
	scale       = [3]float64{1, 1, 1}
)

func angle_norm(a float64) float64 {
	if 360.0 < a {
		return a - 360.0
	}
	if a < -360.0 {
		return a + 360.0
	}
	return a
}

func cameraView(window *glfw.Window) {
	// better angle value
	camera.alpha = angle_norm(camera.alpha)
	camera.betta = angle_norm(camera.betta)

	w, h := window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	ratio := 1.0
	{
		// for avoid 3D cutting back model
		var Z_ratio float64 = 100.0
		// renaming
		cx := camera.center.X
		cy := camera.center.Y
		cz := camera.center.Z
		// scaling monitor 3d model on screen
		if w < h {
			ratio = float64(w) / float64(h)
			gl.Ortho(
				(-camera.R)+cx, (camera.R)+cx,
				(-camera.R)/ratio+cy, (camera.R)/ratio+cy,
				(-camera.R-cz)*Z_ratio, (camera.R+cz)*Z_ratio)
		} else {
			ratio = float64(h) / float64(w)
			gl.Ortho(
				(-camera.R)/ratio+cx, (camera.R)/ratio+cx,
				(-camera.R)+cy, (camera.R)+cy,
				(-camera.R-cz)*Z_ratio, (camera.R+cz)*Z_ratio)
		}
	}
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Translated(camera.center.X, camera.center.Y, camera.center.Z)
	gl.Rotated(camera.betta, 1.0, 0.0, 0.0)
	gl.Rotated(camera.alpha, 0.0, 1.0, 0.0)
	gl.Translated(-camera.center.X, -camera.center.Y, -camera.center.Z)

	// minimal R
	if camera.R < 0.1 {
		camera.R = 0.1
	}
}

var updateModel bool // TODO remove

func model3d(window *glfw.Window) {
	gl.PushMatrix()
	gl.Rotated(orientation[0], 1, 0, 0)
	gl.Rotated(orientation[1], 0, 1, 0)
	gl.Rotated(orientation[2], 0, 0, 1)
	gl.Translated(position[0], position[1], position[2])
	gl.Scaled(scale[0], scale[1], scale[2])
	defer func() {
		gl.PopMatrix()
	}()

	if updateModel {
		updateModel = false
		// default values
		// angle in global plate XOZ
		camera.alpha = 0.0
		// angle in global plate XOY
		camera.betta = 0.0
		// distance from center to camera
		camera.R = 1.0
		if len(model.Points) == 0 {
			return
		}
		// renaming
		ps := model.Points
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
		camera.R = math.Max(xmax-xmin, camera.R)
		camera.R = math.Max(ymax-ymin, camera.R)
		camera.R = math.Max(zmax-zmin, camera.R)
		camera.center = Point{
			(xmax + xmin) / 2.0,
			(ymax + ymin) / 2.0,
			(zmax + zmin) / 2.0,
		}
	}

	// Point
	gl.PointSize(5)
	gl.Begin(gl.POINTS)
	gl.Color3d(0, 0, 0)
	for i := range model.Points {
		gl.Vertex3d(model.Points[i].X, model.Points[i].Y, model.Points[i].Z)
	}
	gl.End()
	// Lines
	gl.LineWidth(2)
	gl.LineStipple(1, 0x00ff)
	gl.Begin(gl.LINES)
	gl.Color3d(0.6, 0.6, 0.6)
	for i := range model.Lines {
		f := model.Points[model.Lines[i][0]]
		t := model.Points[model.Lines[i][1]]
		gl.Vertex3d(f.X, f.Y, f.Z)
		gl.Vertex3d(t.X, t.Y, t.Z)
	}
	gl.End()
	// Triangle
	gl.Begin(gl.TRIANGLES)
	gl.Color3d(0.6, 0.0, 0.6)
	for i := range model.Triangles {
		for p := 0; p < 3; p++ {
			gl.Vertex3d(
				model.Points[model.Triangles[i][p]].X,
				model.Points[model.Triangles[i][p]].Y,
				model.Points[model.Triangles[i][p]].Z)
		}
	}
	gl.End()

	// Point text
	// w, h := window.GetSize()
	// for i := range ps {
	// 	x := float64(w)/2 + (ps[i].Z-camera.center.Z)*camera.R
	// 	y := float64(h)/2 + (ps[i].Y-camera.center.Y)*camera.R
	// 	font.Draw(fmt.Sprintf("%d", i), int(x), int(y))
	// }
}

func axe(window *glfw.Window) {
	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.TEXTURE_2D)

	gl.LineWidth(1)
	gl.LineStipple(1, 0x00ff)

	w, h := window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(w), 0, float64(h), float64(-100.0), float64(100.0))

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	s := math.Max(50.0, float64(h)/8.0)
	b := 5.0 // distance from window border

	center_x := float64(w) - b - s/2.0
	center_y := b + s/2.0
	gl.Begin(gl.QUADS)
	gl.Color3d(0.8, 0.8, 0.8)
	{
		gl.Vertex3d(center_x-s/2, center_y-s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y-s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y+s/2, 0)
		gl.Vertex3d(center_x-s/2, center_y+s/2, 0)
	}
	gl.End()
	gl.Begin(gl.LINES)
	gl.Color3d(0.1, 0.1, 0.1)
	{
		gl.Vertex3d(center_x-s/2, center_y-s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y-s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y-s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y+s/2, 0)
		gl.Vertex3d(center_x+s/2, center_y+s/2, 0)
		gl.Vertex3d(center_x-s/2, center_y+s/2, 0)
		gl.Vertex3d(center_x-s/2, center_y+s/2, 0)
		gl.Vertex3d(center_x-s/2, center_y-s/2, 0)
	}
	gl.End()

	gl.Translated(center_x, center_y, 0)
	gl.Rotated(camera.betta, 1.0, 0.0, 0.0)
	gl.Rotated(camera.alpha, 0.0, 1.0, 0.0)
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

func scrollCallback(window *glfw.Window, xoffset, yoffset float64) {
	const factor = 0.05
	switch {
	case 0 <= yoffset:
		camera.R /= (1 + factor)
	case yoffset <= 0:
		camera.R *= (1 + factor)
	}
}

// func mouseButtonCallback(
// 	w *glfw.Window,
// 	button glfw.MouseButton,
// 	action glfw.Action,
// 	mods glfw.ModifierKey,
// ) {
// 	if button == glfw.MouseButton1 && action == glfw.Press {
// 		x, y := w.GetMousePos()
// 		xlast = x
// 		_ = y
// 		// camera.alpha += 5
// 	}
// }

var (
	xlast float64
	ylast float64
)

func cursorPosCallback(w *glfw.Window, xpos, ypos float64) {
	const angle = 5.0
	if w.GetMouseButton(glfw.MouseButton1) == glfw.Press {
		switch {
		case xpos < xlast:
			camera.alpha -= angle
		case xlast < xpos:
			camera.alpha += angle
		}
		switch {
		case ypos < ylast:
			camera.betta -= angle
		case ylast < ypos:
			camera.betta += angle
		}
		xlast = xpos
		ylast = ypos
	}
}

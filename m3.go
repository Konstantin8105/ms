package ms

import (
	_ "image/png"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

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

	if err := gl.Init(); err != nil {
		panic(err)
	}

	glfw.SwapInterval(1) // Enable vsync

	// ???

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.TEXTURE_2D)

	for !window.ShouldClose() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1, 1, 1, 1)
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.TEXTURE_2D)

		cameraView(window)
		model(window)
		axe(window)

		{ // TODO remove
			camera.alpha += 0.2
			camera.betta += 0.05
		}

		window.MakeContextCurrent()
		window.SwapBuffers()
	}
}

type Point struct {
	X, Y, Z float64
}

var (
	ps          []Point
	ls          [][2]int
	ts          [][3]int
	updateModel bool
)

func init() { // TODO remove
	var (
		Ri     = 0.5
		Ro     = 2.5
		da     = 30.0 // degree
		dy     = 0.2
		amount = 80
		len_ps = amount * 2
		len_ls = amount + 2*(amount-1)
		len_ts = 2 * (amount - 1)
	)
	ps = make([]Point, len_ps)
	ls = make([][2]int, len_ls)
	ts = make([][3]int, len_ts)
	for i := 0; i < amount; i++ {
		ps[2*i+0].X = Ri * math.Sin(float64(i)*da*math.Pi/180.0)
		ps[2*i+0].Z = Ri * math.Cos(float64(i)*da*math.Pi/180.0)
		ps[2*i+0].Y = float64(i) * dy
		ps[2*i+1].X = Ro * math.Sin(float64(i)*da*math.Pi/180.0)
		ps[2*i+1].Z = Ro * math.Cos(float64(i)*da*math.Pi/180.0)
		ps[2*i+1].Y = float64(i) * dy
		ls[i][0] = 2*i + 0
		ls[i][1] = 2*i + 1
		if i != 0 {
			ls[1*(amount-1)+i][0] = 2*(i-1) + 0
			ls[1*(amount-1)+i][1] = 2*(i-0) + 0
			ls[2*(amount-1)+i][0] = 2*(i-1) + 1
			ls[2*(amount-1)+i][1] = 2*(i-0) + 1
		}
		if i != 0 {
			ts[i-1][0] = 2*(i-1) + 0
			ts[i-1][1] = 2*(i-1) + 1
			ts[i-1][2] = 2*(i-0) + 0
			ts[amount-1+i-1][0] = 2*(i-1) + 1
			ts[amount-1+i-1][1] = 2*(i-0) + 0
			ts[amount-1+i-1][2] = 2*(i-0) + 1
		}
	}
	updateModel = true
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

// maximal radius of object
var (
	X       float64 = 0.0
	Y       float64 = 0.0
	Z_ratio float64 = 100.0 // for avoid 3D cutting back model
)

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
		cx := camera.center.X
		cy := camera.center.Y
		cz := camera.center.Z
		// scaling monitor 3d model on screen
		if w < h {
			ratio = float64(w) / float64(h)
			gl.Ortho(
				(-camera.R-X)+cx, (camera.R-X)+cx,
				(-camera.R-Y)/ratio+cy, (camera.R-Y)/ratio+cy,
				(-camera.R-cz)*Z_ratio, (camera.R+cz)*Z_ratio)
		} else {
			ratio = float64(h) / float64(w)
			gl.Ortho(
				(-camera.R-X)/ratio+cx, (camera.R-X)/ratio+cx,
				(-camera.R-Y)+cy, (camera.R-Y)+cy,
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

func model(window *glfw.Window) {
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
		if len(ps) == 0 {
			return
		}
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
	for i := range ps {
		gl.Vertex3d(ps[i].X, ps[i].Y, ps[i].Z)
	}
	gl.End()
	// Lines
	gl.LineWidth(2)
	gl.LineStipple(1, 0x00ff)
	gl.Begin(gl.LINES)
	gl.Color3d(0.6, 0.6, 0.6)
	for i := range ls {
		f := ps[ls[i][0]]
		t := ps[ls[i][1]]
		gl.Vertex3d(f.X, f.Y, f.Z)
		gl.Vertex3d(t.X, t.Y, t.Z)
	}
	gl.End()
	// Triangle
	gl.Begin(gl.TRIANGLES)
	gl.Color3d(0.6, 0.0, 0.6)
	for i := range ts {
		for p := 0; p < 3; p++ {
			gl.Vertex3d(
				ps[ts[i][p]].X,
				ps[ts[i][p]].Y,
				ps[ts[i][p]].Z)
		}
	}
	gl.End()
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

	var s float64 = 50
	if sh := float64(h / 8.0); s < sh {
		s = sh
	}
	var b float64 = 5 // distance from window border

	var center_x float64 = float64(w) - b - s/2.0
	var center_y float64 = b + s/2.0
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

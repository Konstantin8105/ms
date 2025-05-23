package ms

import (
	"fmt"
	"math"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/pow"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var _ ds.Window = (*Opengl)(nil)

var WindowRatio float64 = 0.4

type Opengl struct {
	font *glsymbol.Font

	actions *chan ds.Action

	x, y, w, h int32

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
	cube struct {
		min, max gog.Point3d
	}

	// mouses
	mouses   [3]Mouse  // left, middle, right
	mouseMid MouseRoll // middle scroll
}

func (op *Opengl) SetFont(f *glsymbol.Font) {
	op.font = f
}
func (op *Opengl) SetMouseButtonCallback(
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
	xcursor, ycursor float64,
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
		x, y := xcursor, ycursor // w.GetCursorPos()
		h := op.h                //_, h := w.GetSize()
		y = float64(h) - y
		switch action {
		case glfw.Press:
			op.mouses[index].WithShiftKey(mods == glfw.ModShift)
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
func (op *Opengl) SetCharCallback(r rune) {
	// TODO
}
func (op *Opengl) SetScrollCallback(
	xcursor, ycursor float64,
	xoffset, yoffset float64,
) {
	op.mouseMid.Roll(int32(xoffset), int32(yoffset), op)
}
func (op *Opengl) SetKeyCallback(
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	switch key {
	case glfw.KeyEscape:
		// deselect all
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
func (op *Opengl) SetCursorPosCallback(
	xpos float64,
	ypos float64,
) {
	h := op.h
	ypos = float64(h) - ypos
	for i := range op.mouses {
		if op.mouses[i] == nil {
			continue
		}
		op.mouses[i].Update(int32(xpos), int32(ypos))
	}
}
func (op *Opengl) Draw(x, y, w, h int32) {
	op.x, op.y, op.w, op.h = x, y, w, h
	gl.Viewport(int32(x), int32(y), int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	gl.Ortho(0, float64(w), 0, float64(h), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Disable(gl.LIGHTING)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	defer func() {
		gl.DepthFunc(gl.LESS)
		gl.Disable(gl.DEPTH_TEST)
	}()
	// gl.Enable(gl.DEPTH_TEST)
	// gl.Enable(gl.BLEND) // Transparency
	// gl.Enable(gl.LINE_SMOOTH)

	// TODO transperency on back side

	// switch to wireframe mode
	// gl.PolygonMode( gl.FRONT_AND_BACK, gl.LINE );
	// switch off wireframe
	// gl.PolygonMode( gl.FRONT_AND_BACK, gl.FILL );
	// gl.PolygonMode( gl.FRONT_AND_BACK, gl.POINT );

	// Avoid panics if Model is changed.
	// Main problem of synchronization.

	op.cameraView(x, y, w, h)
	op.model3d(op.state, randomPoint)

	// draw axe coordinates
	openGlScreenCoordinate(x, y, w, h)
	op.drawAxes(w, h)

	// minimal screen notes
	openGlScreenCoordinate(x, y, w, h)
	if op.mesh != nil {
		gl.Color3ub(0, 0, 0) // black
		op.font.Printf(10, 25, fmt.Sprintf("Nodes     : %6d",
			len(op.mesh.GetCoords())))
		op.font.Printf(10, 40, fmt.Sprintf("Elements  : %6d",
			len(op.mesh.GetElements())))
	}

	for i := range op.mouses {
		if op.mouses[i] == nil {
			continue
		}
		if op.mouses[i].ReadyAction() {
			op.mouses[i].Action(op)
		}
		if op.mouses[i].ReadyPreview() {
			op.mouses[i].Preview(op.x, op.y)
		}
		// view status of mouses
		x := 10 + float32(w-10)/float32(len(op.mouses))*float32(i)
		y := float32(h) - 20
		gl.Color3ub(0, 0, 0) // black
		op.font.Printf(x, y, op.mouses[i].Status())
	}
	{
		// view status
		op.font.Printf(10, float32(h)-35, op.state.String())

		name := "Select:"
		if op.cursorLeft&selectPoints != 0 {
			name += fmt.Sprintf(" %s", selectPoints)
		}
		if op.cursorLeft&selectLines != 0 {
			name += fmt.Sprintf(" %s", selectLines)
		}
		if op.cursorLeft&selectTriangles != 0 {
			name += fmt.Sprintf(" %s", selectTriangles)
		}
		if op.cursorLeft&selectQuadrs != 0 {
			name += fmt.Sprintf(" %s", selectQuadrs)
		}
		gl.Color3ub(0, 0, 0) // black
		op.font.Printf(10, float32(h)-50, name)
	}

	// TODO : REMOVE: gl.Disable(gl.DEPTH_TEST)
	// TODO : REMOVE: ui(window)

	// 			op.window.MakeContextCurrent()
	// 			op.window.SwapBuffers()

	// TODO
}

func (op *Opengl) Init() {
	op.updateModel = true
	if op.state != normal && op.state != colorEdgeElements {
		op.state = normal
	}
	op.cursorLeft = selectPoints
	*op.actions <- func() bool { return true }
}

func NewOpengl(m Mesh, actions *chan ds.Action) (op *Opengl, err error) {
	op = new(Opengl)
	op.mesh = m
	op.actions = actions

	// 	if err = glfw.Init(); err != nil {
	// 		err = fmt.Errorf("failed to initialize glfw: %v", err)
	// 		return
	// 	}
	//
	// 	glfw.WindowHint(glfw.Resizable, glfw.True)
	// 	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	// 	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	//
	// 	op.window, err = glfw.CreateWindow(800, 600, "3D model", nil, nil)
	// 	if err != nil {
	// 		return
	// 	}
	// 	op.window.MakeContextCurrent()
	//
	// 	op.window.SetMouseButtonCallback(op.mouseButton)
	// 	op.window.SetScrollCallback(op.scroll)
	// 	op.window.SetCursorPosCallback(op.cursorPos)
	// 	op.window.SetKeyCallback(op.key)
	//
	// 	if err = gl.Init(); err != nil {
	// 		return
	// 	}
	//
	// 	glfw.SwapInterval(1) // Enable vsync
	//
	// 	// ???
	//
	// 	// create new Font
	// 	op.font, err = glsymbol.DefaultFont()
	// 	if err != nil {
	// 		return
	// 	}
	// 	// font is prepared
	//

	// TODO : gl.Disable(gl.LIGHTING)

	// mouse initialize
	op.MouseDefault()

	// initialize
	op.Init()

	return
}

func (op *Opengl) MouseDefault() {
	op.mouses[0] = new(MouseSelect) // left button
	op.mouses[1] = new(MouseMove)   // right button
	op.mouses[2] = new(MouseRotate) // middle button
	op.mouseMid = new(MouseZoom)    // middle scroll
}

// TODO: aadd to TUI and other
func (op *Opengl) ViewAll() {
	op.updateModel = true
}

func (op *Opengl) UpdateModel() {
	op.updateModel = true
	// TODO  add logic
}

func (op *Opengl) cameraView(x, y, w, h int32) {
	// better angle value
	op.camera.alpha = angle360(op.camera.alpha)
	op.camera.betta = angle360(op.camera.betta)

	// w, h := op.window.GetSize()
	gl.Viewport(int32(x), int32(y), int32(w), int32(h))
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

// https://blog.jayway.com/2009/12/04/opengl-es-tutorial-for-android-part-ii-building-a-polygon/
// http://web.archive.org/web/20120527185124/http://cgg-journal.com/2008-2/06/index.html
func (op *Opengl) model3d(s viewState, fill selectState) {
	if op.mesh == nil {
		return
	}

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

		op.camera.center = gog.Point3d{}
		op.camera.R = 1.0
		op.camera.moveX = 0
		op.camera.moveY = 0
		// take only coordinates, because all element are limit be coordinates
		cos := op.mesh.GetCoords()
		if len(cos) == 0 {
			return
		}
		// for empty coordinate no need to do anythink
		var ps []gog.Point3d
		for i := range cos {
			if cos[i].hided || cos[i].Removed {
				continue
			}
			ps = append(ps, cos[i].Point3d)
		}
		// update camera
		min, max := gog.BorderPoints3d(ps...)
		// minimal camera radius is diameter of cylinder with object inside
		dia := math.Sqrt(pow.E2(max[0]-min[0]) + pow.E2(max[2]-min[2]))
		op.camera.R = dia
		// size after rotation
		dz := math.Abs(max[1] - min[1])
		op.camera.R = math.Max(
			op.camera.R,
			math.Abs(dz*math.Cos(op.camera.betta*math.Pi/180)),
		)
		op.camera.center = gog.Point3d{
			(max[0] + min[0]) / 2.0,
			(max[1] + min[1]) / 2.0,
			(max[2] + min[2]) / 2.0,
		}
		op.camera.R *= 0.5 + 0.05
		op.cube.min = min
		op.cube.max = max
	}

	// for debugging only:
	// {
	// 	min := op.cube.min
	// 	max := op.cube.max
	// 	gl.Color3ub(240, 240, 240)
	// 	gl.PointSize(1)
	// 	gl.Begin(gl.POINTS)
	// 	gl.Vertex3d(min[0], min[1], min[2])
	// 	gl.Vertex3d(min[0], max[1], min[2])
	// 	gl.Vertex3d(max[0], max[1], min[2])
	// 	gl.Vertex3d(max[0], min[1], min[2])
	// 	gl.Vertex3d(min[0], min[1], max[2])
	// 	gl.Vertex3d(min[0], max[1], max[2])
	// 	gl.Vertex3d(max[0], max[1], max[2])
	// 	gl.Vertex3d(max[0], min[1], max[2])
	// 	gl.End()
	// }

	if !(s == selectTriangles ||
		s == selectQuadrs ||
		s == selectLines ||
		s == selectPoints) {
		// TODO CREATE A GREAT LINES
		gl.Disable(gl.POLYGON_OFFSET_FILL)
	}

	op.drawElements(s, fill)
	op.drawPoints(s, fill)
}

func (op *Opengl) drawPoints(s viewState, fill selectState) {
	cos := op.mesh.GetCoords()

	gl.Disable(gl.DEPTH_TEST)
	defer func() {
		gl.Enable(gl.DEPTH_TEST)
	}()
	// prepare colors
	var r, g, b uint8
	const pointSize = 4
	// Point
	switch s {
	case normal, colorEdgeElements:
		gl.PointSize(pointSize)
		gl.Begin(gl.POINTS)
		for i := range cos {
			if cos[i].Removed {
				continue
			}
			if cos[i].hided {
				continue
			}
			if cos[i].selected {
				r, g, b = 255, 1, 1
			} else {
				r, g, b = 1, 1, 1
			}
			gl.Color3ub(r, g, b)
			gl.Vertex3d(cos[i].Point3d[0], cos[i].Point3d[1], cos[i].Point3d[2])
		}
		gl.End()
	case selectPoints:
		gl.PointSize(pointSize)
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
			r, g, b = convertToColor(i)
			gl.Color3ub(r, g, b)
			gl.Vertex3d(cos[i].Point3d[0], cos[i].Point3d[1], cos[i].Point3d[2])
		}
		gl.End()
	case selectLines, selectTriangles, selectQuadrs:
		// do nothing
	default:
		logger.Printf("not valid selection : %v", s)
	}
}

func (op *Opengl) drawElements(s viewState, fill selectState) {
	cos := op.mesh.GetCoords()
	els := op.mesh.GetElements()

	switch s {
	case normal, colorEdgeElements:
		gl.ShadeModel(gl.SMOOTH) // for points color
		gl.Enable(gl.POLYGON_OFFSET_FILL)
		gl.PolygonOffset(1.0, 1.0)
	case selectPoints, selectLines, selectTriangles, selectQuadrs:
		gl.ShadeModel(gl.FLAT)
		gl.Disable(gl.LINE_SMOOTH)
		gl.Disable(gl.POLYGON_OFFSET_FILL)
	}

	// random point for fast selection
	randomPoint := func(iel int) {
		el := els[iel]
		// ramdom point on border of element
		ratio := 0.1 + 0.8*float64(iel)/float64(len(els)) // 0.1 ... 0.9
		if ratio <= 0.0 || 1.0 <= ratio {
			logger.Printf("%e %v %v", ratio, iel, len(els))
			ratio = 0.5
		}
		size := len(el.Indexes) // 2 for Line2, 3 for Triangle3, ...
		b := cos[el.Indexes[size-2]].Point3d
		f := cos[el.Indexes[size-1]].Point3d

		// draw
		gl.PointSize(1)
		gl.Begin(gl.POINTS)
		gl.Vertex3d(
			b[0]+ratio*(f[0]-b[0]),
			b[1]+ratio*(f[1]-b[1]),
			b[2]+ratio*(f[2]-b[2]),
		)
		gl.End()
	}

	// prepare colors
	var r, g, b uint8
	// Elements
	for iel, el := range els {
		// 		if op.mesh.IsIgnore(uint(iel)) {
		// 			continue
		// 		}
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
		switch el.ElementType {
		///////////////////////////////////
		case Line2:
			switch s {
			case normal:
				gl.LineWidth(3)
				gl.Enable(gl.LINE_SMOOTH)
				gl.Begin(gl.LINES)
				for _, k := range el.Indexes {
					c := cos[k]
					if el.selected {
						r, g, b = 255, 50, 50
					} else {
						r, g, b = 155, 155, 155
					}
					gl.Color3ub(r, g, b)
					gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
				}
				gl.End()
			case colorEdgeElements:
				gl.LineWidth(3)
				gl.Enable(gl.LINE_SMOOTH)
				gl.Begin(gl.LINES)
				for p, k := range el.Indexes {
					c := cos[k]
					if el.selected {
						r, g, b = 255, 50, 50
					} else {
						r, g, b = edgeColor(p)
					}
					gl.Color3ub(r, g, b)
					gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
				}
				gl.End()
			case selectPoints:
				// do nothing
			case selectLines:
				gl.LineWidth(3)
				r, g, b = convertToColor(iel)
				gl.Color3ub(r, g, b)
				if fill {
					gl.Begin(gl.LINES)
					for _, k := range el.Indexes {
						c := cos[k]
						gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
					}
					gl.End()
				} else {
					randomPoint(iel)
				}
			case selectTriangles:
				// do nothing
			case selectQuadrs:
				// do nothing
			default:
				logger.Printf("undefined type: %v", s)
			}
		///////////////////////////////////
		case Triangle3, Quadr4:
			switch s {
			case normal:
				// COMMENTED FOR PERFOMANCE :
				// gl.Begin(gl.POLYGON)
				// for _, k := range el.Indexes {
				// 	c := cos[k]
				// 	if el.selected {
				// 		r, g, b = 255, 90, 90
				// 	} else {
				// 		r, g, b = 155, 0, 155
				// 	}
				// 	gl.Color3ub(r, g, b)
				// 	gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
				// }
				// gl.End()
				// borders
				var mid [3]float64
				for _, k := range el.Indexes {
					c := cos[k]
					for p := 0; p < 3; p++ {
						mid[p] += c.Point3d[p]
					}
				}
				for p := 0; p < 3; p++ {
					mid[p] /= float64(len(el.Indexes))
				}
				if el.selected {
					r, g, b = 235, 70, 70
				} else {
					r, g, b = 123, 0, 123
				}
				gl.Color3ub(r, g, b)
				gl.LineWidth(1)
				gl.Disable(gl.LINE_SMOOTH)

				ratio := 0.1
				for p := range el.Indexes {
					gl.Begin(gl.LINES)

					from, to := p, p+1
					if to == len(el.Indexes) {
						from = el.Indexes[from]
						to = el.Indexes[0]
					} else {
						from = el.Indexes[from]
						to = el.Indexes[to]
					}
					gl.Vertex3d(
						ratio*mid[0]+(1-ratio)*cos[from].Point3d[0],
						ratio*mid[1]+(1-ratio)*cos[from].Point3d[1],
						ratio*mid[2]+(1-ratio)*cos[from].Point3d[2],
					)
					gl.Vertex3d(
						ratio*mid[0]+(1-ratio)*cos[to].Point3d[0],
						ratio*mid[1]+(1-ratio)*cos[to].Point3d[1],
						ratio*mid[2]+(1-ratio)*cos[to].Point3d[2],
					)
					gl.End()
				}
			case colorEdgeElements:
				gl.Begin(gl.POLYGON)
				for p, k := range el.Indexes {
					c := cos[k]
					if el.selected {
						r, g, b = 255, 90, 90
					} else {
						r, g, b = edgeColor(p)
					}
					gl.Color3ub(r, g, b)
					gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
				}
				gl.End()
			case selectPoints:
				// do nothing
			case selectLines:
				// do nothing
			case selectTriangles, selectQuadrs:
				if (s == selectTriangles && el.ElementType != Triangle3) ||
					(s == selectQuadrs && el.ElementType != Quadr4) {
					break
				}
				r, g, b = convertToColor(iel)
				gl.Color3ub(r, g, b)
				if fill {
					gl.Begin(gl.POLYGON)
					for _, k := range el.Indexes {
						c := cos[k]
						gl.Vertex3d(c.Point3d[0], c.Point3d[1], c.Point3d[2])
					}
					gl.End()
				} else {
					randomPoint(iel)
				}
			default:
				logger.Printf("undefined type: %v", s)
			}
		///////////////////////////////////
		default:
			logger.Printf("undefined type: %v", s)
			// switch s {
			// case normal:
			// 	if el.selected {
			// 	} else {
			// 	}
			// case colorEdgeElements:
			// 	if el.selected {
			// 	} else {
			// 	}
			// case selectPoints:
			// case selectLines:
			// case selectTriangles:
			// case selectQuadrs:
			// }
		}
	}
}

func openGlScreenCoordinate(x, y, w, h int32) { // window *glfw.Window) {
	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.TEXTURE_2D)

	// w, h := window.GetSize()
	gl.Viewport(int32(x), int32(y), int32(w), int32(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(w), 0, float64(h), float64(-100.0), float64(100.0))

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func (op *Opengl) drawAxes(w, h int32) {
	gl.Disable(gl.LINE_SMOOTH)
	gl.LineWidth(1) // default lines width
	// w, h := op.window.GetSize()

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

type viewState uint64

const (
	normal            viewState = 1 << iota // 1
	colorEdgeElements                       // 2
	selectPoints                            // 4
	selectLines                             // 8
	selectTriangles                         // 16
	selectQuadrs                            // 32
)

type selectState bool

const (
	filling     selectState = true
	randomPoint             = false
)

func (s viewState) String() string {
	switch s {
	case normal:
		return "Normal state"
	case colorEdgeElements:
		return "Color edge state"
	case selectPoints:
		return "points"
	case selectLines:
		return "lines"
	case selectTriangles:
		return "triangles"
	case selectQuadrs:
		return "quadrs"
	}
	return fmt.Sprintf("%d", s)
}

func edgeColor(pos int) (r, g, b uint8) {
	switch pos {
	case 0: // yellow
		r, g, b = 255, 255, 0
		return
	case 1: // blue
		r, g, b = 0, 0, 255
		return
	case 2: // green
		r, g, b = 0, 255, 0
		return
	case 3: // purple
		r, g, b = 255, 0, 125
		return
	}
	logger.Printf("not valid pos: %d", pos)
	return 100, 100, 100
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

func convertToColor(i int) (r, g, b uint8) {
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
	r = uint8(value[0])
	g = uint8(value[1])
	b = uint8(value[2])
	return
}

// func (op *Opengl) mouseButton(
// 	w *glfw.Window,
// 	button glfw.MouseButton,
// 	action glfw.Action,
// 	mods glfw.ModifierKey,
// ) {
// 	var index int
// 	switch button {
// 	case glfw.MouseButton1:
// 		index = 0
// 	case glfw.MouseButton2:
// 		index = 1
// 	case glfw.MouseButton3:
// 		index = 2
// 	default:
// 		return
// 	}
// 	for i := range op.mouses {
// 		if op.mouses[i] == nil {
// 			continue
// 		}
// 		if i != index {
// 			op.mouses[i].Reset()
// 			continue
// 		}
// 		x, y := w.GetCursorPos()
// 		_, h := w.GetSize()
// 		y = float64(h) - y
// 		switch action {
// 		case glfw.Press:
// 			op.mouses[index].Press(int32(x), int32(y))
// 		case glfw.Release:
// 			op.mouses[index].Release(int32(x), int32(y))
// 		default:
// 			// case glfw.Repeat:
// 			// do nothing
// 		}
// 		return
// 	}
// }

// func (op *Opengl) cursorPos(w *glfw.Window, xpos, ypos float64) {
// 	_, h := w.GetSize()
// 	ypos = float64(h) - ypos
// 	for i := range op.mouses {
// 		if op.mouses[i] == nil {
// 			continue
// 		}
// 		op.mouses[i].Update(int32(xpos), int32(ypos))
// 	}
// }

// func (op *Opengl) key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
// 	switch key {
// 	case glfw.KeyEscape:
// 		// deselect all
// 		op.Init()
// 		op.mesh.DeselectAll()
// 		for i := range op.mouses {
// 			if op.mouses[i] == nil {
// 				continue
// 			}
// 			op.mouses[i].Reset()
// 		}
// 		op.MouseDefault()
// 	}
// }

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
	case StandardViewIsometric:
		op.camera.alpha = -45.0
		op.camera.betta = 30.0
	}
}

func (op *Opengl) ColorEdge(isColor bool) {
	if isColor {
		op.state = colorEdgeElements
	} else {
		op.state = normal
	}
}

func (op *Opengl) SelectLeftCursor(nodes bool, elements []bool) {
	op.cursorLeft = 0
	if nodes {
		op.cursorLeft |= selectPoints
	}
	for el := Line2; el < lastElement; el = el + 1 {
		if !elements[int(el)] {
			continue
		}
		op.cursorLeft |= el.getSelect()
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
	AddLinesLC LeftCursor = iota
	AddTrianglesLC
	AddQuardsLC
	endLC
)

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
	WithShiftKey(key bool)
	Press(x, y int32)
	Update(x, y int32)
	Release(x, y int32)

	ReadyPreview() bool
	Preview(xinit, yinit int32)

	ReadyAction() bool
	Action(op *Opengl)

	Reset()

	Status() string
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

func (m2 *Mouse2P) WithShiftKey(key bool) {
	// do nothing
}

func (m2 *Mouse2P) Status() string {
	// do nothing
	return ""
}

type MouseSelect struct {
	AddShift bool
	Mouse2P
}

func (ms *MouseSelect) WithShiftKey(key bool) {
	ms.AddShift = key
}

func (ms *MouseSelect) Preview(xinit, yinit int32) {
	if !ms.ReadyPreview() {
		return
	}
	var (
		x1 = ms.from[0]
		y1 = ms.from[1]
		x2 = ms.to[0]
		y2 = ms.to[1]
	)

	// draw select rectangle
	if x2 < x1 {
		gl.Color3ub(255, 0, 0) // Red
		gl.Begin(gl.LINES)
		defer func() {
			gl.End()
		}()
	} else {
		gl.LineStipple(1, 0x00FF)
		gl.Enable(gl.LINE_STIPPLE)
		gl.Begin(gl.LINE_STRIP)
		defer func() {
			gl.End()
			gl.Disable(gl.LINE_STIPPLE)
		}()
		gl.Color3ub(0, 0, 255) // Blue
	}
	gl.LineWidth(1)
	gl.Vertex2i(x1, y1)
	gl.Vertex2i(x1, y2)

	gl.Vertex2i(x1, y2)
	gl.Vertex2i(x2, y2)

	gl.Vertex2i(x2, y2)
	gl.Vertex2i(x2, y1)

	gl.Vertex2i(x2, y1)
	gl.Vertex2i(x1, y1)
}

func (ms *MouseSelect) Action(op *Opengl) {
	if !ms.ReadyAction() {
		return
	}
	defer ms.Reset()

	leftToRight := ms.from[0] < ms.to[0]

	// DEBUG : start := time.Now()
	// DEBUG : logger.Printf("MouseSelect time %v", time.Now())
	// DEBUG : defer func() {
	// DEBUG : 	logger.Printf("MouseSelect duration summary : %v", time.Since(start))
	// DEBUG : }()

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
	ms.from[0] += op.x
	ms.from[1] += op.y
	ms.to[0] += op.x
	ms.to[1] += op.y

	cos, els := op.mesh.GetCoords(), op.mesh.GetElements()

	// store present selections
	selP := make([]bool, len(cos))
	for i := range cos {
		selP[i] = cos[i].selected
		cos[i].selected = false
	}
	selE := make([]bool, len(els))
	for i := range els {
		selE[i] = els[i].selected
		els[i].selected = false
	}

	var s []bool
	var added bool
	state := op.cursorLeft
	if leftToRight && op.cursorLeft&selectPoints == 0 {
		added = true
		op.cursorLeft |= selectPoints
		ns := op.mesh.GetCoords()
		s = make([]bool, len(ns))
		for i := range ns {
			s[i] = ns[i].selected
		}
	}

	// find new selected
	for is, s := range []struct {
		st viewState
		sf func(index int) (found bool)
	}{
		{st: selectPoints, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(cos) <= index {
				logger.Printf("selectPoints index outside: %d", index)
				return false
			}
			cos[index].selected = true
			return true
		}}, {st: selectLines, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(els) <= index {
				logger.Printf("selectLines index outside: %d", index)
				return false
			}
			if els[index].ElementType != Line2 {
				logger.Printf("selectLines index is not line: %d. %v",
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
				logger.Printf("selectTriangles index outside: %d", index)
				return false
			}
			if els[index].ElementType != Triangle3 {
				logger.Printf("selectTriangles index is not triangle: %d", index)
				return false
			}
			els[index].selected = true
			return true
		}}, {st: selectQuadrs, sf: func(index int) bool {
			if index < 0 {
				return false
			}
			if len(els) <= index {
				logger.Printf("selectQuadrs index outside: %d", index)
				return false
			}
			if els[index].ElementType != Quadr4 {
				logger.Printf("selectQuadrs index is not triangle: %d", index)
				return false
			}
			els[index].selected = true
			return true
		}},
	} {
		if op.cursorLeft&s.st == 0 {
			continue
		}
		_ = is // for debugging

		// find selection
		for _, fill := range []selectState{randomPoint, filling} {
			// DEBUG : logger.Printf("MouseSelect step %d type %v {%v} start",
			// DEBUG : 	is, s.st, fill)
			found := true
			for found { // TODO : infinite loop
				found = false
				gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
				// gl.ClearColorxOES(0, 0, 0, 0) // ???? TODO ??
				gl.ClearColor(1, 1, 1, 1)
				gl.Enable(gl.DEPTH_TEST)
				// TODO  gl.Disable(gl.DEPTH_TEST)
				gl.Disable(gl.LIGHTING)
				gl.Disable(gl.LINE_SMOOTH)
				gl.Disable(gl.BLEND) // Transparency
				op.cameraView(op.x, op.y, op.w, op.h)
				// color initialize

				op.model3d(s.st, fill)

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
			// DEBUG : logger.Printf("MouseSelect step %d type %v {%v} duration: %v",
			// DEBUG : 	is, s.st, fill, time.Since(start))
		}
	}

	if leftToRight {
		// find real selected elements, if all they coordinate selected
		for i := range els {
			if !els[i].selected {
				continue
			}
			// element selected
			for _, index := range els[i].Indexes {
				if cos[index].selected {
					continue
				}
				els[i].selected = false
				break
			}
		}
		// select coordinates by selected elements
		for i := range els {
			if !els[i].selected {
				continue
			}
			// element selected
			for _, index := range els[i].Indexes {
				cos[index].selected = true
			}
		}
		if added {
			op.cursorLeft = state
			ns := op.mesh.GetCoords()
			for i := range ns {
				ns[i].selected = s[i]
			}
		}
	}

	if ms.AddShift {
		// add to selections
		for i := range cos {
			cos[i].selected = selP[i] || cos[i].selected
		}
		for i := range els {
			els[i].selected = selE[i] || els[i].selected
		}
	}
}

type MouseRotate struct {
	Mouse2P
}

func (mr *MouseRotate) Preview(xinit, yinit int32) {}

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

func (mr *MouseMove) Preview(xinit, yinit int32) {
	if !mr.ReadyPreview() {
		return
	}
	var (
		x1 = mr.from[0]
		y1 = mr.from[1]
		x2 = mr.to[0]
		y2 = mr.to[1]
	)

	// draw select rectangle
	gl.LineWidth(1)

	gl.Begin(gl.LINES)
	gl.Color3d(1.0, 0.4, 0.4)
	gl.Vertex2i(x1-5, y1)
	gl.Vertex2i(x1+5, y1)
	gl.End()
	gl.Begin(gl.LINES)
	gl.Color3d(1.0, 0.4, 0.4)
	gl.Vertex2i(x1, y1-5)
	gl.Vertex2i(x1, y1+5)
	gl.End()

	gl.Begin(gl.LINES)
	gl.Color3d(1.0, 0.0, 0.0) // Red
	gl.Vertex2i(x1, y1)
	gl.Vertex2i(x2, y2)
	gl.End()
}

func (mr *MouseMove) Action(op *Opengl) {
	if !mr.ReadyAction() {
		return
	}
	defer mr.Reset()
	// action
	const factor = 0.15
	dx := float64(mr.from[0] - mr.to[0])
	dy := float64(mr.from[1] - mr.to[1])
	g := math.Sqrt(pow.E2(dx) + pow.E2(dy))
	if g < 10 { // minimal amount pixel distance
		return
	}
	dx /= g
	dy /= g
	op.camera.moveX -= op.camera.R * factor * dx
	op.camera.moveY -= op.camera.R * factor * dy
}

type MouseAdd struct {
	MouseSelect

	LC LeftCursor
	ps []uint
}

func (ma *MouseAdd) Status() string {
	return fmt.Sprintf("Added:%d", len(ma.ps))
}

func (ma *MouseAdd) Action(op *Opengl) {
	if !ma.ReadyAction() {
		return
	}
	defer ma.MouseSelect.Reset()

	// store node
	op.mesh.DeselectAll()
	ma.MouseSelect.Action(op)
	ids := op.mesh.GetSelectNodes(Single)
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

	for i := range valids {
		if valids[i].lc != ma.LC {
			continue
		}
		if len(ma.ps) != valids[i].amount {
			return
		} else {
			break
		}
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
	case AddQuardsLC:
		op.mesh.AddQuadr4ByNodeNumber(
			ma.ps[0],
			ma.ps[1],
			ma.ps[2],
			ma.ps[3],
		)
	}
	ma.Reset()
}

func (ma *MouseAdd) Reset() {
	ma.ps = nil
	ma.MouseSelect.Reset()
}

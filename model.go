package ms

import "math"

type selectable struct {
	Selected bool
}

type Point struct {
	selectable
	X, Y, Z float64
}

type Line struct {
	selectable
	Index [2]int
}

type Triangle struct {
	selectable
	Index [3]int
}

var model Model

type Model struct {
	Points    []Point
	Lines     []Line
	Triangles []Triangle
}

func init() { // TODO remove
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 256
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)

	for i := 0; i < levels; i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0
		model.Points = append(model.Points,
			Point{X: Ri * math.Sin(angle), Y: float64(i) * dy, Z: Ri * math.Cos(angle)},
			Point{X: Ro * math.Sin(angle), Y: float64(i) * dy, Z: Ro * math.Cos(angle)},
		)
		model.Lines = append(model.Lines, Line{Index: [2]int{2 * i, 2*i + 1}})
		if 0 < i {
			model.Lines = append(model.Lines,
				Line{Index: [2]int{2 * (i - 1), 2 * i}},
				Line{Index: [2]int{2*(i-1) + 1, 2*i + 1}})
			model.Triangles = append(model.Triangles,
				Triangle{Index: [3]int{2 * (i - 1), 2 * i, 2*(i-1) + 1}},
				Triangle{Index: [3]int{2 * i, 2*(i-1) + 1, 2*i + 1}})
		}
	}
	updateModel = true // TODO  remove
}

func (m *Model) AddNode(X, Y, Z float64) {
	m.Points = append(m.Points, Point{X: X, Y: Y, Z: Z})
}

// func (m *Model) AddNodeByDistance(line, distance string, atBegin bool) {
// }

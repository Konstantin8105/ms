package ms

import "math"

type Point struct{X,Y,Z float64}

var model Model

type Model struct {
	Points    []Point
	Lines     [][2]int
	Triangles [][3]int
	Quadr4s   [][4]int
}

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
	ps:= make([]Point, len_ps)
	ls:= make([][2]int, len_ls)
	ts:= make([][3]int, len_ts)
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
	// TODO updateModel = true

	model.Points = ps
	model.Lines = ls
	model.Triangles = ts
}

func (m *Model) AddNode(X, Y, Z float64) {
	m.Points = append(m.Points, Point{X: X, Y: Y, Z: Z})
}

// func (m *Model) AddNodeByDistance(line, distance string, atBegin bool) {
// }

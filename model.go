package ms

import (
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Konstantin8105/pow"
)

// 3D model variables
type object3d struct {
	selected bool
	hided    bool
}

type ElType uint8

const (
	Line2     ElType = iota + 1 // 1
	Triangle3                   // 2
	ElRemove                    // 3
)

// Element is typical element for FEM. Examples:
//
//	Line o======o                                             //
//	ElType : 2                                                //
//	Indexes: 2 (amount indexes of coordinates)                //
//
//	Triangle o======o                                         //
//	          \    /                                          //
//	           \  /                                           //
//	            o                                             //
//	ElType : 3                                                //
//	Indexes: 3 (amount indexes of coordinates)                //
//
//	Quadr4 o======o                                           //
//	       |      |                                           //
//	       |      |                                           //
//	       o======o                                           //
//	ElType : 4                                                //
//	Indexes: 4 (amount indexes of coordinates)                //
//
type Element struct {
	object3d
	ElementType ElType
	Indexes     []int // index of coordinate
}

// valid matrix element constants
var valid = [...][2]int{{int(Line2), 2}, {int(Triangle3), 3}, {int(ElRemove), 0}}

func (e Element) Check() error {
	for i := range valid {
		if int(e.ElementType) == valid[i][0] && len(e.Indexes) != valid[i][1] {
			return fmt.Errorf("unacceptable element: %v", e)
		}
	}
	return fmt.Errorf("undefined element: %v", e)
}

// Coordinate store coordinate of points
type Coordinate struct {
	object3d
	Removed bool // TODO check everywhere
	X, Y, Z float64

	// TODO
	// index int    // index of Models
	// C [3]float64 // coordinates
}

// Named intermediant named structure
type Named struct{ Name string }
type Ignored struct{ IgnoreElements []bool }

// TODO : type MultiModel struct { Models []Model}

type Model struct {
	// actual = 0, then change Model
	// 0 < actual, then change Parts[actual - 1]
	actual int

	Named
	Ignored
	Elements []Element
	Coords   []Coordinate

	Parts []Part
}

type Part struct {
	Named
	Ignored
}

func (mm *Model) Undo() {}
func (mm *Model) Redo() {}

func clearPartName(name *string) {
	*name = strings.ReplaceAll(*name, "\n", "")
	*name = strings.ReplaceAll(*name, "\r", "")
	*name = strings.ReplaceAll(*name, "\t", "")
}

func (mm *Model) PartsName() (names []string) {
	names = append(names, mm.Name)
	for i := range mm.Parts {
		names = append(names, mm.Parts[i].Name)
	}
	return
}

func (mm *Model) PartPresent() (id uint) {
	return uint(mm.actual)
}

func (mm *Model) PartChange(id uint) {
	if id == 0 {
		mm.actual = 0
		return
	}
	id = id - 1 // convert to part indexes
	if int(id) <= len(mm.Parts) {
		mm.actual = int(id) + 1
	}
	// no changes
}

func (mm *Model) PartNew(name string) {
	clearPartName(&name)
	var p Part
	p.Name = name
	mm.Parts = append(mm.Parts, p)
	mm.actual = len(mm.Parts) // no need `-1`, because base model
}

func (mm *Model) PartRename(id uint, name string) {
	clearPartName(&name)
	if id == 0 {
		mm.Name = name
		return
	}
	if len(mm.Parts) < int(id) {
		return
	}
	mm.Parts[id-1].Name = name
}

func (mm *Model) AddModel(m Model) {
	newID := make([]int, len(m.Coords))
	for i := range m.Coords {
		if m.Coords[i].Removed {
			continue
		}
		id := mm.AddNode(
			m.Coords[i].X,
			m.Coords[i].Y,
			m.Coords[i].Z,
		)
		newID[i] = int(id)
	}
	for k := range m.Elements {
		el := m.Elements[k]
		for p := range el.Indexes {
			el.Indexes[p] = newID[el.Indexes[p]]
		}
	}
	for i := range m.Elements {
		el := m.Elements[i]
		switch el.ElementType {
		case ElRemove:
			// do nothing
		case Line2:
			mm.AddLineByNodeNumber(
				uint(el.Indexes[0]),
				uint(el.Indexes[1]),
			)
		case Triangle3:
			mm.AddTriangle3ByNodeNumber(
				uint(el.Indexes[0]),
				uint(el.Indexes[1]),
				uint(el.Indexes[2]),
			)
		default:
			panic(fmt.Errorf("not implemented %v", el))
		}
	}
}

func (mm *Model) DemoSpiral() {
	var m Model
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 25
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)
	for i := 0; i < levels; i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0
		m.Coords = append(m.Coords,
			Coordinate{X: Ri * math.Sin(angle), Y: float64(i) * dy, Z: Ri * math.Cos(angle)},
			Coordinate{X: Ro * math.Sin(angle), Y: float64(i) * dy, Z: Ro * math.Cos(angle)},
		)
		m.Elements = append(m.Elements, Element{ElementType: Line2,
			Indexes: []int{2 * i, 2*i + 1},
		})
		if 0 < i {
			m.Elements = append(m.Elements,
				Element{ElementType: Line2,
					Indexes: []int{2 * (i - 1), 2 * i},
				}, Element{ElementType: Line2,
					Indexes: []int{2*(i-1) + 1, 2*i + 1},
				})
			m.Elements = append(m.Elements,
				Element{ElementType: Triangle3,
					Indexes: []int{2 * (i - 1), 2 * i, 2*(i-1) + 1},
				}, Element{ElementType: Triangle3,
					Indexes: []int{2 * i, 2*(i-1) + 1, 2*i + 1},
				})
		}
	}
	mm.AddModel(m)
}

const distanceError = 1e-6

func (mm *Model) AddNode(X, Y, Z float64) (id uint) {
	// check is this coordinate exist?
	for i := range mm.Coords {
		if mm.Coords[i].Removed {
			continue
		}
		// fast algorithm
		dX := mm.Coords[i].X - X
		if distanceError < math.Abs(dX) {
			continue
		}
		dY := mm.Coords[i].Y - Y
		if distanceError < math.Abs(dY) {
			continue
		}
		dZ := mm.Coords[i].Z - Z
		distance := math.Sqrt(pow.E2(dX) + pow.E2(dY) + pow.E2(dZ))
		if distance < distanceError {
			return uint(i)
		}
	}
	// append
	mm.Coords = append(mm.Coords, Coordinate{X: X, Y: Y, Z: Z})
	return uint(len(mm.Coords) - 1)
}

func (mm *Model) AddLineByNodeNumber(n1, n2 uint) (id uint) {
	// type convection
	ni1 := int(n1)
	ni2 := int(n2)
	// check is this coordinate exist?
	for i, el := range mm.Elements {
		if el.ElementType != Line2 {
			continue
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni2 {
			return uint(i)
		}
		if el.Indexes[1] == ni1 && el.Indexes[0] == ni2 {
			return uint(i)
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Line2,
		Indexes:     []int{ni1, ni2},
	})
	return uint(len(mm.Elements) - 1)
}

func (mm *Model) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint) {
	// type convection
	ni1 := int(n1)
	ni2 := int(n2)
	ni3 := int(n3)
	// check is this coordinate exist?
	for i, el := range mm.Elements {
		if el.ElementType != Triangle3 {
			continue
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni2 && el.Indexes[2] == ni3 {
			return uint(i)
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni3 && el.Indexes[2] == ni1 {
			return uint(i)
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni1 && el.Indexes[2] == ni2 {
			return uint(i)
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni2 && el.Indexes[2] == ni1 {
			return uint(i)
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni1 && el.Indexes[2] == ni3 {
			return uint(i)
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni3 && el.Indexes[2] == ni2 {
			return uint(i)
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Triangle3,
		Indexes:     []int{ni1, ni2, ni3},
	})
	return uint(len(mm.Elements) - 1)
}

func (mm *Model) AddLeftCursor(lc LeftCursor) {
	AddInfo("Model not implemented AddLeftCursor: %v", lc)
}

func (mm *Model) GetCoords() []Coordinate {
	return mm.Coords
}

func (mm *Model) GetElements() []Element {
	return mm.Elements
}

func (mm *Model) Remove(nodes, elements []uint) {
	// it is part/model
	// do not remove nodes in ignore list
	ignore := make([]bool, len(nodes))
	for ind, p := range nodes {
		if mm.Coords[int(p)].hided || mm.Coords[int(p)].Removed {
			ignore[ind] = true
			continue
		}
		for i := range mm.Elements {
			if mm.Elements[i].ElementType == ElRemove {
				continue
			}
			if !mm.IsIgnore(uint(i)) {
				continue
			}
			// ignored coordinate on ignored elements
			for k := range mm.Elements[i].Indexes {
				if mm.Elements[i].Indexes[k] == int(p) {
					ignore[ind] = true
				}
			}
		}
	}
	// remove
	for ind, p := range nodes {
		if ignore[ind] {
			continue
		}
		// removing coordinates
		mm.Coords[p].Removed = true
		// remove elements with coordinate
		for i := range mm.Elements {
			if mm.IsIgnore(uint(i)) {
				continue
			}
			// ignored coordinate on ignored elements
			for k := range mm.Elements[i].Indexes {
				if mm.Elements[i].Indexes[k] == int(p) {
					elements = append(elements, uint(i))
					break
				}
			}
		}
	}
	// remove elements
	for _, p := range elements {
		mm.Elements[p].ElementType = ElRemove
		mm.Elements[p].Indexes = nil
	}
}

func (mm *Model) IsIgnore(elID uint) bool {
	if 0 < mm.actual && int(elID) < len(mm.Parts[mm.actual-1].IgnoreElements) {
		// it is part
		return mm.Parts[mm.actual-1].IgnoreElements[int(elID)]
	}
	if int(elID) < len(mm.IgnoreElements) {
		return mm.IgnoreElements[int(elID)]
	}
	return false
}

func (mm *Model) ColorEdge(isColor bool) {
	AddInfo("Model not implemented ColorEdge: %v", isColor)
}

func (mm *Model) IgnoreModelElements(ids []uint) {
	if len(ids) == 0 {
		return
	}
	ignore := &mm.IgnoreElements
	if 0 < mm.actual {
		ignore = &mm.Parts[mm.actual-1].IgnoreElements
	}
	if len(mm.Elements) < len(*ignore) {
		*ignore = (*ignore)[:len(mm.Elements)]
	}
	if len(*ignore) != len(mm.Elements) {
		*ignore = append(*ignore, make([]bool, len(mm.Elements)-len(*ignore))...)
	}
	for _, p := range ids {
		(*ignore)[p] = true
	}
}

func (mm *Model) Unignore() {
	ignore := &mm.IgnoreElements
	if 0 < mm.actual {
		ignore = &mm.Parts[mm.actual-1].IgnoreElements
	}
	*ignore = nil
}

func (mm *Model) SelectLeftCursor(nodes, lines, tria bool) {
	AddInfo("Model not implemented SelectLeftCursor: %v %v %v",
		nodes, lines, tria)
}

func (mm *Model) SelectNodes(single bool) (ids []uint) {
	for i := range mm.Coords {
		if !mm.Coords[i].selected {
			continue
		}
		if mm.Coords[i].Removed {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectLines(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !vis || !ok {
			continue
		}
		if !el.selected {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectTriangles(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		if !el.selected {
			continue
		}
		if el.ElementType != Triangle3 {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectElements(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		if el.ElementType == ElRemove {
			continue
		}
		if !el.selected {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) InvertSelect(nodes, lines, triangles bool) {
	if nodes {
		for i := range mm.Coords {
			if mm.Coords[i].Removed {
				continue
			}
			mm.Coords[i].selected = !mm.Coords[i].selected
		}
	}
	if lines {
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != Line2 {
				continue
			}
			mm.Elements[i].selected = !mm.Elements[i].selected
		}
	}
	if triangles {
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != Triangle3 {
				continue
			}
			mm.Elements[i].selected = !mm.Elements[i].selected
		}
	}
}

func (mm *Model) SelectLinesOrtho(x, y, z bool) {
	if !x && !y && !z {
		return
	}
	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !vis || !ok {
			continue
		}
		if el.selected {
			continue
		}
		var (
			dx = math.Abs(mm.Coords[el.Indexes[1]].X - mm.Coords[el.Indexes[0]].X)
			dy = math.Abs(mm.Coords[el.Indexes[1]].Y - mm.Coords[el.Indexes[0]].Y)
			dz = math.Abs(mm.Coords[el.Indexes[1]].Z - mm.Coords[el.Indexes[0]].Z)
		)
		if x && dy < distanceError && dz < distanceError {
			mm.Elements[i].selected = true
		}
		if y && dx < distanceError && dz < distanceError {
			mm.Elements[i].selected = true
		}
		if z && dx < distanceError && dy < distanceError {
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectLinesOnPlane(xoy, yoz, xoz bool) {
	if !xoy && !yoz && !xoz {
		return
	}
	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !vis || !ok {
			continue
		}
		if el.selected {
			continue
		}
		var (
			dX = math.Abs(mm.Coords[el.Indexes[1]].X - mm.Coords[el.Indexes[0]].X)
			dY = math.Abs(mm.Coords[el.Indexes[1]].Y - mm.Coords[el.Indexes[0]].Y)
			dZ = math.Abs(mm.Coords[el.Indexes[1]].Z - mm.Coords[el.Indexes[0]].Z)
		)
		if xoy && dZ < distanceError {
			mm.Elements[i].selected = true
		}
		if yoz && dX < distanceError {
			mm.Elements[i].selected = true
		}
		if xoz && dY < distanceError {
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectLinesParallel(lines []uint) {
	// check input data
	for _, p := range lines {
		if len(mm.Elements) <= int(p) || int(p) < 0 {
			AddInfo("SelectLinesParallel: not valid index %d", p)
			return
		}
		if mm.Elements[p].ElementType != Line2 {
			AddInfo("SelectLinesParallel: is not line %v", mm.Elements[p])
			return
		}
	}
	// selection
	type ratio struct{ dx, dy, dz float64 }

	toOne := func(el Element) (r ratio, ok bool) {
		r.dx = mm.Coords[el.Indexes[0]].X - mm.Coords[el.Indexes[1]].X
		r.dy = mm.Coords[el.Indexes[0]].Y - mm.Coords[el.Indexes[1]].Y
		r.dz = mm.Coords[el.Indexes[0]].Z - mm.Coords[el.Indexes[1]].Z
		if math.Abs(r.dx) < distanceError &&
			math.Abs(r.dy) < distanceError &&
			math.Abs(r.dz) < distanceError {
			// ignore zero lines
			return
		}
		amplitude := math.Sqrt(pow.E2(r.dx) + pow.E2(r.dy) + pow.E2(r.dz))
		r.dx /= amplitude
		r.dy /= amplitude
		r.dz /= amplitude
		return r, true
	}

	compare := func(r1, r2 ratio) bool {
		if distanceError < math.Abs(r1.dx-r2.dx) {
			return false
		}
		if distanceError < math.Abs(r1.dy-r2.dy) {
			return false
		}
		if distanceError < math.Abs(r1.dz-r2.dz) {
			return false
		}
		return true
	}

	var ratios []ratio
	for _, p := range lines {
		vis, ok := mm.IsVisibleLine(p)
		if !vis || !ok {
			continue
		}
		r, ok := toOne(mm.Elements[p])
		if !ok {
			continue
		}
		ratios = append(ratios, r)
	}

	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !ok || !vis {
			continue
		}
		if el.selected {
			continue
		}
		var found bool
		for _, p := range lines {
			if int(p) == i {
				found = true
				mm.Elements[p].selected = true
				break
			}
		}
		if found {
			continue
		}
		re, ok := toOne(el)
		if !ok {
			continue
		}
		for ri := range ratios {
			if compare(re, ratios[ri]) {
				mm.Elements[i].selected = true
			}
		}
	}
}

func (mm *Model) SelectLinesByLenght(more bool, lenght float64) {
	if lenght <= 0.0 {
		return
	}
	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !vis || !ok {
			continue
		}
		if el.selected {
			continue
		}
		var (
			dx = mm.Coords[el.Indexes[0]].X - mm.Coords[el.Indexes[1]].X
			dy = mm.Coords[el.Indexes[0]].Y - mm.Coords[el.Indexes[1]].Y
			dz = mm.Coords[el.Indexes[0]].Z - mm.Coords[el.Indexes[1]].Z
		)
		if math.Abs(dx) < distanceError &&
			math.Abs(dy) < distanceError &&
			math.Abs(dz) < distanceError {
			// ignore zero lines
			continue
		}
		L := math.Sqrt(pow.E2(dx) + pow.E2(dy) + pow.E2(dz))
		if (more && lenght <= L) || (!more && L <= lenght) {
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectLinesCylindrical(node uint, radiant, conc bool, axe Direction) {
	if int(node) < 0 || len(mm.Coords) <= int(node) {
		AddInfo("SelectLinesCylindrical: not valid node %d", node)
		return
	}
	for i, el := range mm.Elements {
		vis, ok := mm.IsVisibleLine(uint(i))
		if !vis || !ok {
			continue
		}
		if el.selected {
			continue
		}
		var r0, r1, dr float64
		var a0, a1, da float64
		switch axe {
		case DirX:
			r0 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[0]].Z-mm.Coords[node].Z) +
					pow.E2(mm.Coords[el.Indexes[0]].Y-mm.Coords[node].Y))
			r1 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[1]].Z-mm.Coords[node].Z) +
					pow.E2(mm.Coords[el.Indexes[1]].Y-mm.Coords[node].Y))
			a0 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[0]].Z-mm.Coords[node].Z),
				pow.E2(mm.Coords[el.Indexes[0]].Y-mm.Coords[node].Y))
			a1 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[1]].Z-mm.Coords[node].Z),
				pow.E2(mm.Coords[el.Indexes[1]].Y-mm.Coords[node].Y))
		case DirY:
			r0 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[0]].X-mm.Coords[node].X) +
					pow.E2(mm.Coords[el.Indexes[0]].Z-mm.Coords[node].Y))
			r1 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[1]].X-mm.Coords[node].X) +
					pow.E2(mm.Coords[el.Indexes[1]].Z-mm.Coords[node].Z))
			a0 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[0]].X-mm.Coords[node].X),
				pow.E2(mm.Coords[el.Indexes[0]].Z-mm.Coords[node].Z))
			a1 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[1]].X-mm.Coords[node].X),
				pow.E2(mm.Coords[el.Indexes[1]].Z-mm.Coords[node].Z))
		case DirZ:
			r0 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[0]].X-mm.Coords[node].X) +
					pow.E2(mm.Coords[el.Indexes[0]].Y-mm.Coords[node].Y))
			r1 = math.Sqrt(
				pow.E2(mm.Coords[el.Indexes[1]].X-mm.Coords[node].X) +
					pow.E2(mm.Coords[el.Indexes[1]].Y-mm.Coords[node].Y))
			a0 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[0]].X-mm.Coords[node].X),
				pow.E2(mm.Coords[el.Indexes[0]].Y-mm.Coords[node].Y))
			a1 = math.Atan2(
				pow.E2(mm.Coords[el.Indexes[1]].X-mm.Coords[node].X),
				pow.E2(mm.Coords[el.Indexes[1]].Y-mm.Coords[node].Y))
		}
		dr = math.Abs(r0 - r1)
		da = math.Abs(a0 - a1)
		if da < distanceError && radiant {
			mm.Elements[i].selected = true
		}
		if dr < distanceError && conc {
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) IsVisibleLine(p uint) (visible, ok bool) {
	if int(p) < 0 || len(mm.Elements) <= int(p) {
		AddInfo("IsVisibleLine: not valid index %d", p)
		return
	}
	if mm.Elements[p].ElementType != Line2 {
		return
	}
	ok = true
	if mm.Elements[p].hided {
		return
	}
	if mm.IsIgnore(uint(p)) {
		return
	}
	visible = true
	return
}

func (mm *Model) DeselectAll() {
	// deselect all
	for i := range mm.Coords {
		mm.Coords[i].selected = false
	}
	for i := range mm.Elements {
		mm.Elements[i].selected = false
	}
}

func (mm *Model) SelectAll(nodes, lines, tria bool) {
	if nodes {
		for i := range mm.Coords {
			mm.Coords[i].selected = true
		}
	}
	if lines {
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != Line2 {
				continue
			}
			mm.Elements[i].selected = true
		}
	}
	if tria {
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != Triangle3 {
				continue
			}
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectScreen(from, to [2]int32) {
	AddInfo("Model is not implement SelectScreen: %v %v", from, to)
}

func (mm *Model) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	defer mm.DeselectAll() // deselect
	if distance == 0 {
		// split by begin/end point
		// do nothing
		return
	}
	if len(lines) == 0 {
		return
	}
	// TODO single change per time Lock/Unlock
	// TODO unique lines list
	// TODO concurrency split
	cs := mm.Coords
	for _, il := range lines {
		// TODO check is line ignored
		if len(mm.Elements) <= int(il) {
			continue
		}
		if mm.Elements[il].ElementType != Line2 {
			continue
		}
		el := mm.Elements[il]
		length := math.Sqrt(
			pow.E2(cs[el.Indexes[0]].X-cs[el.Indexes[1]].X) +
				pow.E2(cs[el.Indexes[0]].Y-cs[el.Indexes[1]].Y) +
				pow.E2(cs[el.Indexes[0]].Z-cs[el.Indexes[1]].Z))
		if length < distanceError {
			// do nothing
			continue
		}
		// split point on line corner
		if atBegin && math.Abs(length-distance) < distanceError {
			// point at end point
			continue
		}
		if !atBegin && math.Abs(length+distance) < distanceError {
			// point at begin point
			continue
		}
		// change !atBegin to atBegin
		if !atBegin {
			mm.SplitLinesByDistance([]uint{il}, length-distance, !atBegin)
			continue
		}
		// split point inside line
		b := cs[el.Indexes[0]] // begin point
		e := cs[el.Indexes[1]] // end point
		proportional := distance / length
		// add new point
		id := mm.AddNode(
			b.X+(e.X-b.X)*proportional,
			b.Y+(e.Y-b.Y)*proportional,
			b.Z+(e.Z-b.Z)*proportional,
		)
		if 0 < distance && distance < length {
			// add new line only if split point inside line
			mm.AddLineByNodeNumber(id, uint(el.Indexes[1]))
			mm.Elements[il].Indexes[1] = int(id)
			continue
		}
		// split point outside line
		// split point near end line point
		// do nothing
	}
}

func (mm *Model) SplitLinesByRatio(lines []uint, proportional float64, atBegin bool) {
	defer mm.DeselectAll() // deselect
	if proportional == 0 || proportional == 1 {
		return
	}
	if len(lines) == 0 {
		return
	}
	// TODO concurrency split
	cs := mm.Coords
	for _, il := range lines {
		if len(mm.Elements) <= int(il) {
			continue
		}
		if mm.Elements[il].ElementType != Line2 {
			continue
		}
		el := mm.Elements[il]
		length := math.Sqrt(
			pow.E2(cs[el.Indexes[0]].X-cs[el.Indexes[1]].X) +
				pow.E2(cs[el.Indexes[0]].Y-cs[el.Indexes[1]].Y) +
				pow.E2(cs[el.Indexes[0]].Z-cs[el.Indexes[1]].Z))
		if length < distanceError {
			// do nothing
			continue
		}
		mm.SplitLinesByDistance([]uint{il}, proportional*length, atBegin)
	}
}

func (mm *Model) SplitLinesByEqualParts(lines []uint, parts uint) {
	defer mm.DeselectAll() // deselect
	if parts < 2 {
		return
	}
	if len(lines) == 0 {
		return
	}
	cs := mm.Coords
	for _, il := range lines {
		if len(mm.Elements) <= int(il) {
			continue
		}
		if mm.Elements[il].ElementType != Line2 {
			continue
		}
		el := mm.Elements[il]
		length := math.Sqrt(
			pow.E2(cs[el.Indexes[0]].X-cs[el.Indexes[1]].X) +
				pow.E2(cs[el.Indexes[0]].Y-cs[el.Indexes[1]].Y) +
				pow.E2(cs[el.Indexes[0]].Z-cs[el.Indexes[1]].Z))
		if length < distanceError {
			// do nothing
			continue
		}
		var ids []uint
		for p := uint(0); p < parts-1; p++ {
			proportional := float64(p+1) / float64(parts)
			b := cs[el.Indexes[0]] // begin point
			e := cs[el.Indexes[1]] // end point
			id := mm.AddNode(
				b.X+(e.X-b.X)*proportional,
				b.Y+(e.Y-b.Y)*proportional,
				b.Z+(e.Z-b.Z)*proportional,
			)
			ids = append(ids, id)
		}
		for i := range ids {
			if i == 0 {
				mm.AddLineByNodeNumber(uint(el.Indexes[0]), ids[0])
				continue
			}
			mm.AddLineByNodeNumber(ids[i-1], ids[i])
		}
		mm.AddLineByNodeNumber(ids[len(ids)-1], uint(el.Indexes[1]))
		el.Indexes[1] = int(ids[0])
	}
}

//func (mm *Model) RemoveEmptyNodes() {
// remove Removed Coordinates
// find Coordinates without elements
// remove Removed Elements
// }

func (mm *Model) MergeNodes(minDistance float64) {
	if minDistance <= 0.0 {
		return
	}
	if minDistance == 0.0 {
		minDistance = distanceError
	}

	type link struct {
		less, more int
	}

	in := make(chan link)
	re := make(chan link)

	compare := func(less, more int) {
		// SQRT(DX^2+DY^2+DZ^2) < D
		//      DX^2+DY^2+DZ^2  < D^2
		// specific cases:
		//	    DX^2            < D^2, DY=0, DZ=0
		if more <= less {
			return
		}
		if mm.Coords[less].Removed {
			return
		}
		if mm.Coords[more].Removed {
			return
		}
		dX := mm.Coords[more].X - mm.Coords[less].X
		if distanceError < math.Abs(dX) {
			return
		}
		dY := mm.Coords[more].Y - mm.Coords[less].Y
		if distanceError < math.Abs(dY) {
			return
		}
		dZ := mm.Coords[more].Z - mm.Coords[less].Z
		if distanceError < math.Abs(dZ) {
			return
		}
		distanceSquare := pow.E2(dX) + pow.E2(dY) + pow.E2(dZ)
		if math.Sqrt(distanceSquare) < minDistance {
			// Coordinates are same
			re <- link{less: less, more: more}
		}
	}

	size := runtime.NumCPU()
	if size < 1 {
		size = 1
	}
	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			for l := range in {
				compare(l.less, l.more)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(re)
	}()

	done := make(chan bool)
	go func() {
		for l := range re {
			for i := range mm.Elements {
				for j := range mm.Elements[i].Indexes {
					if mm.Elements[i].Indexes[j] == l.less {
						mm.Elements[i].Indexes[j] = l.more
					}
				}
			}
			mm.Coords[l.more].Removed = true
		}
		done <- true
	}()

	for i := range mm.Coords {
		for j := range mm.Coords {
			in <- link{less: i, more: j}
		}
	}
	close(in)

	<-done
	close(done)
	// TODO loads merge
}

func (mm *Model) Intersection(nodes, elements []uint) {
	// TODO Intersection Coordinate-Coordinate
	// TODO Intersection Coordinate-Line2
	// TODO Intersection Coordinate-Triangle3
	// TODO Intersection Line2     -Triangle3
	//
}

func (mm *Model) SplitTri3To3Tri3(tris []uint) {
	defer mm.DeselectAll() // deselect
	if len(tris) == 0 {
		return
	}
	const one3 = 1.0 / 3.0
	for _, it := range tris {
		if len(mm.Elements) <= int(it) {
			continue
		}
		el := mm.Elements[it]
		if el.ElementType != Triangle3 {
			continue
		}
		ns := []Coordinate{
			mm.Coords[el.Indexes[0]],
			mm.Coords[el.Indexes[1]],
			mm.Coords[el.Indexes[2]],
		}
		id := mm.AddNode(
			one3*ns[0].X+one3*ns[1].X+one3*ns[2].X,
			one3*ns[0].Y+one3*ns[1].Y+one3*ns[2].Y,
			one3*ns[0].Z+one3*ns[1].Z+one3*ns[2].Z,
		)
		// TODO loads on all elements
		mm.AddTriangle3ByNodeNumber(uint(el.Indexes[1]), uint(el.Indexes[2]), id)
		mm.AddTriangle3ByNodeNumber(uint(el.Indexes[2]), uint(el.Indexes[0]), id)
		el.Indexes = []int{el.Indexes[0], el.Indexes[1], int(id)}
	}
}

func (mm *Model) Hide(nodes, elements []uint) {
	for _, p := range nodes {
		mm.Coords[p].hided = true
	}
	for _, p := range elements {
		mm.Elements[p].hided = true
	}
	for i := range mm.Elements {
		el := mm.Elements[i]
		if el.hided {
			continue
		}
		for k := range el.Indexes {
			mm.Coords[el.Indexes[k]].hided = false
		}
	}
}

func (mm *Model) UnhideAll() {
	for i := range mm.Coords {
		mm.Coords[i].hided = false
	}
	for i := range mm.Elements {
		mm.Elements[i].hided = false
	}
}

func (mm *Model) MoveCopyDistance(nodes, elements []uint, coords [3]float64,
	intermediantParts uint,
	copy, addLines, addTri bool) {
	defer mm.DeselectAll() // deselect
	if distance := math.Sqrt(pow.E2(coords[0]) + pow.E2(coords[1]) + pow.E2(coords[2])); distance < distanceError {
		return
	}
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	nodes = uniqUint(nodes)
	elements = uniqUint(elements)
	if len(nodes) == 0 && len(elements) == 0 {
		return
	}
	if !copy { // move
		for _, id := range nodes {
			mm.Coords[id].X += coords[0]
			mm.Coords[id].Y += coords[1]
			mm.Coords[id].Z += coords[2]
		}
		return
	}
	// add nodes
	newNodes := make([][]uint, len(mm.Coords))
	for _, p := range nodes {
		for i := uint(0); i <= intermediantParts; i++ {
			factor := float64(i+1) / float64(intermediantParts+1)
			if i == intermediantParts {
				factor = 1.0
			}
			id := mm.AddNode(
				mm.Coords[p].X+coords[0]*factor,
				mm.Coords[p].Y+coords[1]*factor,
				mm.Coords[p].Z+coords[2]*factor,
			)
			newNodes[p] = append(newNodes[p], id)
		}
	}
	// add intermediant lines
	if addLines {
		for i := range newNodes {
			for j, p := range newNodes[i] {
				if j == 0 {
					mm.AddLineByNodeNumber(uint(i), p)
					continue
				}
				mm.AddLineByNodeNumber(newNodes[i][j-1], p)
			}
		}
	}
	// add elements
	for _, p := range elements {
		el := mm.Elements[p]
		switch el.ElementType {
		case ElRemove:
			// do nothing
		case Line2:
			for i := uint(0); i <= intermediantParts; i++ {
				mm.AddLineByNodeNumber(
					newNodes[el.Indexes[0]][i],
					newNodes[el.Indexes[1]][i],
				)
			}
		case Triangle3:
			for i := uint(0); i <= intermediantParts; i++ {
				mm.AddTriangle3ByNodeNumber(
					newNodes[el.Indexes[0]][i],
					newNodes[el.Indexes[1]][i],
					newNodes[el.Indexes[2]][i],
				)
			}
		default:
			// TODO:
			panic(fmt.Errorf("add implementation: %v", el))
		}
	}
	// add intermediant triangles for Line2
	if addTri {
		for _, p := range elements {
			el := mm.Elements[p]
			if el.ElementType != Line2 {
				continue
			}
			//  before0-------------->after0	//
			//	|                     |     	//
			//	|                     |         //
			//	before1-------------->after1	//
			for i := uint(0); i <= intermediantParts; i++ {
				var before [2]uint
				if i == 0 {
					before[0] = uint(el.Indexes[0])
					before[1] = uint(el.Indexes[1])
				} else {
					before[0] = newNodes[el.Indexes[0]][i-1]
					before[1] = newNodes[el.Indexes[1]][i-1]
				}
				after := [2]uint{
					newNodes[el.Indexes[0]][i],
					newNodes[el.Indexes[1]][i],
				}
				mm.AddTriangle3ByNodeNumber(before[0], before[1], after[1])
				mm.AddTriangle3ByNodeNumber(before[0], after[1], after[0])
			}
		}
	}
	// TODO check triangles on one line
}

func (mm *Model) MoveCopyN1N2(nodes, elements []uint, from, to uint,
	intermediantParts uint,
	copy, addLines, addTri bool) {
	if len(mm.Coords) <= int(from) {
		return
	}
	if len(mm.Coords) <= int(to) {
		return
	}
	mm.MoveCopyDistance(nodes, elements, [3]float64{
		mm.Coords[to].X - mm.Coords[from].X,
		mm.Coords[to].Y - mm.Coords[from].Y,
		mm.Coords[to].Z - mm.Coords[from].Z,
	}, intermediantParts, copy, addLines, addTri)
}

func (mm *Model) StandardView(view SView) {
	AddInfo("Model not implemented StandardView: %v", view)
}

//
// Approach is not aurogenerate model, but approach is
// fast create model.
//
// Union of 2 models
//	   m1                                                     //
//	+------------+                                            //
//	|            |                                            //
//	|            |                                            //
//	+------------+                                            //
//	                                                          //
//	               m2                m3                       //
//	        +-----------+         +--------+                  //
//	        |           |         |        |                  //
//	        |           |         |        |                  //
//	        +-----------+         +--------+                  //
//	                                                          //
//	   m1     +    m2                                         //
//	+-------+----+------+                                     //
//	|xxxxxxx|xxxx|xxxxxx|                                     //
//	|xxxxxxx|xxxx|xxxxxx|                                     //
//	+-------+----+------+                                     //
//	                                                          //
//	   m1                +           m3                       //
//	+------------+                +--------+                  //
//	|xxxxxxxxxxxx|                |xxxxxxxx|                  //
//	|xxxxxxxxxxxx|                |xxxxxxxx|                  //
//	+------------+                +--------+                  //
//
// Both model must based on single mesh
// type Union struct {
// 	Models []int
// TODO:
// problem with union points.
// How to connect 2 models?
// }
//
// FilterModel
//	   m1                                                     //
//	+------------+                                            //
//	|            |                                            //
//	|            |                                            //
//	+------------+                                            //
//	                                                          //
//	          m2                                              //
//	        +----+                                            //
//	        |    |                                            //
//	        |    |                                            //
//	        +----+                                            //
//	                                                          //
//	   m1   - m2                                              //
//	+-------+----+                                            //
//	|xxxxxxx|    |                                            //
//	|xxxxxxx|    |                                            //
//	+-------+----+                                            //
//
// Both models may based on different mesh
//
// Examples of use:
//	* Erection cases
//
// type FilterModel struct {
// 	Base          int
// 	IgnoreElement []bool
// }
//
// Modificator from `base` model to `after` model after run Updater.
// Amount coordinates, elements are `after` model equal `base` model.
//
// Examples of use:
//	* Coords based on buckling imperfections.
//	* Coords based on deformation imperfections.
//	* Move/Rotate base model.
//
// type Modificator struct {
// Update from `base` model to `after` model.
// Specific code for update models.
// Examples of code:
//	// move and rotate models №0
//	Clone 0
//	Move 0,0,2000
//	Rotate 0,90,0
//	// add imperfection buckling shape №1 at 0.750 and №2 at 0.250
//	Clone 0
//	Imperfection Buckl 1 0.750 2 0.250
//	// union models 0 and 1
//	Union 0 1 TODO: ????? MERGE POINTS
// 	Updater string
// Model after run function Update().
// If `after` is not valid, then use `base` model value.
// 	After int
// }

//	Erection case:
//	"Model based on Model with hided some parts"
//
//	Buckling imperfection:
//	"Model based on Model with modifications in according to buckling shape"
//
//	Submodel:
//	"Model based on Model with hided some parts and modifications on both Models"
//
//	Building:
//	"Model create by copy of one Model with modification of elevation"
//
//	Combine:
//	"Combine 2 models with new supports/loads"
//
//	 MODEL                  MODEL                                //
//	 |                      |                                    //
//	 |          --- X ----  |                                    //
//	 |          --------->  |                                    //
//	 +-- MESH   <-------->  +-- MESH                             //
//	 |          +-------->  |                                    //
//	 |          |  +----->  |                                    //
//	 |          |  |        |                                    //
//	 +-- LOADS  |  | -----  +-- LOADS                            //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- COMB   |  | -----  +-- COMB                             //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- SUP    |  | -----  +-- SUPPORTS                         //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- BUCKLING  |        +--                                  //
//	 |             |                                             //
//	 |             |                                             //
//	 |             |                                             //
//	 +--------- FREQUENCY                                        //
//	                                                             //

var testCoverageFunc func(m Mesh)

func Run(quit <-chan struct{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%v\n%v", err, r, string(debug.Stack()))
		}
	}()
	// initialize undo chain
	var mm Undo
	mm.model = new(Model)
	// initialize tui
	tui, err := NewTui(&mm)
	if err != nil {
		return
	}
	mm.tui = tui
	// initialize opengl view
	op, err := NewOpengl(&mm)
	if err != nil {
		return
	}
	mm.op = op
	// run test function
	go func() {
		if testCoverageFunc == nil {
			return
		}
		testCoverageFunc(&mm)
	}()
	// run opengl
	go func() { op.Run() }()
	// run tui
	return tui.Run(quit)
}

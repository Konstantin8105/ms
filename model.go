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

	op  *Opengl // for 3d
	tui *Tui    // for terminal ui
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
	mm.op.updateModel = true // Update camera parameter
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
	mm.op.updateModel = true // Update camera parameter
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
	mm.op.updateModel = true // Update camera parameter
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
	mm.op.updateModel = true // Update camera parameter
	return uint(len(mm.Elements) - 1)
}

func (mm *Model) AddLeftCursor(lc LeftCursor) {
	mm.op.AddLeftCursor(lc)
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
	mm.op.ColorEdge(isColor)
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
	mm.op.SelectLeftCursor(nodes, lines, tria)
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
		if !el.selected {
			continue
		}
		if el.ElementType != Line2 {
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
		if el.hided {
			continue
		}
		if el.selected {
			continue
		}
		if el.ElementType != Line2 {
			continue
		}
		var (
			dX = math.Abs(mm.Coords[el.Indexes[1]].X - mm.Coords[el.Indexes[0]].X)
			dY = math.Abs(mm.Coords[el.Indexes[1]].Y - mm.Coords[el.Indexes[0]].Y)
			dZ = math.Abs(mm.Coords[el.Indexes[1]].Z - mm.Coords[el.Indexes[0]].Z)
		)
		if x && dY < distanceError && dZ < distanceError {
			mm.Elements[i].selected = true
		}
		if y && dX < distanceError && dZ < distanceError {
			mm.Elements[i].selected = true
		}
		if z && dX < distanceError && dY < distanceError {
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectLinesOnPlane(xoy, yoz, xoz bool) {
	if !xoy && !yoz && !xoz {
		return
	}
	for i, el := range mm.Elements {
		if el.hided {
			continue
		}
		if el.selected {
			continue
		}
		if el.ElementType != Line2 {
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
	mm.op.SelectScreen(from, to)
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

func (mm *Model) MoveCopyNodesDistance(nodes, elements []uint, coords [3]float64, copy, addLines, addTri bool) {
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
	newNodes := make([]int, len(mm.Coords))
	for _, p := range nodes {
		id := mm.AddNode(
			mm.Coords[p].X+coords[0],
			mm.Coords[p].Y+coords[1],
			mm.Coords[p].Z+coords[2],
		)
		newNodes[p] = int(id)
		if addLines {
			mm.AddLineByNodeNumber(p, id)
		}
	}
	// add elements
	for _, p := range elements {
		el := mm.Elements[p]
		switch el.ElementType {
		case ElRemove:
			// do nothing
		case Line2:
			mm.AddLineByNodeNumber(
				uint(newNodes[el.Indexes[0]]),
				uint(newNodes[el.Indexes[1]]),
			)
		case Triangle3:
			mm.AddTriangle3ByNodeNumber(
				uint(newNodes[el.Indexes[0]]),
				uint(newNodes[el.Indexes[1]]),
				uint(newNodes[el.Indexes[2]]),
			)
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
			mm.AddTriangle3ByNodeNumber(
				uint(el.Indexes[0]),
				uint(el.Indexes[1]),
				uint(newNodes[el.Indexes[1]]),
			)
			mm.AddTriangle3ByNodeNumber(
				uint(el.Indexes[0]),
				uint(newNodes[el.Indexes[1]]),
				uint(newNodes[el.Indexes[0]]),
			)
		}
	}
	// TODO check triangles on one line
}

func (mm *Model) MoveCopyNodesN1N2(nodes, elements []uint, from, to uint, copy, addLines, addTri bool) {
	if len(mm.Coords) <= int(from) {
		return
	}
	if len(mm.Coords) <= int(to) {
		return
	}
	mm.MoveCopyNodesDistance(nodes, elements, [3]float64{
		mm.Coords[to].X - mm.Coords[from].X,
		mm.Coords[to].Y - mm.Coords[from].Y,
		mm.Coords[to].Z - mm.Coords[from].Z,
	}, copy, addLines, addTri)
}

func (mm *Model) StandardView(view SView) {
	mm.op.StandardView(view)
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
	var mm Undo
	mm.model = new(Model)

	tui, err := NewTui(&mm)
	if err != nil {
		return
	}

	op, err := NewOpengl()
	if err != nil {
		return
	}

	op.ChangeModel(mm.model)
	tui.ChangeModel(mm.model)

	go func() {
		if testCoverageFunc == nil {
			return
		}
		testCoverageFunc(&mm)
	}()

	go func() {
		op.Run()
	}()

	return tui.Run(quit)
}

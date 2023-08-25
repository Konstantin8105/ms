package ms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/ms/window"
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
	gog.Point3d
	Removed bool // TODO check everywhere

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
			m.Coords[i].Point3d[0],
			m.Coords[i].Point3d[1],
			m.Coords[i].Point3d[2],
		)
		newID[i] = int(id)
	}
	for _, el := range m.Elements {
		switch el.ElementType {
		case ElRemove:
			// do nothing
		case Line2:
			mm.AddLineByNodeNumber(
				uint(newID[el.Indexes[0]]),
				uint(newID[el.Indexes[1]]),
			)
		case Triangle3:
			mm.AddTriangle3ByNodeNumber(
				uint(newID[el.Indexes[0]]),
				uint(newID[el.Indexes[1]]),
				uint(newID[el.Indexes[2]]),
			)
		default:
			panic(fmt.Errorf("not implemented %v", el))
		}
	}
}

func (mm *Model) DemoSpiral(n uint) {
	var m Model
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = n
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)
	for i := 0; i < int(levels); i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0
		{
			var c Coordinate
			c.Point3d[0] = Ri * math.Sin(angle)
			c.Point3d[1] = float64(i) * dy
			c.Point3d[2] = Ri * math.Cos(angle)
			m.Coords = append(m.Coords, c)
		}
		{
			var c Coordinate
			c.Point3d[0] = Ro * math.Sin(angle)
			c.Point3d[1] = float64(i) * dy
			c.Point3d[2] = Ro * math.Cos(angle)
			m.Coords = append(m.Coords, c)
		}
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
	var c Coordinate
	c.Point3d[0] = X
	c.Point3d[1] = Y
	c.Point3d[2] = Z
	// check is this coordinate exist?
	for i := range mm.Coords {
		if mm.Coords[i].Removed {
			continue
		}
		// fast algorithm
		if gog.SamePoints3d(mm.Coords[i].Point3d, c.Point3d) {
			return uint(i)
		}
	}
	// append
	mm.Coords = append(mm.Coords, c)
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

func (mm *Model) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint, ok bool) {
	// type convection
	ni1 := int(n1)
	ni2 := int(n2)
	ni3 := int(n3)
	if mm.Coords[ni1].Removed || mm.Coords[ni2].Removed || mm.Coords[ni3].Removed {
		AddInfo("AddTriangle3ByNodeNumber: removed coordinate")
		return
	}
	// triangle not on one line
	if gog.ZeroTriangle3d(
		mm.Coords[ni1].Point3d,
		mm.Coords[ni2].Point3d,
		mm.Coords[ni3].Point3d,
	) {
		AddInfo("AddTriangle3ByNodeNumber: ZeroTriangle3d")
		// TODO: AddInfo("Zero:\n%v\n%v\n%v",
		// TODO: 	mm.Coords[ni1].Point3d,
		// TODO: 	mm.Coords[ni2].Point3d,
		// TODO: 	mm.Coords[ni3].Point3d,
		// TODO: )
		return
	}
	// check is this coordinate exist?
	for i, el := range mm.Elements {
		if el.ElementType != Triangle3 {
			continue
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni2 && el.Indexes[2] == ni3 {
			return uint(i), true
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni3 && el.Indexes[2] == ni1 {
			return uint(i), true
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni1 && el.Indexes[2] == ni2 {
			return uint(i), true
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni2 && el.Indexes[2] == ni1 {
			return uint(i), true
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni1 && el.Indexes[2] == ni3 {
			return uint(i), true
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni3 && el.Indexes[2] == ni2 {
			return uint(i), true
		}
		// TODO is 3 points on 1 line
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Triangle3,
		Indexes:     []int{ni1, ni2, ni3},
	})
	return uint(len(mm.Elements) - 1), true
}

func (mm *Model) AddLeftCursor(lc LeftCursor) {
	AddInfo("Model not implemented AddLeftCursor: %v", lc)
}

func (mm *Model) GetCoordByID(id uint) (c gog.Point3d, ok bool) {
	if len(mm.Coords) <= int(id) {
		return
	}
	if mm.Coords[id].Removed {
		return
	}
	return mm.Coords[id].Point3d, true
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

func (mm *Model) RemoveSameCoordinates() {
	for i := range mm.Coords {
		if mm.Coords[i].Removed {
			continue
		}
		for j := range mm.Coords {
			if mm.Coords[j].Removed {
				continue
			}
			if i <= j {
				continue
			}
			if !gog.SamePoints3d(
				mm.Coords[i].Point3d,
				mm.Coords[j].Point3d,
			) {
				continue
			}
			// fix coordinate index in elements
			from, to := j, i
			for k, el := range mm.Elements {
				if el.ElementType == ElRemove {
					continue
				}
				for g := range el.Indexes {
					if el.Indexes[g] == from {
						mm.Elements[k].Indexes[g] = to
					}
				}
			}
			// remove coordinate
			mm.Coords[j].Removed = true
		}
	}
}

func (mm *Model) RemoveNodesWithoutElements() {
	connect := make([]bool, len(mm.Coords))
	for _, el := range mm.Elements {
		if el.ElementType == ElRemove {
			continue
		}
		for i := range el.Indexes {
			connect[el.Indexes[i]] = true
		}
	}
	for i := range connect {
		if connect[i] {
			continue
		}
		mm.Coords[i].Removed = true
	}
}

func (mm *Model) RemoveZeroLines() {
	for i, el := range mm.Elements {
		if el.ElementType != Line2 {
			continue
		}
		if gog.ZeroLine3d(
			mm.Coords[el.Indexes[0]].Point3d,
			mm.Coords[el.Indexes[1]].Point3d,
		) {
			mm.Elements[i].ElementType = ElRemove
		}
	}
}

func (mm *Model) RemoveZeroTriangles() {
	for i, el := range mm.Elements {
		if el.ElementType != Triangle3 {
			continue
		}
		if gog.ZeroTriangle3d(
			mm.Coords[el.Indexes[0]].Point3d,
			mm.Coords[el.Indexes[1]].Point3d,
			mm.Coords[el.Indexes[2]].Point3d,
		) {
			mm.Elements[i].ElementType = ElRemove
		}
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

func (mm *Model) GetSelectNodes(single bool) (ids []uint) {
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

func (mm *Model) GetSelectLines(single bool) (ids []uint) {
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

func (mm *Model) GetSelectTriangles(single bool) (ids []uint) {
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

func (mm *Model) GetSelectElements(single bool) (ids []uint) {
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
			dx = math.Abs(mm.Coords[el.Indexes[1]].Point3d[0] -
				mm.Coords[el.Indexes[0]].Point3d[0])
			dy = math.Abs(mm.Coords[el.Indexes[1]].Point3d[1] -
				mm.Coords[el.Indexes[0]].Point3d[1])
			dz = math.Abs(mm.Coords[el.Indexes[1]].Point3d[2] -
				mm.Coords[el.Indexes[0]].Point3d[2])
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
			dX = math.Abs(mm.Coords[el.Indexes[1]].Point3d[0] -
				mm.Coords[el.Indexes[0]].Point3d[0])
			dY = math.Abs(mm.Coords[el.Indexes[1]].Point3d[1] -
				mm.Coords[el.Indexes[0]].Point3d[1])
			dZ = math.Abs(mm.Coords[el.Indexes[1]].Point3d[2] -
				mm.Coords[el.Indexes[0]].Point3d[2])
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
		r.dx = mm.Coords[el.Indexes[0]].Point3d[0] - mm.Coords[el.Indexes[1]].Point3d[0]
		r.dy = mm.Coords[el.Indexes[0]].Point3d[1] - mm.Coords[el.Indexes[1]].Point3d[1]
		r.dz = mm.Coords[el.Indexes[0]].Point3d[2] - mm.Coords[el.Indexes[1]].Point3d[2]
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
		L := gog.Distance3d(
			mm.Coords[el.Indexes[0]].Point3d,
			mm.Coords[el.Indexes[1]].Point3d,
		)
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

		ci0 := mm.Coords[el.Indexes[0]]
		ci1 := mm.Coords[el.Indexes[1]]
		ccn := mm.Coords[node]

		switch axe {
		case DirX:
			r0 = math.Sqrt(pow.E2(ci0.Point3d[2]-ccn.Point3d[2]) +
				pow.E2(ci0.Point3d[1]-ccn.Point3d[1]))
			r1 = math.Sqrt(pow.E2(ci1.Point3d[2]-ccn.Point3d[2]) +
				pow.E2(ci1.Point3d[1]-ccn.Point3d[1]))
			a0 = math.Atan2(ci0.Point3d[2]-ccn.Point3d[2],
				ci0.Point3d[1]-ccn.Point3d[1])
			a1 = math.Atan2(ci1.Point3d[2]-ccn.Point3d[2],
				ci1.Point3d[1]-ccn.Point3d[1])
		case DirY:
			r0 = math.Sqrt(pow.E2(ci0.Point3d[0]-ccn.Point3d[0]) +
				pow.E2(ci0.Point3d[2]-ccn.Point3d[2]))
			r1 = math.Sqrt(pow.E2(ci1.Point3d[0]-ccn.Point3d[0]) +
				pow.E2(ci1.Point3d[2]-ccn.Point3d[2]))
			a0 = math.Atan2(ci0.Point3d[0]-ccn.Point3d[0],
				ci0.Point3d[2]-ccn.Point3d[2])
			a1 = math.Atan2(ci1.Point3d[0]-ccn.Point3d[0],
				ci1.Point3d[2]-ccn.Point3d[2])
		case DirZ:
			r0 = math.Sqrt(pow.E2(ci0.Point3d[0]-ccn.Point3d[0]) +
				pow.E2(ci0.Point3d[1]-ccn.Point3d[1]))
			r1 = math.Sqrt(pow.E2(ci1.Point3d[0]-ccn.Point3d[0]) +
				pow.E2(ci1.Point3d[1]-ccn.Point3d[1]))
			a0 = math.Atan2(ci0.Point3d[0]-ccn.Point3d[0],
				ci0.Point3d[1]-ccn.Point3d[1])
			a1 = math.Atan2(ci1.Point3d[0]-ccn.Point3d[0],
				ci1.Point3d[1]-ccn.Point3d[1])
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
		length := gog.Distance3d(
			cs[el.Indexes[0]].Point3d,
			cs[el.Indexes[1]].Point3d,
		)
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
			b.Point3d[0]+(e.Point3d[0]-b.Point3d[0])*proportional,
			b.Point3d[1]+(e.Point3d[1]-b.Point3d[1])*proportional,
			b.Point3d[2]+(e.Point3d[2]-b.Point3d[2])*proportional,
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
		length := gog.Distance3d(
			cs[el.Indexes[0]].Point3d,
			cs[el.Indexes[1]].Point3d,
		)
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
		length := gog.Distance3d(
			cs[el.Indexes[0]].Point3d,
			cs[el.Indexes[1]].Point3d,
		)
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
				b.Point3d[0]+(e.Point3d[0]-b.Point3d[0])*proportional,
				b.Point3d[1]+(e.Point3d[1]-b.Point3d[1])*proportional,
				b.Point3d[2]+(e.Point3d[2]-b.Point3d[2])*proportional,
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
		if gog.SamePoints3d(mm.Coords[more].Point3d, mm.Coords[less].Point3d) {
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

func (mm *Model) MergeLines(lines []uint) {
	// 2 lines merged in case:
	//	* common point of 2 that lines
	//	* common point not removed at the end
	lines = uniqUint(lines)

	counter := 1 // default value
	for iter := 0; 0 < counter; iter++ {
		counter = 0
		for i := range lines {
			eli := mm.Elements[lines[i]]
			if eli.ElementType != Line2 {
				continue
			}
			for j := range lines {
				if i <= j {
					continue
				}
				elj := mm.Elements[lines[j]]
				if elj.ElementType != Line2 {
					continue
				}
				var ok bool
				switch {
				case eli.Indexes[0] == elj.Indexes[0]:
					eli.Indexes[0] = elj.Indexes[1]
					ok = true

				case eli.Indexes[1] == elj.Indexes[0]:
					eli.Indexes[1] = elj.Indexes[1]
					ok = true

				case eli.Indexes[0] == elj.Indexes[1]:
					eli.Indexes[0] = elj.Indexes[0]
					ok = true

				case eli.Indexes[1] == elj.Indexes[1]:
					eli.Indexes[1] = elj.Indexes[0]
					ok = true
				}
				if ok {
					mm.Elements[lines[j]].ElementType = ElRemove
					counter++
					break
				}
			}
		}
		if 1000 < iter {
			AddInfo("Too many iterations in MergeLines")
			break
		}
	}
}

var IntersectionThreads = 6

func (mm *Model) Intersection(nodes, elements []uint) {
	mm.DeselectAll()
	// remove not valid coordinates and elements
	{
		var wg sync.WaitGroup
		fs := []func(){
			mm.RemoveZeroLines,
			mm.RemoveZeroTriangles,
		}
		wg.Add(len(fs))
		for i := range fs {
			go func(i int) {
				fs[i]()
				wg.Done()
			}(i)
		}
		wg.Wait()
	}
	// remove not valid coordinates
	mm.RemoveSameCoordinates()
	// remove removed nodes, elements
	{
		var nn []uint
		for _, p := range nodes {
			if len(mm.Coords) <= int(p) {
				continue
			}
			if mm.Coords[p].Removed {
				continue
			}
			nn = append(nn, p)
		}
		nodes = nn
	}
	{
		var ne []uint
		for _, p := range elements {
			if len(mm.Elements) <= int(p) {
				continue
			}
			if mm.Elements[p].ElementType == ElRemove {
				continue
			}
			ne = append(ne, p)
		}
		elements = ne
	}
	// New intersections points

	chNewPoints := make(chan gog.Point3d)
	stop := make(chan struct{})
	var newPoints []gog.Point3d
	go func() {
		for p := range chNewPoints {
			newPoints = append(newPoints, p)
		}
		close(stop)
	}()

	var wg sync.WaitGroup

	var LLTT = [3][2]int{{0, 1}, {1, 2}, {2, 0}}
	intersectElements := []func(_, _ int){
		func(div, ost int) {
			for i0, tr0 := range elements {
				if mm.Elements[tr0].ElementType != Triangle3 {
					continue
				}
				if i0%div != ost {
					continue
				}
				// Triangle3-Triangle3
				for i1, tr1 := range elements {
					if mm.Elements[tr1].ElementType != Triangle3 {
						continue
					}
					if i0 <= i1 {
						continue
					}
					// Intersection of boxes
					outside := false
					for i := 0; i < 3; i++ {
						c0 := max(
							mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d[i],
						)
						c1 := min(
							mm.Coords[mm.Elements[tr1].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr1].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr1].Indexes[2]].Point3d[i],
						)
						if c0 < c1 {
							outside = true
							break
						}
						c2 := min(
							mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d[i],
						)
						c3 := max(
							mm.Coords[mm.Elements[tr1].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr1].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr1].Indexes[2]].Point3d[i],
						)
						if c3 < c2 {
							outside = true
							break
						}
					}
					if outside {
						continue
					}

					if intersect, pi := gog.TriangleTriangle3d(
						// coordinates triangle 0
						mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d,
						mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d,
						// coordinates triangle 1
						mm.Coords[mm.Elements[tr1].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[tr1].Indexes[1]].Point3d,
						mm.Coords[mm.Elements[tr1].Indexes[2]].Point3d,
					); intersect {
						for _, p := range pi {
							chNewPoints <- p
						}
					}
					// Triangle edges
					for _, v0 := range LLTT {
						for _, v1 := range LLTT {
							var (
								a0 = mm.Coords[mm.Elements[tr0].Indexes[v0[0]]].Point3d
								a1 = mm.Coords[mm.Elements[tr0].Indexes[v0[1]]].Point3d
								b0 = mm.Coords[mm.Elements[tr1].Indexes[v1[0]]].Point3d
								b1 = mm.Coords[mm.Elements[tr1].Indexes[v1[1]]].Point3d
							)
							rA, rB, intersect := gog.LineLine3d(a0, a1, b0, b1)
							if !intersect {
								continue
							}
							if 0 < rA && rA < 1 && 0 < rB && rB < 1 {
								chNewPoints <- gog.PointLineRatio3d(a0, a1, rA)
								chNewPoints <- gog.PointLineRatio3d(b0, b1, rB)
							}
						}
					}
				}
			}
		},
		func(div, ost int) {
			for i0, tr0 := range elements {
				if mm.Elements[tr0].ElementType != Triangle3 {
					continue
				}
				if i0%div != ost {
					continue
				}
				// Line2-Triangle3
				for _, li1 := range elements {
					if mm.Elements[li1].ElementType != Line2 {
						continue
					}
					// Intersection of boxes
					outside := false
					for i := 0; i < 3; i++ {
						c0 := max(
							mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d[i],
						)
						c1 := min(
							mm.Coords[mm.Elements[li1].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[li1].Indexes[1]].Point3d[i],
						)
						if c0 < c1 {
							outside = true
							break
						}
						c2 := min(
							mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d[i],
							mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d[i],
						)
						c3 := max(
							mm.Coords[mm.Elements[li1].Indexes[0]].Point3d[i],
							mm.Coords[mm.Elements[li1].Indexes[1]].Point3d[i],
						)
						if c3 < c2 {
							outside = true
							break
						}
					}
					if outside {
						continue
					}
					// Intersection
					for _, f := range [...]func(
						gog.Point3d, gog.Point3d, gog.Point3d, gog.Point3d, gog.Point3d,
					) (bool, []gog.Point3d){
						gog.LineTriangle3dI1,
						gog.LineTriangle3dI2,
					} {
						l0 := mm.Coords[mm.Elements[li1].Indexes[0]].Point3d
						l1 := mm.Coords[mm.Elements[li1].Indexes[1]].Point3d

						if intersect, pi := f(
							// Line2
							l0, l1,
							// Triangle3
							mm.Coords[mm.Elements[tr0].Indexes[0]].Point3d,
							mm.Coords[mm.Elements[tr0].Indexes[1]].Point3d,
							mm.Coords[mm.Elements[tr0].Indexes[2]].Point3d,
						); intersect {
							for _, p := range pi {
								chNewPoints <- p
							}
						}
					}
				}
			}
		},
		func(div, ost int) {
			// Line2-Line2
			for i0, li0 := range elements {
				if mm.Elements[li0].ElementType != Line2 {
					continue
				}
				if i0%div != ost {
					continue
				}
				for i1, li1 := range elements {
					if mm.Elements[li1].ElementType != Line2 {
						continue
					}
					if i0 <= i1 {
						continue
					}
					// Intersection of boxes
					if gog.SamePoints3d(
						mm.Coords[mm.Elements[li0].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[li1].Indexes[0]].Point3d,
					) {
						continue
					}
					if gog.SamePoints3d(
						mm.Coords[mm.Elements[li0].Indexes[1]].Point3d,
						mm.Coords[mm.Elements[li1].Indexes[0]].Point3d,
					) {
						continue
					}
					if gog.SamePoints3d(
						mm.Coords[mm.Elements[li0].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[li1].Indexes[1]].Point3d,
					) {
						continue
					}
					if gog.SamePoints3d(
						mm.Coords[mm.Elements[li0].Indexes[1]].Point3d,
						mm.Coords[mm.Elements[li1].Indexes[1]].Point3d,
					) {
						continue
					}
					// Intersection
					var (
						a0 = mm.Coords[mm.Elements[li0].Indexes[0]].Point3d
						a1 = mm.Coords[mm.Elements[li0].Indexes[1]].Point3d
						b0 = mm.Coords[mm.Elements[li1].Indexes[0]].Point3d
						b1 = mm.Coords[mm.Elements[li1].Indexes[1]].Point3d
					)
					if rA, rB, intersect := gog.LineLine3d(
						a0, a1,
						b0, b1,
					); intersect {
						if 0 < rA && rA < 1 && 0 < rB && rB < 1 {
							chNewPoints <- gog.PointLineRatio3d(a0, a1, rA)
							chNewPoints <- gog.PointLineRatio3d(b0, b1, rB)
						}
					}
				}
			}
		},
	}

	var size = IntersectionThreads
	if size < 1 {
		size = 1
	}
	wg.Add(len(intersectElements) * size)
	for i := range intersectElements {
		for k := 0; k < size; k++ {
			go func(i, k int) {
				intersectElements[i](size, k)
				wg.Done()
			}(i, k)
		}
	}
	wg.Wait()
	close(chNewPoints)
	_ = <-stop

	AddInfo("Intersection: find new %d points", len(newPoints))

	// fix zero coordinates
	for i := range newPoints {
		for j := range newPoints[i] {
			if math.Abs(newPoints[i][j]) < gog.Eps3D {
				newPoints[i][j] = 0.0
			}
		}
	}

	// add nodes into newPoints
	for _, pn := range nodes {
		if mm.Coords[pn].Removed {
			continue
		}
		newPoints = append(newPoints, mm.Coords[pn].Point3d)
	}

	// Coordinate-Coordinate
	// remove same points at list of new points
	nodes = nil
	for i := range newPoints {
		found := false
		for j := range newPoints {
			if i <= j {
				continue
			}
			if gog.SamePoints3d(newPoints[i], newPoints[j]) {
				found = true
			}
		}
		if found {
			continue
		}
		id := mm.AddNode(
			newPoints[i][0],
			newPoints[i][1],
			newPoints[i][2],
		)
		nodes = append(nodes, id)
	}
	AddInfo("Intersection: %d nodes", len(nodes))

	for iter := 0; ; iter++ { // TODO avoid infinite
		var newElements []uint
		for _, pe := range elements {
			for _, n := range nodes {
				// avoid Coordinate-Coordinate
				found := false
				for _, ind := range mm.Elements[pe].Indexes {
					if gog.SamePoints3d(
						mm.Coords[n].Point3d,
						mm.Coords[ind].Point3d,
					) {
						found = true
						break
					}
				}
				if found {
					continue
				}
				// intersection
				switch mm.Elements[pe].ElementType {
				case ElRemove:
					// do nothing
				case Line2:
					// Coordinate-Line2
					if !gog.PointLine3d(
						mm.Coords[n].Point3d,
						mm.Coords[mm.Elements[pe].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[pe].Indexes[1]].Point3d,
					) {
						continue
					}
					// point on line
					nl := mm.AddLineByNodeNumber(n, uint(mm.Elements[pe].Indexes[1]))
					mm.Elements[pe].Indexes[1] = int(n)
					newElements = append(newElements, nl)

				case Triangle3:
					// split point on Triangle3 edge
					ind := mm.Elements[pe].Indexes
					if gog.PointLine3d(
						mm.Coords[n].Point3d,
						mm.Coords[ind[0]].Point3d,
						mm.Coords[ind[1]].Point3d,
					) {
						nt, ok := mm.AddTriangle3ByNodeNumber(
							n, uint(ind[1]), uint(ind[2]),
						)
						// AddInfo("Intersection 0-1: %v", ok)
						if !ok {
							AddInfo("Intersection: split point on triangle edge 01 invalid")
							continue
						}
						ind[1] = int(n)
						newElements = append(newElements, nt)
						continue
					}
					if gog.PointLine3d(
						mm.Coords[n].Point3d,
						mm.Coords[ind[1]].Point3d,
						mm.Coords[ind[2]].Point3d,
					) {
						nt, ok := mm.AddTriangle3ByNodeNumber(
							n, uint(ind[0]), uint(ind[2]),
						)
						// AddInfo("Intersection 1-2: %v", ok)
						if !ok {
							AddInfo("Intersection: split point on triangle edge 12 invalid")
							continue
						}
						ind[2] = int(n)
						newElements = append(newElements, nt)
						continue
					}
					if gog.PointLine3d(
						mm.Coords[n].Point3d,
						mm.Coords[ind[2]].Point3d,
						mm.Coords[ind[0]].Point3d,
					) {
						nt, ok := mm.AddTriangle3ByNodeNumber(
							n, uint(ind[1]), uint(ind[2]),
						)
						// AddInfo("Intersection 2-0: %v", ok)
						if !ok {
							AddInfo("Intersection: split point on triangle edge 20 invalid")
							continue
						}
						ind[2] = int(n)
						newElements = append(newElements, nt)
						continue
					}

					// Coordinate-Triangle3
					if gog.PointTriangle3d(
						mm.Coords[n].Point3d,
						mm.Coords[mm.Elements[pe].Indexes[0]].Point3d,
						mm.Coords[mm.Elements[pe].Indexes[1]].Point3d,
						mm.Coords[mm.Elements[pe].Indexes[2]].Point3d,
					) {
						// point inside triangle
						t0, ok0 := mm.AddTriangle3ByNodeNumber(
							n,
							uint(mm.Elements[pe].Indexes[0]),
							uint(mm.Elements[pe].Indexes[1]),
						)
						t1, ok1 := mm.AddTriangle3ByNodeNumber(
							n,
							uint(mm.Elements[pe].Indexes[1]),
							uint(mm.Elements[pe].Indexes[2]),
						)
						// AddInfo("Intersection Coordinate-Triangle: %v %v", ok0, ok1)
						if !(ok0 && ok1) {
							AddInfo("Intersection: not valid triangles")
							continue
						}
						mm.Elements[pe].Indexes[1] = int(n)
						newElements = append(newElements, t0, t1)
						continue
					}

				default:
					panic("not implemented")
				}
			}
		}
		AddInfo("Intersection: add %d elements", len(newElements))
		if len(newElements) == 0 {
			break
		}
		elements = append(elements, newElements...)
		if 100 < iter {
			AddInfo("Intersection iterations break")
			break
		}
	}
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
			one3*ns[0].Point3d[0]+one3*ns[1].Point3d[0]+one3*ns[2].Point3d[0],
			one3*ns[0].Point3d[1]+one3*ns[1].Point3d[1]+one3*ns[2].Point3d[1],
			one3*ns[0].Point3d[2]+one3*ns[1].Point3d[2]+one3*ns[2].Point3d[2],
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

func (mm *Model) Move(nodes, elements []uint,
	basePoint [3]float64,
	path diffCoordinate) {
	defer mm.DeselectAll() // deselect
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	nodes = uniqUint(nodes)
	if len(nodes) == 0 {
		return
	}
	// moving
	for _, id := range nodes {
		from := [3]float64(mm.Coords[id].Point3d)
		move(&from, basePoint, path)
		mm.Coords[id].Point3d = from
	}
}

func move(coord *[3]float64, basePoint [3]float64, dc diffCoordinate) {
	// moving
	coord[0] += dc[0]
	coord[1] += dc[1]
	coord[2] += dc[2]
	// rotate
	{
		// around X
		point := gog.Point{
			X: coord[1], // Y
			Y: coord[2], // Z
		}
		point = gog.Rotate(
			basePoint[1], basePoint[2],
			dc[3]*radToDegree, point)
		coord[1] = point.X
		coord[2] = point.Y
	}
	{
		// around Y
		point := gog.Point{
			X: coord[0], // X
			Y: coord[2], // Z
		}
		point = gog.Rotate(
			basePoint[0], basePoint[2],
			dc[4]*radToDegree, point)
		coord[0] = point.X
		coord[2] = point.Y
	}
	{
		// around Z
		point := gog.Point{
			X: coord[0], // X
			Y: coord[1], // Y
		}
		point = gog.Rotate(
			basePoint[0], basePoint[1],
			dc[5]*radToDegree, point)
		coord[0] = point.X
		coord[1] = point.Y
	}
}

const radToDegree = math.Pi / 180.0

func (mm *Model) Copy(nodes, elements []uint,
	basePoint [3]float64,
	paths []diffCoordinate,
	addLines, addTri bool) {
	defer mm.DeselectAll() // deselect
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	nodes = uniqUint(nodes)
	elements = uniqUint(elements)
	if len(nodes) == 0 || len(elements) == 0 {
		return
	}
	if len(paths) == 0 {
		return
	}
	// create copy of model
	var cModel Model
	for _, n := range nodes { // add all points
		cModel.AddNode(
			mm.Coords[n].Point3d[0],
			mm.Coords[n].Point3d[1],
			mm.Coords[n].Point3d[2],
		)
	}
	for _, pe := range elements { // add all elements
		el := mm.Elements[pe]
		if el.ElementType == ElRemove {
			// do nothing
			continue
		}
		ids := make([]uint, len(el.Indexes))
		for i := range ids {
			id := cModel.AddNode(
				mm.Coords[el.Indexes[i]].Point3d[0],
				mm.Coords[el.Indexes[i]].Point3d[1],
				mm.Coords[el.Indexes[i]].Point3d[2],
			)
			ids[i] = id
		}
		switch el.ElementType {
		case Line2:
			cModel.AddLineByNodeNumber(ids[0], ids[1])
		case Triangle3:
			cModel.AddTriangle3ByNodeNumber(ids[0], ids[1], ids[2])
		default:
			panic(fmt.Errorf("Undefined: %v", el.ElementType))
		}
	}
	// main model
	var (
		beginCoord = make([]Coordinate, len(cModel.Coords))
		endCoord   = make([]Coordinate, len(cModel.Coords))
	)
	copy(beginCoord, cModel.Coords)
	for _, path := range paths {
		for i := range cModel.Coords {
			from := [3]float64(cModel.Coords[i].Point3d)
			move(&from, basePoint, path)
			cModel.Coords[i].Point3d = from
		}
		copy(endCoord, cModel.Coords)

		if addLines {
			var pre Model
			for i := range beginCoord {
				beginID := pre.AddNode(
					beginCoord[i].Point3d[0],
					beginCoord[i].Point3d[1],
					beginCoord[i].Point3d[2],
				)
				endID := pre.AddNode(
					endCoord[i].Point3d[0],
					endCoord[i].Point3d[1],
					endCoord[i].Point3d[2],
				)
				pre.AddLineByNodeNumber(beginID, endID)
			}
			mm.AddModel(pre)
		}
		if addTri {
			var pre Model
			for _, el := range cModel.Elements {
				if el.ElementType != Line2 {
					continue
				}
				var (
					begin0 = pre.AddNode(
						beginCoord[el.Indexes[0]].Point3d[0],
						beginCoord[el.Indexes[0]].Point3d[1],
						beginCoord[el.Indexes[0]].Point3d[2],
					)
					begin1 = pre.AddNode(
						beginCoord[el.Indexes[1]].Point3d[0],
						beginCoord[el.Indexes[1]].Point3d[1],
						beginCoord[el.Indexes[1]].Point3d[2],
					)
					end0 = pre.AddNode(
						endCoord[el.Indexes[0]].Point3d[0],
						endCoord[el.Indexes[0]].Point3d[1],
						endCoord[el.Indexes[0]].Point3d[2],
					)
					end1 = pre.AddNode(
						endCoord[el.Indexes[1]].Point3d[0],
						endCoord[el.Indexes[1]].Point3d[1],
						endCoord[el.Indexes[1]].Point3d[2],
					)
				)
				pre.AddTriangle3ByNodeNumber(begin0, begin1, end1)
				pre.AddTriangle3ByNodeNumber(end1, end0, begin0)
			}
			mm.AddModel(pre)
		}
		beginCoord, endCoord = endCoord, beginCoord

		copyBase := [3]float64{
			basePoint[0],
			basePoint[1],
			basePoint[2],
		}
		move(&copyBase, basePoint, path)

		mm.AddModel(cModel)
	}
}

func (mm *Model) Mirror(nodes, elements []uint,
	basePoint [3][3]float64,
	copy bool,
	addLines, addTri bool) {
	// TODO gog.Mirror3d for points
}

func (mm *Model) StandardView(view SView) {
	AddInfo("Model not implemented StandardView: %v", view)
}

func max(xs ...float64) (res float64) {
	if len(xs) == 0 {
		panic("not valid: zero lenght")
	}
	res = xs[0]
	for i := range xs {
		res = math.Max(res, xs[i])
	}
	return
}

func min(xs ...float64) (res float64) {
	if len(xs) == 0 {
		panic("not valid: zero lenght")
	}
	res = xs[0]
	for i := range xs {
		res = math.Min(res, xs[i])
	}
	return
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

var testCoverageFunc func(m Mesh, ch *chan func())

func Run(filename string, quit <-chan struct{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%v\n%v", err, r, string(debug.Stack()))
		}
	}()
	// initialize undo chain
	var mm Undo
	// prepare model
	if filename == "" {
		mm.model = new(Model)
	} else if strings.HasSuffix(strings.ToLower(filename), ".geo") {
		// read gmsh file
		var model Model
		model, err = Gmsh2Model(filename)
		if err != nil {
			return
		}
		mm.model = &model
	} else {
		// read native json file format
		var b []byte
		b, err = ioutil.ReadFile(filename)
		if err != nil {
			return
		}
		var model Model
		if err = json.Unmarshal(b, &model); err != nil {
			return
		}
		mm.model = &model
	}

	closedApp := false
	ch := make(chan func(), 1000)

	var ws [2]ds.Window

	tui, err := NewTui(&mm, &closedApp, &ch)
	if err != nil {
		return
	}
	tuiWindow := window.NewTui(tui)
	ws[0] = tuiWindow

	opWindow, err := NewOpengl(&mm, &ch)
	if err != nil {
		return
	}
	ws[1] = opWindow

	screen, err := ds.New("Demo", ws, &ch)
	if err != nil {
		return
	}
	// add fonts
	f, err := glsymbol.DefaultFont()
	if err != nil {
		return
	}
	tuiWindow.SetFont(f)
	opWindow.SetFont(f)

	// initialize tui
	// tui, err := NewTui(&mm)
	// if err != nil {
	// 	return
	// }
	// mm.tui = tui // TODO Why????

	// initialize opengl view
	// op, err := NewOpengl(&mm)
	// if err != nil {
	// 	return
	// }
	mm.op = opWindow
	// // run test function
	go func() {
		if testCoverageFunc == nil {
			return
		}
		testCoverageFunc(&mm, &ch)
	}()
	// // run opengl
	// go func() { op.Run() }()
	// run tui
	ch <- func() {
		screen.ChangeRatio(0.4) // TODO: add to interface
	}
	screen.Run(quit)
	closedApp = true
	time.Sleep(2 * time.Second)
	close(ch)
	return
}

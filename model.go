package ms

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
	"runtime/debug"
	"time"

	"github.com/Konstantin8105/ds"
	etree "github.com/Konstantin8105/errors"
	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/ms/groups"
	"github.com/Konstantin8105/ms/window"
	"github.com/Konstantin8105/pow"
	"github.com/Konstantin8105/vl"
)

const FileExtension = "ms"

// 3D model variables
type object3d struct {
	selected bool
	hided    bool
}

type ElType uint8 // from 0 to 255

const (
	Line2     ElType = iota + 1 // 1
	Triangle3                   // 2
	Quadr4                      // 3
	lastElement
	ElRemove = math.MaxUint8 // 255
)

func (e ElType) String() string {
	switch e {
	case Line2:
		return "Line with 2 points"
	case Triangle3:
		return "Triangle with 3 points"
	case Quadr4:
		return "Quard with 4 points"
	}
	return "Undefined type element"
}

func (e ElType) getSelect() viewState {
	switch e {
	case Line2:
		return selectLines
	case Triangle3:
		return selectTriangles
	case Quadr4:
		return selectQuadrs
	}
	panic(fmt.Errorf("undefined getSelect: %v", e))
}

// Element is typical element for FEM. Examples:
//
//	Line o======o
//	ElType : 1
//	Indexes: 2 (amount indexes of coordinates)
//
//	Triangle o======o
//	          \    /
//	           \  /
//	            o
//	ElType : 2
//	Indexes: 3 (amount indexes of coordinates)
//
//	Quadr4 o======o
//	       |      |
//	       |      |
//	       o======o
//	ElType : 3
//	Indexes: 4 (amount indexes of coordinates)
type Element struct {
	object3d
	ElementType ElType
	Indexes     []int // index of coordinate
}

var valids = []struct {
	e      ElType
	amount int
	lc     LeftCursor
}{
	{Line2, 2, AddLinesLC},
	{Triangle3, 3, AddTrianglesLC},
	{Quadr4, 4, AddQuardsLC},
	{e: ElRemove, amount: 0},
}

func (e Element) Check() error {
	index := -1
	for i := range valids {
		if e.ElementType == valids[i].e {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("undefined type: %v", e)
	}
	if len(e.Indexes) != valids[index].amount {
		return fmt.Errorf("unacceptable element: %v", e)
	}
	return nil
}

// Coordinate store coordinate of points
type Coordinate struct {
	object3d
	gog.Point3d
	Removed bool

	// TODO
	// index int    // index of Models
	// C [3]float64 // coordinates
}

func (c Coordinate) Check() error {
	if max := 1e6; max < math.Abs(c.Point3d[0]) ||
		max < math.Abs(c.Point3d[1]) ||
		max < math.Abs(c.Point3d[2]) ||
		false {
		return fmt.Errorf("Coordinate is too big")
	}
	return nil
}

// Named intermediant named structure
// type Named struct{ Name string }

// type Ignored struct{ IgnoreElements []bool }

// TODO : type MultiModel struct { Models []Model}

type Model struct {
	Elements []Element
	Coords   []Coordinate
	Groups   struct {
		Data string
		meta groups.Meta
	}
	filename string
}

func (mm Model) getPoint3d(index uint) (ps []gog.Point3d) {
	if len(mm.Elements) <= int(index) {
		logger.Printf("getPoint3d: not valid index: %d %d", len(mm.Elements), index)
		return
	}
	el := mm.Elements[index]
	if el.ElementType == ElRemove {
		logger.Printf("getPoint3d: removed element: %d", index)
		return
	}
	for _, ind := range el.Indexes {
		ps = append(ps, mm.Coords[ind].Point3d)
	}
	return
}

func (mm *Model) Check() error {
	et := etree.New("check model")
	for i, c := range mm.Coords {
		if err := c.Check(); err != nil {
			et.Add(fmt.Errorf("Coordinate: %d\n%v", i, err))
		}
	}
	for i := range mm.Coords {
		for _, v := range mm.Coords[i].Point3d {
			if mm.isValidValue(v) {
				continue
			}
			et.Add(fmt.Errorf("Coords: %d\nNot valid value", i))
		}
	}
	for i, el := range mm.Elements {
		if err := el.Check(); err != nil {
			et.Add(fmt.Errorf("Element type `%d`: %d\n%v", el.ElementType, i, err))
		}
		for _, p := range el.Indexes {
			if p < 0 {
				et.Add(fmt.Errorf("Element: %d\nCoordinate index is negative", i))
			}
			if len(mm.Coords) <= p {
				et.Add(fmt.Errorf("Element: %d\nCoordinate index is too big", i))
			}
		}
		for i := range el.Indexes {
			for j := range el.Indexes {
				if i <= j {
					continue
				}
				if el.Indexes[i] == el.Indexes[j] {
					et.Add(fmt.Errorf("Element: same indexes %v", el.Indexes))
				}
			}
		}
	}
	if et.IsError() {
		return et
	}
	// TODO check - not same coordiantes
	// TODO check - Triangle3 , Quadr4 not on one line
	// TODO check - Quadr4 on one plane
	return nil
}

// TODO
// type Metadata struct {
// TODO Group by parts
// TODO Material
// TODO Text on point
// TODO Loads
// TODO Local axes
// TODO reverse localc axes
// }

// type Part struct {
// 	Named
// 	Ignored
// }

func (mm *Model) Undo() {
	// do nothing
}
func (mm *Model) Redo() {
	// go nothing
}

// func clearPartName(name *string) {
// 	*name = strings.ReplaceAll(*name, "\n", "")
// 	*name = strings.ReplaceAll(*name, "\r", "")
// 	*name = strings.ReplaceAll(*name, "\t", "")
// }
//
// func (mm *Model) PartsName() (names []string) {
// 	names = append(names, mm.Name)
// 	for i := range mm.Parts {
// 		names = append(names, mm.Parts[i].Name)
// 	}
// 	return
// }
//
// func (mm *Model) PartPresent() (id uint) {
// 	return uint(mm.actual)
// }
//
// func (mm *Model) PartChange(id uint) {
// 	if id == 0 {
// 		mm.actual = 0
// 		return
// 	}
// 	id = id - 1 // convert to part indexes
// 	if int(id) <= len(mm.Parts) {
// 		mm.actual = int(id) + 1
// 	}
// 	// no changes
// }
//
// func (mm *Model) PartNew(name string) {
// 	clearPartName(&name)
// 	var p Part
// 	p.Name = name
// 	mm.Parts = append(mm.Parts, p)
// 	mm.actual = len(mm.Parts) // no need `-1`, because base model
// }
//
// func (mm *Model) PartRename(id uint, name string) {
// 	clearPartName(&name)
// 	if id == 0 {
// 		mm.Name = name
// 		return
// 	}
// 	if len(mm.Parts) < int(id) {
// 		return
// 	}
// 	mm.Parts[id-1].Name = name
// }

func (mm *Model) GetPresentFilename() (name string) {
	logger.Printf("GetPresentFilename")
	return mm.filename
}

func (mm *Model) Save() (err error) {
	logger.Printf("Save")
	// actions
	var bs []byte
	bs, err = groups.SaveGroup(&mm.Groups.meta)
	if err != nil {
		logger.Printf("Save: %v", err)
		return
	}
	mm.Groups.Data = string(bs)
	b, err := json.MarshalIndent(mm, "", "  ")
	if err != nil {
		logger.Printf("Save: %v", err)
		return
	}
	err = os.WriteFile(mm.filename, b, 0666)
	if err != nil {
		logger.Printf("Save: %v", err)
		return
	}
	return
}

func (mm *Model) SaveAs(filename string) (err error) {
	logger.Printf("SaveAs")
	// check
	if filename == "" {
		err = fmt.Errorf("empty filename")
		return
	}
	// actions
	name := mm.filename
	mm.filename = filename
	if err := mm.Save(); err != nil {
		mm.filename = name
		return err
	}
	return nil
}

func (mm *Model) Open(mesh groups.Mesh, filename string) (err error) {
	logger.Printf("Open")
	// check
	if filename == "" {
		err = fmt.Errorf("empty filename")
		return
	}
	if mesh == nil {
		err = fmt.Errorf("mesh is nil")
		return
	}
	// actions
	// read native json file format
	var b []byte
	b, err = ioutil.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("Open error: %v", err)
		return
	}
	var model Model
	if err = json.Unmarshal(b, &model); err != nil {
		err = fmt.Errorf("Open error: %v", err)
		return
	}
	*mm = model
	mm.filename = filename
	var gr groups.Group
	gr, err = groups.ParseGroup([]byte(mm.Groups.Data))
	if err != nil {
		err = fmt.Errorf("Open error: %v", err)
		return
	}
	if m, ok := gr.(*groups.Meta); ok {
		mm.Groups.meta = *m
		groups.FixMesh(mesh)
	} else {
		err = fmt.Errorf("Open error: is not Meta")
	}
	return
}

func (mm *Model) AddModel(m Model) {
	if err := m.Check(); err != nil {
		logger.Printf("AddModel: Model not valid\n%v", err)
		return
	}
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
		case Quadr4:
			mm.AddQuadr4ByNodeNumber(
				uint(newID[el.Indexes[0]]),
				uint(newID[el.Indexes[1]]),
				uint(newID[el.Indexes[2]]),
				uint(newID[el.Indexes[3]]),
			)
		default:
			logger.Printf("AddModel: not implemented %v", el)
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

func (mm *Model) AddNode(X, Y, Z float64) (id uint) {
	// check
	for _, p := range []*float64{&X, &Y, &Z} {
		if math.Abs(*p) < gog.Eps3D {
			*p = 0
		}
		if !mm.isValidValue(*p) {
			logger.Printf("AddNode: not valid value: %v", *p)
			return
		}
	}
	// actions
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
	// check
	if s := []uint{n1, n2}; !mm.isValidNodeId(s) {
		logger.Printf("AddLineByNodeNumber: not valid node id: %v", s)
		return
	}
	if gog.ZeroLine3d(
		mm.Coords[n1].Point3d,
		mm.Coords[n2].Point3d,
	) {
		logger.Printf("AddLineByNodeNumber: ZeroLine3d")
		return
	}
	if n1 == n2 {
		logger.Printf("AddLineByNodeNumber: same indexes: %d", n1)
		return
	}
	// actions
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
	// check
	if s := []uint{n1, n2, n3}; !mm.isValidNodeId(s) {
		logger.Printf("AddTriangle3ByNodeNumber: not valid node id: %v", s)
		return
	}
	// actions
	// triangle not on one line
	if gog.ZeroTriangle3d(
		mm.Coords[int(n1)].Point3d,
		mm.Coords[int(n2)].Point3d,
		mm.Coords[int(n3)].Point3d,
	) {
		logger.Printf("AddTriangle3ByNodeNumber: ZeroTriangle3d")
		return
	}
	// check that triangle is not exist
	nis := [][3]uint{
		{n1, n2, n3}, {n2, n3, n1}, {n3, n1, n2},
		{n3, n2, n1}, {n2, n1, n3}, {n1, n3, n2},
	}
	for i, el := range mm.Elements {
		if el.ElementType != Triangle3 {
			continue
		}
		for _, ni := range nis {
			if el.Indexes[0] == int(ni[0]) &&
				el.Indexes[1] == int(ni[1]) &&
				el.Indexes[2] == int(ni[2]) {
				return uint(i), true
			}
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Triangle3,
		Indexes:     []int{int(n1), int(n2), int(n3)},
	})
	return uint(len(mm.Elements) - 1), true
}

func (mm *Model) AddQuadr4ByNodeNumber(n1, n2, n3, n4 uint) (id uint, ok bool) {
	// check
	if s := []uint{n1, n2, n3, n4}; !mm.isValidNodeId(s) {
		logger.Printf("AddQuadr4ByNodeNumber: not valid node id: %v", s)
		return
	}
	// actions
	// check triangle not on one line
	if gog.ZeroTriangle3d(
		mm.Coords[int(n1)].Point3d,
		mm.Coords[int(n2)].Point3d,
		mm.Coords[int(n3)].Point3d,
	) || gog.ZeroTriangle3d(
		mm.Coords[int(n2)].Point3d,
		mm.Coords[int(n3)].Point3d,
		mm.Coords[int(n4)].Point3d,
	) {
		logger.Printf("AddQuadr4ByNodeNumber: ZeroTriangle3d")
		return
	}
	// check all points on one plane
	{
		A, B, C, D := gog.Plane(
			mm.Coords[int(n1)].Point3d,
			mm.Coords[int(n2)].Point3d,
			mm.Coords[int(n3)].Point3d,
		)
		if !gog.PointOnPlane3d(
			A, B, C, D,
			mm.Coords[int(n4)].Point3d,
		) {
			logger.Printf("AddQuadr4ByNodeNumber: not on one plane")
			return
		}
	}
	// check that triangle is not exist
	nis := [][4]uint{
		{n1, n2, n3, n4},
		{n2, n3, n4, n1},
		{n3, n4, n1, n2},
		{n4, n1, n2, n3},

		{n1, n2, n4, n3},
		{n2, n4, n3, n1},
		{n4, n3, n1, n2},
		{n3, n1, n2, n4},

		{n1, n3, n2, n4},
		{n3, n2, n4, n1},
		{n2, n4, n1, n3},
		{n4, n1, n3, n2},

		{n1, n4, n3, n2},
		{n4, n3, n2, n1},
		{n3, n2, n1, n4},
		{n2, n1, n4, n3},
	}
	for i, el := range mm.Elements {
		if el.ElementType != Quadr4 {
			continue
		}
		for _, ni := range nis {
			if el.Indexes[0] == int(ni[0]) &&
				el.Indexes[1] == int(ni[1]) &&
				el.Indexes[2] == int(ni[2]) &&
				el.Indexes[3] == int(ni[3]) {
				return uint(i), true
			}
		}
	}
	// avoid insection with it-self
	{
		_, _, intersection1 := gog.LineLine3d(
			mm.Coords[int(n1)].Point3d,
			mm.Coords[int(n2)].Point3d,
			mm.Coords[int(n3)].Point3d,
			mm.Coords[int(n4)].Point3d,
		)
		_, _, intersection2 := gog.LineLine3d(
			mm.Coords[int(n2)].Point3d,
			mm.Coords[int(n3)].Point3d,
			mm.Coords[int(n4)].Point3d,
			mm.Coords[int(n1)].Point3d,
		)
		if intersection1 || intersection2 {
			n3, n4 = n4, n3 // swap nodes
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Quadr4,
		Indexes:     []int{int(n1), int(n2), int(n3), int(n4)},
	})
	return uint(len(mm.Elements) - 1), true
}

func (mm *Model) AddConvexLines(nodes, elements []uint) {
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("AddConvexLines: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("AddConvexLines: not valid elements id: %v", s)
		return
	}
	// actions
	// check
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll()
	// add nodes from elements
	for _, e := range elements {
		el := mm.Elements[e]
		if el.ElementType == ElRemove {
			continue
		}
		for _, n := range el.Indexes {
			nodes = append(nodes, uint(n))
		}
	}
	nodes = uniqUint(nodes)
	if len(nodes) < 2 {
		logger.Printf("AddConvexLines: not enought nodes")
		return
	}
	// calculate average plane
	var ps []gog.Point3d
	for _, n := range nodes {
		ps = append(ps, mm.Coords[n].Point3d)
	}
	min, max := gog.BorderPoints(ps...)
	dm := gog.Point3d{
		math.Abs(max[0] - min[0]),
		math.Abs(max[1] - min[1]),
		math.Abs(max[2] - min[2]),
	}
	p2 := make([]gog.Point, len(ps))
	if dm[0] <= dm[1] && dm[0] <= dm[2] {
		// Y-Z plane
		for i := range ps {
			p2[i].X = ps[i][1]
			p2[i].Y = ps[i][2]
		}
	} else if dm[1] <= dm[0] && dm[1] <= dm[2] {
		// X-Z plane
		for i := range ps {
			p2[i].X = ps[i][0]
			p2[i].Y = ps[i][2]
		}
	} else if dm[2] <= dm[0] && dm[2] <= dm[1] {
		// X-Y plane
		for i := range ps {
			p2[i].X = ps[i][0]
			p2[i].Y = ps[i][1]
		}
	} else {
		return
	}
	// convex hull
	chain, _ := gog.ConvexHull(p2, false)
	for i := range chain {
		chain[i] = int(nodes[chain[i]])
	}
	// add convex lines
	for i := range chain {
		var b, f int
		if i == 0 {
			b = len(chain) - 1
		} else {
			b = i - 1
		}
		f = i
		mm.AddLineByNodeNumber(uint(chain[b]), uint(chain[f]))
	}
}

func (mm *Model) AddLeftCursor(lc LeftCursor) {
	// do nothing
}

func (mm *Model) GetCoordByID(id uint) (c gog.Point3d, ok bool) {
	// check
	if s := []uint{id}; !mm.isValidNodeId(s) {
		logger.Printf("GetCoordByID: not valid node id: %v", s)
		return
	}
	// actions
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
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	// check
	if s := nodes; !mm.isValidNodeId(s) {
		logger.Printf("Remove: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(elements, nil) {
		logger.Printf("Remove: not valid element id: %v", s)
		return
	}
	// actions
	// it is part/model
	// do not remove nodes in ignore list
	ignore := make([]bool, len(nodes))
	// for ind, p := range nodes {
	// 	if mm.Coords[int(p)].hided || mm.Coords[int(p)].Removed {
	// 		ignore[ind] = true
	// 		continue
	// 	}
	// 	for i := range mm.Elements {
	// 		if mm.Elements[i].ElementType == ElRemove {
	// 			continue
	// 		}
	// 		if !mm.IsIgnore(uint(i)) {
	// 			continue
	// 		}
	// 		// ignored coordinate on ignored elements
	// 		for k := range mm.Elements[i].Indexes {
	// 			if mm.Elements[i].Indexes[k] == int(p) {
	// 				ignore[ind] = true
	// 			}
	// 		}
	// 	}
	// }
	// remove
	for ind, p := range nodes {
		if ignore[ind] {
			continue
		}
		// removing coordinates
		mm.Coords[p].Removed = true
		// remove elements with coordinate
		for i := range mm.Elements {
			// if mm.IsIgnore(uint(i)) {
			// 	continue
			// }
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

// TODO remove
func (mm *Model) RemoveSameCoordinates() {
	mm.MergeNodes(0.0)
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

// func (mm *Model) IsIgnore(elID uint) bool {
// 	if 0 < mm.actual && int(elID) < len(mm.Parts[mm.actual-1].IgnoreElements) {
// 		// it is part
// 		return mm.Parts[mm.actual-1].IgnoreElements[int(elID)]
// 	}
// 	if int(elID) < len(mm.IgnoreElements) {
// 		return mm.IgnoreElements[int(elID)]
// 	}
// 	return false
// }

func (mm *Model) ColorEdge(isColor bool) {
	// do nothing
}

// func (mm *Model) IgnoreModelElements(ids []uint) {
// 	if len(ids) == 0 {
// 		return
// 	}
// 	ignore := &mm.IgnoreElements
// 	if 0 < mm.actual {
// 		ignore = &mm.Parts[mm.actual-1].IgnoreElements
// 	}
// 	if len(mm.Elements) < len(*ignore) {
// 		*ignore = (*ignore)[:len(mm.Elements)]
// 	}
// 	if len(*ignore) != len(mm.Elements) {
// 		*ignore = append(*ignore, make([]bool, len(mm.Elements)-len(*ignore))...)
// 	}
// 	for _, p := range ids {
// 		(*ignore)[p] = true
// 	}
// }
//
// func (mm *Model) Unignore() {
// 	ignore := &mm.IgnoreElements
// 	if 0 < mm.actual {
// 		ignore = &mm.Parts[mm.actual-1].IgnoreElements
// 	}
// 	*ignore = nil
// }

func (mm *Model) SelectLeftCursor(nodes bool, elements []bool) {
	// do nothing
}

func (mm *Model) GetSelectNodes(single bool) (ids []uint) {
	for i := range mm.Coords {
		if !mm.Coords[i].selected {
			continue
		}
		if mm.Coords[i].Removed {
			continue
		}
		if mm.Coords[i].hided {
			continue
		}
		ids = append(ids, uint(i))
		if single {
			return
		}
	}
	return
}

func (mm *Model) GetSelectElements(single bool, filter func(_ ElType) (acceptable bool)) (ids []uint) {
	for i, el := range mm.Elements {
		if !el.selected {
			continue
		}
		if el.hided {
			continue
		}
		if filter != nil {
			if !filter(el.ElementType) {
				continue
			}
		}
		ids = append(ids, uint(i))
		if single {
			return
		}
	}
	return
}

func (mm *Model) InvertSelect(nodes bool, elements []bool) {
	if nodes {
		for i := range mm.Coords {
			if mm.Coords[i].Removed {
				continue
			}
			mm.Coords[i].selected = !mm.Coords[i].selected
		}
	}
	for el := Line2; el < lastElement; el++ {
		if !elements[el] {
			continue
		}
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != el {
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
	mm.filterByElement(func(e ElType) bool { // filter
		return e == Line2
	}, func(id int) { // run
		if mm.Elements[id].selected {
			return
		}
		var (
			el = mm.Elements[id]
			dx = math.Abs(mm.Coords[el.Indexes[1]].Point3d[0] -
				mm.Coords[el.Indexes[0]].Point3d[0])
			dy = math.Abs(mm.Coords[el.Indexes[1]].Point3d[1] -
				mm.Coords[el.Indexes[0]].Point3d[1])
			dz = math.Abs(mm.Coords[el.Indexes[1]].Point3d[2] -
				mm.Coords[el.Indexes[0]].Point3d[2])
		)
		if x && dy < gog.Eps3D && dz < gog.Eps3D {
			mm.Elements[id].selected = true
		}
		if y && dx < gog.Eps3D && dz < gog.Eps3D {
			mm.Elements[id].selected = true
		}
		if z && dx < gog.Eps3D && dy < gog.Eps3D {
			mm.Elements[id].selected = true
		}
	})
}

func (mm *Model) SelectElementsOnPlane(xoy, yoz, xoz bool, elements []bool) {
	if !xoy && !yoz && !xoz {
		return
	}
	for etype := Line2; etype < lastElement; etype++ {
		if !elements[etype] {
			continue
		}
		// check each elements type
		mm.filterByElement(func(e ElType) bool { // filter
			return e == etype
		}, func(id int) { // run
			el := mm.Elements[id]
			var dx, dy, dz float64
			for p := 1; p < len(el.Indexes); p++ {
				var (
					b  = mm.Coords[el.Indexes[p-1]]
					f  = mm.Coords[el.Indexes[p]]
					dX = math.Abs(f.Point3d[0] - b.Point3d[0])
					dY = math.Abs(f.Point3d[1] - b.Point3d[1])
					dZ = math.Abs(f.Point3d[2] - b.Point3d[2])
				)
				dx = math.Max(dx, dX)
				dy = math.Max(dy, dY)
				dz = math.Max(dz, dZ)
			}
			if xoy && dz < gog.Eps3D {
				mm.Elements[id].selected = true
			}
			if yoz && dx < gog.Eps3D {
				mm.Elements[id].selected = true
			}
			if xoz && dy < gog.Eps3D {
				mm.Elements[id].selected = true
			}
		})
	}
}

func (mm *Model) SelectLinesParallel(lines []uint) {
	// check
	if !mm.isValidElementId(lines, func(e ElType) bool { return e == Line2 }) {
		logger.Printf("SelectLinesParallel: not valid line id: %v", lines)
		return
	}
	// actions
	if len(lines) == 0 {
		// do nothing
		return
	}
	lines = uniqUint(lines)
	// selection
	type ratio struct{ dx, dy, dz float64 }
	toOne := func(el Element) (r ratio, ok bool) {
		r.dx = mm.Coords[el.Indexes[0]].Point3d[0] - mm.Coords[el.Indexes[1]].Point3d[0]
		r.dy = mm.Coords[el.Indexes[0]].Point3d[1] - mm.Coords[el.Indexes[1]].Point3d[1]
		r.dz = mm.Coords[el.Indexes[0]].Point3d[2] - mm.Coords[el.Indexes[1]].Point3d[2]
		amplitude := math.Sqrt(pow.E2(r.dx) + pow.E2(r.dy) + pow.E2(r.dz))
		if amplitude < gog.Eps3D { // zero lenght line
			return
		}
		r.dx /= amplitude
		r.dy /= amplitude
		r.dz /= amplitude
		return r, true
	}
	// create list of vectors
	var ratios []ratio
	for _, n := range lines {
		r, ok := toOne(mm.Elements[n])
		if !ok {
			continue
		}
		ratios = append(ratios, r)
	}
	if len(ratios) == 0 {
		// do nothing
		return
	}
	// comparing of vectors
	same := func(r1, r2 ratio) bool {
		if gog.Eps3D < math.Abs(r1.dx-r2.dx) {
			return false
		}
		if gog.Eps3D < math.Abs(r1.dy-r2.dy) {
			return false
		}
		if gog.Eps3D < math.Abs(r1.dz-r2.dz) {
			return false
		}
		return true
	}
	// selection
	mm.filterByElement(func(e ElType) bool { // filter
		return e == Line2
	}, func(id int) { // run
		r, ok := toOne(mm.Elements[id])
		if !ok {
			return
		}
		found := false
		for i := range ratios {
			if same(r, ratios[i]) {
				found = true
				break
			}
		}
		if !found {
			return
		}
		mm.Elements[id].selected = true
	})
	// var ratios []ratio
	//
	//	for _, p := range lines {
	//		ok := mm.IsVisibleLine(p)
	//		if !ok {
	//			continue
	//		}
	//		r, ok := toOne(mm.Elements[p])
	//		if !ok {
	//			continue
	//		}
	//		ratios = append(ratios, r)
	//	}
	//
	//	for i, el := range mm.Elements {
	//		ok := mm.IsVisibleLine(uint(i))
	//		if !ok {
	//			continue
	//		}
	//		if el.selected {
	//			continue
	//		}
	//		var found bool
	//		for _, p := range lines {
	//			if int(p) == i {
	//				found = true
	//				mm.Elements[p].selected = true
	//				break
	//			}
	//		}
	//		if found {
	//			continue
	//		}
	//		re, ok := toOne(el)
	//		if !ok {
	//			continue
	//		}
	//		for ri := range ratios {
	//			if compare(re, ratios[ri]) {
	//				mm.Elements[i].selected = true
	//			}
	//		}
	//	}
}

func (mm *Model) SelectLinesByLenght(more bool, lenght float64) {
	// check
	for _, p := range []*float64{&lenght} {
		if math.Abs(*p) < gog.Eps3D {
			*p = 0
		}
		if !mm.isValidValue(*p) {
			logger.Printf("SelectLinesByLenght: not valid value: %v", *p)
			return
		}
	}
	// actions
	if lenght <= 0.0 {
		return
	}
	mm.filterByElement(func(e ElType) bool { // filter
		return e == Line2
	}, func(id int) { // run
		var (
			b = mm.Coords[mm.Elements[id].Indexes[0]]
			f = mm.Coords[mm.Elements[id].Indexes[1]]
		)
		L := gog.Distance3d(b.Point3d, f.Point3d)
		if (more && lenght <= L) || (!more && L <= lenght) {
			mm.Elements[id].selected = true
		}
	})
}

func (mm *Model) SelectLinesCylindrical(node uint, radiant, conc bool, axe Direction) {
	// check
	if s := []uint{node}; !mm.isValidNodeId(s) {
		logger.Printf("SelectLinesCylindrical: not valid node id: %v", s)
		return
	}
	// actions
	mm.filterByElement(func(e ElType) bool { // filter
		return e == Line2
	}, func(id int) { // run
		var (
			r0, r1, dr float64
			a0, a1, da float64
			el         = mm.Elements[id]
			ci0        = mm.Coords[el.Indexes[0]]
			ci1        = mm.Coords[el.Indexes[1]]
			ccn        = mm.Coords[node]
		)
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
		if da < gog.Eps3D && radiant {
			mm.Elements[id].selected = true
		}
		if dr < gog.Eps3D && conc {
			mm.Elements[id].selected = true
		}
	})
}

// TODO remove
func (mm *Model) filterByElement(filter func(e ElType) bool, run func(id int)) {
	for id, el := range mm.Elements {
		if !filter(el.ElementType) {
			continue
		}
		if el.hided {
			continue
		}
		run(id)
	}
	// // check
	//
	//	if s := []uint{p}; !mm.isValidElementId(s, func(e ElType) bool { return e == Line2 }) {
	//		logger.Printf("IsVisibleLine: not valid line id: %v", s)
	//		return false
	//	}
	//
	// // actions
	//
	//	if mm.Elements[p].hided {
	//		return false
	//	}
	//
	// // 	if mm.IsIgnore(uint(p)) {
	// // 		return
	// // 	}
	// return true
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

func (mm *Model) Select(nodes, elements []uint) {
	mm.DeselectAll()
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Select: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Select: not valid elements id: %v", s)
		return
	}
	for _, id := range nodes {
		mm.Coords[id].selected = true
	}
	for _, id := range elements {
		mm.Elements[id].selected = true
	}
}

func (mm *Model) SelectAll(nodes bool, elements []bool) {
	if nodes {
		for i := range mm.Coords {
			mm.Coords[i].selected = true
		}
	}
	for el := Line2; el < lastElement; el++ {
		if !elements[el] {
			continue
		}
		for i := range mm.Elements {
			if mm.Elements[i].ElementType != el {
				continue
			}
			mm.Elements[i].selected = true
		}
	}
}

func (mm *Model) SelectScreen(from, to [2]int32) {
	// do nothing
}

func (mm *Model) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	// check
	if s := lines; !mm.isValidElementId(s, nil) {
		logger.Printf("SplitLinesByDistance: not valid lines id: %v", s)
		return
	}
	// actions
	if len(lines) == 0 {
		// do nothing
		return
	}
	if distance == 0 {
		// split by begin/end point
		// do nothing
		return
	}
	defer mm.DeselectAll() // deselect
	cs := mm.Coords
	for _, il := range lines {
		el := mm.Elements[il]
		if el.ElementType != Line2 {
			continue
		}
		length := gog.Distance3d(
			cs[el.Indexes[0]].Point3d,
			cs[el.Indexes[1]].Point3d,
		)
		if length < gog.Eps3D {
			// do nothing
			continue
		}
		// split point on line corner
		if atBegin && math.Abs(length-distance) < gog.Eps3D {
			// point at end point
			continue
		}
		if !atBegin && math.Abs(length+distance) < gog.Eps3D {
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
	// check
	if s := lines; !mm.isValidElementId(s, nil) {
		logger.Printf("SplitLinesByRatio: not valid lines id: %v", s)
		return
	}
	// actions
	if len(lines) == 0 {
		// do nothing
		return
	}
	if proportional == 0 || proportional == 1 {
		return
	}
	defer mm.DeselectAll() // deselect
	for _, il := range lines {
		el := mm.Elements[il]
		if el.ElementType != Line2 {
			continue
		}
		length := gog.Distance3d(
			mm.Coords[el.Indexes[0]].Point3d,
			mm.Coords[el.Indexes[1]].Point3d,
		)
		if length < gog.Eps3D {
			// do nothing
			continue
		}
		mm.SplitLinesByDistance([]uint{il}, proportional*length, atBegin)
	}
}

func (mm *Model) SplitLinesByEqualParts(lines []uint, parts uint) {
	// check
	if s := lines; !mm.isValidElementId(s, nil) {
		logger.Printf("SplitLinesByEqualParts: not valid lines id: %v", s)
		return
	}
	// actions
	if len(lines) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll() // deselect
	if parts < 2 {
		return
	}
	if len(lines) == 0 {
		return
	}
	cs := mm.Coords
	for _, il := range lines {
		el := mm.Elements[il]
		if el.ElementType != Line2 {
			continue
		}
		length := gog.Distance3d(
			cs[el.Indexes[0]].Point3d,
			cs[el.Indexes[1]].Point3d,
		)
		if length < gog.Eps3D {
			// do nothing
			continue
		}
		ids := []uint{uint(el.Indexes[0])}
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
		ids = append(ids, uint(el.Indexes[1]))
		for i := range ids {
			if i == 0 {
				continue
			}
			if i == 1 {
				mm.Elements[il].Indexes[1] = int(ids[1])
				// mm.AddLineByNodeNumber(uint(el.Indexes[0]), ids[0])
				continue
			}
			mm.AddLineByNodeNumber(ids[i-1], ids[i])
		}
		// logger.Printf("SplitLinesByEqualParts: Add Points %v", ids)
		// mm.AddLineByNodeNumber(ids[len(ids)-1], uint(el.Indexes[1]))
		// el.Indexes[1] = int(ids[0])
	}
}

//func (mm *Model) RemoveEmptyNodes() {
// remove Removed Coordinates
// find Coordinates without elements
// remove Removed Elements
// }

func (mm *Model) MergeNodes(minDistance float64) {
	// check
	for _, p := range []*float64{&minDistance} {
		if math.Abs(*p) < gog.Eps3D {
			*p = 0
		}
		if !mm.isValidValue(*p) {
			logger.Printf("MergeNodes: not valid value: %v", *p)
			return
		}
	}
	// actions
	if minDistance < 0.0 {
		return
	}
	if minDistance == 0.0 {
		minDistance = gog.Eps3D
	}
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
			if !gog.SamePoints3d(mm.Coords[i].Point3d, mm.Coords[j].Point3d) {
				// Coordinates are not same
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
	//	type link struct {
	//		less, more int
	//	}
	//
	//	in := make(chan link)
	//	re := make(chan link)
	//
	//	compare := func(less, more int) {
	//		// SQRT(DX^2+DY^2+DZ^2) < D
	//		//      DX^2+DY^2+DZ^2  < D^2
	//		// specific cases:
	//		//	    DX^2            < D^2, DY=0, DZ=0
	//		if more <= less {
	//			return
	//		}
	//		if mm.Coords[less].Removed {
	//			return
	//		}
	//		if mm.Coords[more].Removed {
	//			return
	//		}
	//		if gog.SamePoints3d(mm.Coords[more].Point3d, mm.Coords[less].Point3d) {
	//			// Coordinates are same
	//			re <- link{less: less, more: more}
	//		}
	//	}
	//
	//	size := runtime.NumCPU() // TODO cannot test concerency algorithm
	//	if size < 1 {
	//		size = 1
	//	}
	//	var wg sync.WaitGroup
	//	wg.Add(size)
	//	for i := 0; i < size; i++ {
	//		go func() {
	//			for l := range in {
	//				compare(l.less, l.more)
	//			}
	//			wg.Done()
	//		}()
	//	}
	//
	//	go func() {
	//		wg.Wait()
	//		close(re)
	//	}()
	//
	//	done := make(chan bool)
	//	go func() {
	//		for l := range re {
	//			for i := range mm.Elements {
	//				for j := range mm.Elements[i].Indexes {
	//					if mm.Elements[i].Indexes[j] == l.less {
	//						mm.Elements[i].Indexes[j] = l.more
	//					}
	//				}
	//			}
	//			mm.Coords[l.more].Removed = true
	//		}
	//		done <- true
	//	}()
	//
	//	for i := range mm.Coords {
	//		for j := range mm.Coords {
	//			in <- link{less: i, more: j}
	//		}
	//	}
	//	close(in)
	//
	//	<-done
	//	close(done)
	//
	// TODO loads merge
}

func (mm *Model) MergeLines(lines []uint) {
	// check
	if s := lines; !mm.isValidElementId(s, nil) {
		logger.Printf("MergeLines: not valid lines id: %v", s)
		return
	}
	// actions
	// uniq lines
	lines = uniqUint(lines)
	if len(lines) < 2 {
		// do nothing
		return
	}
	// initialization
	if len(lines) == 0 {
		// do nothing
		return
	}
	// action
	// merge 2 lines rules:
	//	* common point of 2 that lines
	//	* common line index = minimal of lines with common point
	//	* common point not removed
	//	* lines on one line
	iteration := func() bool {
		for i := range lines {
			if mm.Elements[i].ElementType != Line2 {
				continue
			}
			for j := range lines {
				if mm.Elements[j].ElementType != Line2 {
					continue
				}
				if i <= j {
					continue
				}
				// j < i
				l1 := &mm.Elements[lines[i]] // will be removed if ok
				l2 := &mm.Elements[lines[j]]
				// on one line
				if !gog.IsParallelLine3d(
					mm.Coords[(*l1).Indexes[0]].Point3d,
					mm.Coords[(*l1).Indexes[1]].Point3d,
					mm.Coords[(*l2).Indexes[0]].Point3d,
					mm.Coords[(*l2).Indexes[1]].Point3d,
				) {
					continue
				}

				// check common point
				if (*l1).Indexes[0] == (*l2).Indexes[0] {
					(*l2).Indexes[0] = (*l1).Indexes[1]
				} else if (*l1).Indexes[1] == (*l2).Indexes[0] {
					(*l2).Indexes[0] = (*l1).Indexes[0]
				} else if (*l1).Indexes[0] == (*l2).Indexes[1] {
					(*l2).Indexes[1] = (*l1).Indexes[1]
				} else if (*l1).Indexes[1] == (*l2).Indexes[1] {
					(*l2).Indexes[1] = (*l1).Indexes[0]
				} else {
					continue
				}
				mm.Remove(nil, []uint{lines[i]})
				lines = append(lines[:i], lines[i+1:]...)
				return true
			}
		}
		return false
	}
	for iter, maxiter := 0, 100000; iter < maxiter; iter++ {
		if !iteration() {
			break
		}
	}
}

func (mm *Model) ScaleOrtho(
	basePoint gog.Point3d, // point for scaling
	scale [3]float64, // sX, sY, sZ
	nodes, elements []uint, // elements of scaling
) {
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("ScaleOrtho: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("ScaleOrtho: not valid elements id: %v", s)
		return
	}
	// actions
	// check
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll()
	// add nodes from elements
	for _, e := range elements {
		el := mm.Elements[e]
		if el.ElementType == ElRemove {
			continue
		}
		for _, n := range el.Indexes {
			nodes = append(nodes, uint(n))
		}
	}
	nodes = uniqUint(nodes)
	// scaling
	for _, n := range nodes {
		if mm.Coords[n].Removed {
			continue
		}
		p3 := mm.Coords[n].Point3d
		for i := range p3 {
			if scale[i] == 1.0 {
				// do nothing
				continue
			}
			// FMA(x,y,z) = x * y + z
			// p3[i] =  (p3[i]-basePoint[i])*scale[i] + basePoint[i]
			p3[i] = math.FMA(p3[i]-basePoint[i], scale[i], basePoint[i])
		}
		mm.Coords[n].Point3d = p3
	}
}

func (mm *Model) Intersection(nodes, elements []uint) {
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Intersection: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Intersection: not valid elements id: %v", s)
		return
	}
	// actions
	// check
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll()
	// remove not valid coordinates and elements
	mm.RemoveZeroLines()
	mm.RemoveZeroTriangles()
	mm.RemoveSameCoordinates()
	defer func() {
		mm.RemoveZeroLines()
		mm.RemoveZeroTriangles()
		mm.RemoveSameCoordinates()
	}()
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

	var LLTT = [3][2]int{{0, 1}, {1, 2}, {2, 0}}
	intersectTris := func(el0, el1 Element) {
		// Triangle inside triangle
		if intersect, pi := gog.TriangleTriangle3d(
			// coordinates triangle 0
			mm.Coords[el0.Indexes[0]].Point3d,
			mm.Coords[el0.Indexes[1]].Point3d,
			mm.Coords[el0.Indexes[2]].Point3d,
			// coordinates triangle 1
			mm.Coords[el1.Indexes[0]].Point3d,
			mm.Coords[el1.Indexes[1]].Point3d,
			mm.Coords[el1.Indexes[2]].Point3d,
		); intersect {
			for _, p := range pi {
				chNewPoints <- p
			}
		}
		// Triangle edges
		for _, v0 := range LLTT {
			for _, v1 := range LLTT {
				var (
					a0 = mm.Coords[el0.Indexes[v0[0]]].Point3d
					a1 = mm.Coords[el0.Indexes[v0[1]]].Point3d
					b0 = mm.Coords[el1.Indexes[v1[0]]].Point3d
					b1 = mm.Coords[el1.Indexes[v1[1]]].Point3d
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

	intersectLineTri := func(el0, el1 Element) {
		for _, f := range [...]func(
			gog.Point3d, gog.Point3d, gog.Point3d, gog.Point3d, gog.Point3d,
		) (bool, []gog.Point3d){
			gog.LineTriangle3dI1,
			gog.LineTriangle3dI2,
		} {
			if intersect, pi := f(
				// Line2
				mm.Coords[el0.Indexes[0]].Point3d,
				mm.Coords[el0.Indexes[1]].Point3d,
				// Triangle3
				mm.Coords[el1.Indexes[0]].Point3d,
				mm.Coords[el1.Indexes[1]].Point3d,
				mm.Coords[el1.Indexes[2]].Point3d,
			); intersect {
				for _, p := range pi {
					chNewPoints <- p
				}
			}
		}
	}

	intersectLines := func(el0, el1 Element) {
		var (
			a0 = mm.Coords[el0.Indexes[0]].Point3d
			a1 = mm.Coords[el0.Indexes[1]].Point3d
			b0 = mm.Coords[el1.Indexes[0]].Point3d
			b1 = mm.Coords[el1.Indexes[1]].Point3d
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

	for i0, li0 := range elements {
		el0 := mm.Elements[li0]
		if el0.ElementType == ElRemove {
			continue
		}
		for i1, li1 := range elements {
			el1 := mm.Elements[li1]
			if el1.ElementType == ElRemove {
				continue
			}
			if i0 <= i1 {
				continue
			}
			if !gog.BorderIntersection(mm.getPoint3d(li0), mm.getPoint3d(li1)) {
				continue
			}
			// action
			switch {
			case el0.ElementType == Line2 && el1.ElementType == Line2:
				intersectLines(el0, el1)
			case el0.ElementType == Line2 && el1.ElementType == Triangle3:
				intersectLineTri(el0, el1)
			case el1.ElementType == Line2 && el0.ElementType == Triangle3:
				intersectLineTri(el1, el0)
			case el0.ElementType == Triangle3 && el1.ElementType == Triangle3:
				intersectTris(el0, el1)
			default:
				logger.Printf("Intersection: not implemented `%s`-`%s`",
					el0.ElementType, el1.ElementType)
			}
		}
	}
	close(chNewPoints)
	<-stop
	logger.Printf("Intersection: find new %d points", len(newPoints))
	// fix zero coordinates
	for i := range newPoints {
		for j := range newPoints[i] {
			if math.Abs(newPoints[i][j]) < gog.Eps3D {
				newPoints[i][j] = 0.0
			}
		}
	}
	// add all points
	for i := range newPoints {
		id := mm.AddNode(
			newPoints[i][0],
			newPoints[i][1],
			newPoints[i][2],
		)
		nodes = append(nodes, id)
	}
	logger.Printf("Intersection: %d nodes", len(nodes))
	// uniq
	nodes = uniqUint(nodes)
	elements = uniqUint(elements)
	logger.Printf("Intersection: %d nodes", len(nodes))
	// interation of intersection
	for iter := 0; ; iter++ { // TODO avoid infinite
		var newElements []uint
		for _, pe := range elements {
			for _, n := range nodes {
				// avoid Coordinate-Coordinate
				// found := false
				// for _, ind := range mm.Elements[pe].Indexes {
				// 	if gog.SamePoints3d(
				// 		mm.Coords[n].Point3d,
				// 		mm.Coords[ind].Point3d,
				// 	) {
				// 		found = true
				// 		break
				// 	}
				// }
				// if found {
				// 	continue
				// }
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
						// logger.Printf("Intersection 0-1: %v", ok)
						if !ok {
							logger.Printf("Intersection: split point on triangle edge 01 invalid")
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
						// logger.Printf("Intersection 1-2: %v", ok)
						if !ok {
							logger.Printf("Intersection: split point on triangle edge 12 invalid")
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
						// logger.Printf("Intersection 2-0: %v", ok)
						if !ok {
							logger.Printf("Intersection: split point on triangle edge 20 invalid")
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
						// logger.Printf("Intersection Coordinate-Triangle: %v %v", ok0, ok1)
						if !(ok0 && ok1) {
							logger.Printf("Intersection: not valid triangles")
							continue
						}
						mm.Elements[pe].Indexes[1] = int(n)
						newElements = append(newElements, t0, t1)
						continue
					}

				default:
					logger.Printf("Intersection: not implemented point to `%s`",
						mm.Elements[pe].ElementType)
				}
			}
		}
		logger.Printf("Intersection: add %d elements", len(newElements))
		if len(newElements) == 0 {
			break
		}
		elements = append(elements, newElements...)
		if 100 < iter {
			logger.Printf("Intersection iterations break")
			break
		}
	}
	// for i := range mm.Elements {
	// 	fmt.Printf("%d ------- %#v\n", i, mm.Elements[i])
	// 	for _, k := range mm.Elements[i].Indexes {
	// 		fmt.Println(k, mm.Coords[k])
	// 	}
	// 	ps := mm.getPoint3d(uint(i))
	// 	min, max := gog.BorderPoints(ps...)
	// 	fmt.Println(">>>>>", min, max)
	// }
}

func (mm *Model) SplitTri3To3Tri3(elements []uint) {
	// check
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("SplitTri3To3Tri3: not valid elements id: %v", s)
		return
	}
	// actions
	if len(elements) == 0 {
		// do nothing
		return
	}
	// action
	defer mm.DeselectAll() // deselect
	const one3 = 1.0 / 3.0
	for _, eid := range elements {
		el := mm.Elements[eid]
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
		mm.AddTriangle3ByNodeNumber(uint(el.Indexes[0]), uint(el.Indexes[1]), id)
		mm.AddTriangle3ByNodeNumber(uint(el.Indexes[1]), uint(el.Indexes[2]), id)
		mm.Elements[eid].Indexes = []int{el.Indexes[2], el.Indexes[0], int(id)}
	}
}

func (mm *Model) Hide(nodes, elements []uint) {
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Hide: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Hide: not valid elements id: %v", s)
		return
	}
	// actions
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	for _, p := range nodes {
		mm.Coords[p].hided = true
	}
	for _, p := range elements {
		mm.Elements[p].hided = true
	}
	for _, el := range mm.Elements {
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
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Move: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Move: not valid elements id: %v", s)
		return
	}
	// actions
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll() // deselect
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	nodes = uniqUint(nodes)
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
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Copy: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Copy: not valid elements id: %v", s)
		return
	}
	// actions
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	if len(paths) == 0 {
		// do nothing
		return
	}
	defer mm.DeselectAll() // deselect
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	nodes = uniqUint(nodes)
	elements = uniqUint(elements)
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
		// find indexes of each point in new model
		ids := make([]uint, len(el.Indexes))
		for i := range ids {
			id := cModel.AddNode(
				mm.Coords[el.Indexes[i]].Point3d[0],
				mm.Coords[el.Indexes[i]].Point3d[1],
				mm.Coords[el.Indexes[i]].Point3d[2],
			)
			ids[i] = id
		}
		// create element in new model
		switch el.ElementType {
		case Line2:
			cModel.AddLineByNodeNumber(ids[0], ids[1])
		case Triangle3:
			cModel.AddTriangle3ByNodeNumber(ids[0], ids[1], ids[2])
		case Quadr4:
			cModel.AddQuadr4ByNodeNumber(ids[0], ids[1], ids[2], ids[3])
		default:
			logger.Printf("Undefined: %v", el.ElementType)
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

		copyBase := [3]float64{basePoint[0], basePoint[1], basePoint[2]}
		move(&copyBase, basePoint, path)

		mm.AddModel(cModel)
	}
}

func (mm *Model) Mirror(nodes, elements []uint,
	basePoint [3]gog.Point3d,
	copyModel bool,
	addLines, addTri bool) {
	// check
	if s := nodes; !mm.isValidNodeId(nodes) {
		logger.Printf("Mirror: not valid node id: %v", s)
		return
	}
	if s := elements; !mm.isValidElementId(s, nil) {
		logger.Printf("Mirror: not valid elements id: %v", s)
		return
	}
	// actions
	if len(nodes) == 0 && len(elements) == 0 {
		// do nothing
		return
	}
	if onOneLine := gog.PointLine3d(
		basePoint[0],
		basePoint[1],
		basePoint[2],
	) || gog.PointLine3d(
		basePoint[1],
		basePoint[0],
		basePoint[2],
	) || gog.PointLine3d(
		basePoint[2],
		basePoint[0],
		basePoint[1],
	); onOneLine {
		logger.Printf("Mirror: base points of plane")
		return
	}
	// nodes appending
	for _, ie := range elements {
		for _, ind := range mm.Elements[ie].Indexes {
			nodes = append(nodes, uint(ind))
		}
	}
	// uniq indexes
	nodes = uniqUint(nodes)
	elements = uniqUint(elements)
	// prepare mirror points
	var points []gog.Point3d
	for _, n := range nodes {
		points = append(points, mm.Coords[n].Point3d)
	}
	mir := gog.Mirror3d(basePoint, points...)
	// mirror move only nodes
	if !copyModel {
		for i, n := range nodes {
			mm.Coords[n].Point3d = mir[i]
		}
		return
	}
	// copy mirror
	var newID []uint
	for i := range nodes { // create id of mirror nodes
		id := mm.AddNode(mir[i][0], mir[i][1], mir[i][2])
		newID = append(newID, id)
	}
	// create mirror elements
	for _, pe := range elements { // add all elements
		el := mm.Elements[pe]
		// find indexes of each point in new model
		ids := make([]uint, len(el.Indexes))
		for i := range ids { // copy old nodes ids
			ids[i] = uint(el.Indexes[i])
		}
		for i := range ids { // convert to new ids
			for pos, n := range nodes {
				if n != ids[i] {
					continue
				}
				ids[i] = newID[pos]
				break
			}
		}
		// create element in new model
		switch el.ElementType {
		case Line2:
			mm.AddLineByNodeNumber(ids[0], ids[1])
		case Triangle3:
			mm.AddTriangle3ByNodeNumber(ids[0], ids[1], ids[2])
		case Quadr4:
			mm.AddQuadr4ByNodeNumber(ids[0], ids[1], ids[2], ids[3])
		default:
			logger.Printf("Undefined: %v", el.ElementType)
		}
		// add triangles by lines
		if addTri {
			if el.ElementType != Line2 {
				continue
			}
			mm.AddTriangle3ByNodeNumber(uint(el.Indexes[0]), ids[0], uint(el.Indexes[1]))
			mm.AddTriangle3ByNodeNumber(uint(el.Indexes[1]), ids[0], ids[1])
		}
	}
	if addLines {
		for i := range nodes {
			if newID[i] == nodes[i] {
				// zero lenght line
				continue
			}
			mm.AddLineByNodeNumber(nodes[i], newID[i])
		}
	}
}

func (mm *Model) StandardView(view SView) {
	// do nothing
}

func (mm *Model) GetRootGroup() groups.Group {
	return &mm.Groups.meta
}

func (mm *Model) Update(nodes, elements *uint) {
	logger.Printf("not implemented Update")
	// TODO
}

///////////////////////////////////////////////////////////////////////////////

func max(xs ...float64) (res float64) {
	if len(xs) == 0 {
		logger.Printf("not valid: zero lenght")
	}
	res = xs[0]
	for i := range xs {
		res = math.Max(res, xs[i])
	}
	return
}

func min(xs ...float64) (res float64) {
	if len(xs) == 0 {
		logger.Printf("not valid: zero lenght")
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

var testCoverageFunc func(
	m Mesh,
	ch *chan ds.Action,
	screenshot func(check func(image.Image)),
)

////////////////////////////////////////////////////////////////////////////////

func (mm *Model) isValidValue(v float64) bool {
	if math.IsNaN(v) {
		return false
	}
	if math.IsInf(v, 0) {
		return false
	}
	return true
}

func (mm *Model) isValidNodeId(ids []uint) bool {
	for _, id := range ids {
		if id < 0 || len(mm.Coords) <= int(id) {
			return false
		}
		if mm.Coords[id].Removed {
			return false
		}
		if mm.Coords[id].hided {
			return false
		}
	}
	return true
}

func (mm *Model) isValidElementId(ids []uint, filter func(ElType) bool) bool {
	for _, id := range ids {
		if id < 0 || len(mm.Elements) <= int(id) {
			logger.Printf("isValidElementId: outside of range: %d", id)
			return false
		}
		if mm.Elements[id].ElementType == ElRemove {
			logger.Printf("isValidElementId: removed: %d", id)
			return false
		}
		if filter != nil && !filter(mm.Elements[id].ElementType) {
			logger.Printf("isValidElementId: not valid type: %d", id)
			return false
		}
		if mm.Elements[id].hided {
			logger.Printf("isValidElementId: hided element: %d", id)
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

func Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%v\n%v", err, r, string(debug.Stack()))
		}
	}()
	quit := make(chan struct{})
	// initialize undo chain
	var (
		mm        Undo
		closedApp = false
		ch        = make(chan ds.Action, 1000)
		ws        = [2]ds.Window{
			new(window.Empty),
			new(window.Empty),
		}
	)

	// use unicode symbols
	vl.SpecificSymbol(false)

	// prepare model
	// if filename == "" {
	mm.model = new(Model)
	groups.FixMesh(&mm)
	mm.quit = &quit
	// } else if strings.HasSuffix(strings.ToLower(filename), FileExtension) {
	// 	// read native json file format
	// 	var b []byte
	// 	b, err = ioutil.ReadFile(filename)
	// 	if err != nil {
	// 		return
	// 	}
	// 	var model Model
	// 	if err = json.Unmarshal(b, &model); err != nil {
	// 		return
	// 	}
	// 	mm.model = &model
	// } else if strings.HasSuffix(strings.ToLower(filename), ".geo") {
	// 	// read gmsh file
	// 	var model Model
	// 	model, err = Gmsh2Model(filename)
	// 	if err != nil {
	// 		return
	// 	}
	// 	mm.model = &model
	// } else {
	// 	err = fmt.Errorf("not valid input data: `%s`", filename)
	// 	return
	// }

	// tui
	tui, initialization, err := NewTui(&mm, &closedApp, &ch)
	if err != nil {
		return
	}
	tuiWindow := window.NewTui(tui)
	ws[0] = tuiWindow
	mm.addTuiInitialization(initialization)
	initialization()

	// 3d model
	opWindow, err := NewOpengl(&mm, &ch)
	if err != nil {
		return
	}
	ws[1] = opWindow

	// screen initialization
	screen, err := ds.New("ms", ws, &ch)
	// TODO filename like: ms. FILENAME in dialog header
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

	//mm.tui = tui // TODO Why????
	mm.op = opWindow

	// run test function
	go func() {
		if f := testCoverageFunc; f != nil {
			testCoverageFunc(&mm, &ch, screen.Screenshot)
		}
	}()
	// run and stop
	ch <- func() (fus bool) {
		screen.ChangeRatio(0.4) // TODO: add to interface
		return false
	}
	screen.Run(&quit)
	closedApp = true
	time.Sleep(2 * time.Second)
	close(ch)
	return
}

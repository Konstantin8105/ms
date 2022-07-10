package ms

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Konstantin8105/tf"
	"github.com/Konstantin8105/vl"
)

type GroupId uint8

const (
	Add GroupId = iota
	Ignore
	Split
	// Plate
	MoveCopy
	// 	Scale
	// 	TypModels
	// 	Check
	// 	Plugin
	endGroup
)

func (g GroupId) String() string {
	switch g {
	case Add:
		return "Add"
	case Ignore:
		return "Ignore"
	case Split:
		return "Split"
		// 	case Plate:
		// 		return "Plate operations"
	case MoveCopy:
		return "Move/Copy"
		// 	case Scale:
		// 		return "Scale"
		// 	case Check:
		// 		return "Check"
		// 	case TypModels:
		// 		return "Typical models"
		// 	case Plugin:
		// 		return "Plugin"
	}
	return fmt.Sprintf("Undefined:%02d", g)
}

type Filable interface {
	// Open
	// Save
	// SaveAs
	// Close
}

type Editable interface {
	// Undo
	// Redo
}

type Viewable interface {
	// Wireframe mode
	// Solid mode
	// Standard view +X
	// Standard view -X
	// Standard view +Y
	// Standard view -Y
	// Standard view +Z
	// Standard view -Z
	// Isometric views
}

type Addable interface {
	AddNode(X, Y, Z float64) error
	AddLineByNodeNumber(n1, n2 uint) error
	AddTriangle3ByNodeNumber(n1, n2, n3 uint) error
	// TODO REMOVE AddQuadr4ByNodeNumber(n1, n2, n3, n4 string)
	// TODO REMOVE AddElementsByNodes(ids string, l2, t3, q4 bool)
	// AddGroup
	// AddCrossSections
}

func init() {
	group := Add
	name := group.String()
	ops := []Operation{{
		Name: "Node by coordinate [X,Y,Z]",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			w, gt := Input3Float(
				[3]string{"X", "Y", "Z"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				var vs [3]float64
				var ok bool
				for i := range vs {
					vs[i], ok = gt[i]()
					if !ok {
						return
					}
				}
				m.AddNode(vs[0], vs[1], vs[2])
			}
			list.Add(&b)
			return &list
		}}, {
		Group: Add,
		Name:  "Line2 by nodes",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			b, bgt := Select("Select node1", Single, m.SelectNodes)
			list.Add(b)
			e, egt := Select("Select node2", Single, m.SelectNodes)
			list.Add(e)

			var bi vl.Button
			bi.SetText(name)
			bi.OnClick = func() {
				b, ok := isOne(bgt)
				if !ok {
					return
				}
				e, ok := isOne(egt)
				if !ok {
					return
				}
				m.AddLineByNodeNumber(b, e)
			}
			list.Add(&bi)

			return &list
		}}, {
		Group: Add,
		Name:  "Triangle3 by nodes",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			n1, n1gt := Select("Select node1", Single, m.SelectNodes)
			list.Add(n1)
			n2, n2gt := Select("Select node2", Single, m.SelectNodes)
			list.Add(n2)
			n3, n3gt := Select("Select node3", Single, m.SelectNodes)
			list.Add(n3)

			var bi vl.Button
			bi.SetText(name)
			bi.OnClick = func() {
				n1, ok := isOne(n1gt)
				if !ok {
					return
				}
				n2, ok := isOne(n2gt)
				if !ok {
					return
				}
				n3, ok := isOne(n3gt)
				if !ok {
					return
				}
				m.AddTriangle3ByNodeNumber(n1, n2, n3)
			}
			list.Add(&bi)

			return &list
		}}}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

type Ignorable interface {
	IgnoreModelElements(ids []uint)
	Unignore()
}

func init() {
	group := Ignore
	name := group.String()
	ops := []Operation{{
		Name: "Ignore elements",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			elf, elfgt := Select("Select elements", Many, m.SelectElements)
			list.Add(elf)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				els := elfgt()
				if len(els) == 0 {
					return
				}
				m.IgnoreModelElements(els)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Clear ignoring elements",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var b vl.Button
			b.SetText("Clear")
			b.OnClick = func() {
				m.Unignore()
			}
			list.Add(&b)

			return &list
		}},
	}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

type Selectable interface {
	SelectNodes(single bool) (ids []uint)
	SelectLines(single bool) (ids []uint)
	SelectTriangles(single bool) (ids []uint)
	SelectElements(single bool) (ids []uint)
	// TODO REMOVE SelectQuadr4(single bool) (ids []uint)
	// InvertNodes
	// InvertLines
	// InvertTriangles
	// InvertQuadr4
	//
	// SelectParallelLines
	// SelectParallelTriangles // XY, YZ, XZ
	// SelectParallelQuadr4
	//
	// SelectByGroup
	//
	// Cursor select node
	// Cursor select beams
	// Cursor select plates
}

// func init() {
// 	ops := []Operation{{}}
// 	for i := range ops {
// 		ops[i].Group = Select
// 	}
// 	Operations = append(Operations, ops...)
// }

type Splitable interface {
	SplitLinesByDistance(lines []uint, distance float64, atBegin bool)
	SplitLinesByRatio(lines []uint, proportional float64, atBegin bool)
	SplitLinesByEqualParts(lines []uint, parts uint)
	SplitTri3To3Tri3(tris []uint)
	// TODO REMOVE SplitTri3To3Quadr4(tris string)
	// SplitTri3To2Tri3(tris string, side uint)
	// SplitQuadr4To2Quadr4(q4s string, side uint)
	// Quadr4 to 4 Triangle3
	// Quadr4 to 4 Quadr4
	// Triangles3, Quadrs4 by Lines2
}

func init() {
	ops := []Operation{{
		Name: "Line2 by distance from node",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			s, sgt := Select("Select lines", Many, m.SelectLines)
			list.Add(s)
			d, dgt := InputFloat("Distance", "meter")
			list.Add(d)

			var rg vl.RadioGroup
			rg.SetText([]string{"from line begin", "from line end"})
			list.Add(&rg)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				d, ok := dgt()
				if !ok {
					return
				}
				m.SplitLinesByDistance(sgt(), d, rg.GetPos() == 0)
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Line2 by ratio",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			s, sgt := Select("Select line", Many, m.SelectLines)
			list.Add(s)
			d, dgt := InputFloat("Ratio", "")
			list.Add(d)

			var rg vl.RadioGroup
			rg.SetText([]string{"from line begin", "from line end"})
			list.Add(&rg)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				r, ok := dgt()
				if !ok {
					return
				}
				m.SplitLinesByRatio(sgt(), r, rg.GetPos() == 0)
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Line2 to equal parts",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			ns, nsgt := Select("Select lines", Many, m.SelectLines)
			list.Add(ns)

			r, rgt := InputUnsigned("Amount parts", "")
			list.Add(r)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				m.SplitLinesByEqualParts(nsgt(), rgt())
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Triangle3 to 3 Triangle3",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			ns, nsgt := Select("Select triangles3", Many, m.SelectTriangles)
			list.Add(ns)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				m.SplitTri3To3Tri3(nsgt())
			}
			list.Add(&bi)

			return &list
		}},
	}
	for i := range ops {
		ops[i].Group = Split
	}
	Operations = append(Operations, ops...)
}

type MoveCopyble interface {
	MoveCopyNodesDistance(nodes, elements []uint, coordinates [3]float64, copy, addLines, addTri bool)
	MoveCopyNodesN1N2(nodes, elements []uint, from, to uint, copy, addLines, addTri bool)
	// Move/Copy to specific plane",
	// Rotate",
	// Mirror",
	// Copy by line path",
	// Translational repeat",
	// Circular repeat/Spiral",
}

func init() {
	ops := []Operation{{
		Name: "Move/Copy by distance [dX,dY,dZ]",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			w, gt := Input3Float(
				[3]string{"dX", "dY", "dZ"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)

			var chLines vl.CheckBox
			chLines.SetText("Add intermediant lines")
			list.Add(&chLines)

			var chTriangles vl.CheckBox
			chTriangles.SetText("Add intermediant triangles")
			list.Add(&chTriangles)

			var rg vl.RadioGroup
			rg.SetText([]string{"Move", "Copy"})
			list.Add(&rg)

			var b vl.Button
			b.SetText("Move/Copy")
			b.OnClick = func() {
				var vs [3]float64
				var ok bool
				for i := range vs {
					vs[i], ok = gt[i]()
					if !ok {
						return
					}
				}
				m.MoveCopyNodesDistance(coordgt(), elgt(), vs, rg.GetPos() == 1,
					chLines.Checked, chTriangles.Checked)
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Move/Copy from node n1 to node n2",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			nf, nfgt := Select("From node", Single, m.SelectNodes)
			list.Add(nf)

			nt, ntgt := Select("To node", Single, m.SelectNodes)
			list.Add(nt)

			var chLines vl.CheckBox
			chLines.SetText("Add intermediant lines")
			list.Add(&chLines)

			var chTriangles vl.CheckBox
			chTriangles.SetText("Add intermediant triangles")
			list.Add(&chTriangles)

			var rg vl.RadioGroup
			rg.SetText([]string{"Move", "Copy"})
			list.Add(&rg)

			var b vl.Button
			b.SetText("Move/Copy")
			b.OnClick = func() {
				f := nfgt()
				if len(f) != 1 {
					return
				}
				t := ntgt()
				if len(t) != 1 {
					return
				}
				m.MoveCopyNodesN1N2(coordgt(), elgt(), f[0], t[0], rg.GetPos() == 1,
					chLines.Checked, chTriangles.Checked)
			}
			list.Add(&b)
			return &list
		}},
	}
	for i := range ops {
		ops[i].Group = MoveCopy
	}
	Operations = append(Operations, ops...)
}

type Platable interface {
	// Triangulation by nodes
	// Triangulation exist plates by area
	// Smooth mesh
}

// func init() {
// 	ops := []Operation{{
// 	}}
// 	for i := range ops {
// 		ops[i].Group = Plate
// 	}
// 	Operations = append(Operations, ops...)
// }

type Scalable interface {
	// By ratio [sX,sY,sZ] and node
	// By cylinder system coordinate
	// By direction on 2 nodes
}

// func init() {
// 	ops := []Operation{{}}
// 	for i := range ops {
// 		ops[i].Group = Scale
// 	}
// 	Operations = append(Operations, ops...)
// }

type Checkable interface {
	// Multiple structures
	// Node duplicate
	// Beam duplicate
	// Plate duplicate
	// Zero length beam
	// Zero length plates
	// Overlapping Collinear beams
	// Empty loads
	// Empty combinations
	// Not connected nodes
	// Unused supports
	// Unused beam properties
	// All ortho elements
}

// func init() {
// 	ops := []Operation{{}}
// 	for i := range ops {
// 		ops[i].Group = Check
// 	}
// 	Operations = append(Operations, ops...)
// }

type Measurementable interface {
	// Distance between 2 nodes
	// Distance between 2 parallel beam
	// Distance between 2 parallel plates
}

// func init() {
// 	ops := []Operation{{}}
// 	for i := range ops {
// 		ops[i].Group = Measurement
// 	}
// 	Operations = append(Operations, ops...)
// }

type Pluginable interface {
	// Cylinder
	// Sphere
	// Cone
	// Disk
	// Cube
	// Pipe branch
	// Frame
	// Beam-beam connection
	// Column-beam connection
	// Column-column connection
	// Beam intersection
	// Merge nodes
	// Merge beams
	// Merge plates
	// Plate intersection
	// Chamfer plates
	// Fillet plates
	// Explode plates
	// Lines offset by direction
	// Split plates by lines
	// Split lines by plates
	// Convert triangles to rectangles
	// Convert rectangles to triangles
	// Plate bending
	// Triangulation
	// 2D offset
	// Twist
	// Extrude
	// Hole circle, square, rectangle on direction
	// Cutoff
	// Bend plates
	// Stamping by point
	// Stiffening rib
	// Weld
}

type Mesh interface {
	Viewable
	Addable
	Ignorable
	Selectable
	// TODO Intersection
	Platable
	Splitable
	MoveCopyble
	Scalable
	Checkable
	Pluginable
	Measurementable
}

const (
	Single = true
	Many   = false
)

type Operation struct {
	Group GroupId
	Name  string
	Part  func(m Mesh) (w vl.Widget)
}

var Operations []Operation

func InputUnsigned(prefix, postfix string) (w vl.Widget, gettext func() uint) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText("2")
	in.Filter(tf.UnsignedInteger)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, func() uint {
		text := in.GetText()
		value, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return 0
		}
		return uint(value)
	}
}

func InputFloat(prefix, postfix string) (w vl.Widget, gettext func() (_ float64, ok bool)) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))

	const Default = "0.000"

	in.SetText(Default)
	in.Filter(tf.Float)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, func() (v float64, ok bool) {
		str := in.GetText()
		clearValue(str)
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			in.SetText(Default)
			return
		}
		return v, true
	}
}

func Input3Float(prefix, postfix [3]string) (w vl.Widget, gettext [3]func() (float64, bool)) {
	var list vl.List
	for i := 0; i < 3; i++ {
		w, gt := InputFloat(prefix[i], postfix[i])
		list.Add(w)
		gettext[i] = gt
	}
	return &list, gettext
}

func SelectAll(m Mesh) (
	w vl.Widget,
	getCoords func() []uint,
	getElements func() []uint,
) {
	var l vl.ListH
	l.Add(vl.TextStatic("Select nodes and elements"))
	const Default = "NONE"

	var coords vl.Text
	coords.SetText(Default)
	l.Add(&coords)

	var els vl.Text
	els.SetText(Default)
	l.Add(&els)

	var b vl.Button
	b.SetText("Select")
	b.OnClick = func() {
		coordinates := m.SelectNodes(Many)
		elements := m.SelectElements(Many)
		if len(coordinates) == 0 && len(elements) == 0 {
			return
		}
		coords.SetText(fmt.Sprintf("%v", coordinates))
		els.SetText(fmt.Sprintf("%v", elements))
	}
	l.Add(&b)
	return &l, func() (ids []uint) {
			return convertUint(coords.GetText())
		}, func() (ids []uint) {
			return convertUint(els.GetText())
		}
}

func Select(name string, single bool, selector func(single bool) []uint) (
	w vl.Widget,
	gettext func() []uint,
) {
	var l vl.ListH
	l.Add(vl.TextStatic(name))
	// For avoid Inputbox
	var id vl.Text
	//
	// Base solution with Inputbox
	// 	var id vl.Inputbox
	// 	id.Filter(tf.UnsignedInteger)

	const Default = "NONE"

	id.SetText(Default)
	l.Add(&id)
	var b vl.Button
	b.SetText("Select")
	b.OnClick = func() {
		ids := selector(single)
		if len(ids) == 0 {
			return
		}
		if single && 1 < len(ids) {
			id.SetText(Default)
			return
		}
		id.SetText(fmt.Sprintf("%v", ids))
	}
	l.Add(&b)
	return &l, func() (ids []uint) {
		return convertUint(id.GetText())
	}
}

///////////////////////////////////////////////////////////////////////////////

var Debug []string

func UserInterface() (root vl.Widget, action chan func(), err error) {
	var (
		scroll vl.Scroll
		list   vl.List
	)
	root = &scroll
	scroll.Root = &list
	action = make(chan func())

	view := make([]bool, len(Operations))
	colHeader := make([]vl.CollapsingHeader, endGroup)
	for g := range colHeader {
		colHeader[g].SetText(GroupId(g).String())
		var sublist vl.List
		colHeader[g].Root = &sublist
		list.Add(&colHeader[g])
	}
	for g := range colHeader {
		for i := range Operations {
			if Operations[i].Group != GroupId(g) {
				continue
			}
			var c vl.CollapsingHeader
			c.SetText(Operations[i].Name)
			part := Operations[i].Part
			if part == nil {
				err = fmt.Errorf("widget %02d is empty: %#v", i, Operations[i])
				return
			}
			r := part(&mm)
			c.Root = r
			colHeader[g].Root.(*vl.List).Add(&c)
			view[i] = true
		}
	}
	{
		var nums []int
		for i := range view {
			if !view[i] {
				nums = append(nums, i)
			}
		}
		if len(nums) != 0 {
			err = fmt.Errorf("do not view next operations: %v", nums)
		}
	}

	// TODO : var txt vl.Text
	// TODO : list.Add(&txt)
	// TODO : go func() {
	// TODO : 	<-time.After(time.Millisecond * 500)
	// TODO : 	var str string
	// TODO : 	for i := range Debug {
	// TODO : 		str += Debug[i] + "\n"
	// TODO : 	}
	// TODO : 	txt.SetText(str)
	// TODO : }()

	return
}

func isOne(f func() []uint) (value uint, ok bool) {
	b := f()
	if len(b) != 1 {
		return
	}
	return b[0], true
}

func clearValue(str string) string {
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")
	return str
}

func convertUint(str string) (ids []uint) {
	clearValue(str)
	fs := strings.Fields(str)
	for i := range fs {
		u, err := strconv.ParseUint(fs[i], 10, 64)
		if err != nil {
			return
		}
		ids = append(ids, uint(u))
	}
	return
}

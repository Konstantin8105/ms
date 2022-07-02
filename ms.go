package ms

import (
	"fmt"

	"github.com/Konstantin8105/tf"
	"github.com/Konstantin8105/vl"
)

type GroupId uint8

const (
	Add GroupId = iota
	Split
	Plate
	MoveCopy
	Scale
	TypModels
	Check
	Plugin
	endGroup
)

func (g GroupId) String() string {
	switch g {
	case Add:
		return "Add"
	case Split:
		return "Split"
	case Plate:
		return "Plate operations"
	case MoveCopy:
		return "Move/Copy"
	case Scale:
		return "Scale"
	case Check:
		return "Check"
	case TypModels:
		return "Typical models"
	case Plugin:
		return "Plugin"
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
	AddNode(X, Y, Z string)
	AddNodeByDistance(line, distance string, pos uint)
	AddNodeByProportional(line, proportional string, pos uint)
	AddLineByNodeNumber(n1, n2 string)
	AddTriangle3ByNodeNumber(n1, n2, n3 string)
	AddQuadr4ByNodeNumber(n1, n2, n3, n4 string)
	AddElementsByNodes(ids string, l2, t3, q4 bool)

	// AddGroup
	// AddCrossSections
}

type Selectable interface {
	SelectNodes(single bool) (ids []uint)
	SelectLines(single bool) (ids []uint)
	SelectTriangles(single bool) (ids []uint)
	SelectQuadr4(single bool) (ids []uint)
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
}

type Splitable interface {
	SplitLinesByRatio(lines, ratio string)
	SplitLinesByEqualParts(lines, parts string)
	SplitTri3To3Quadr4(tris string)
	SplitTri3To2Tri3(tris string, side uint)
	SplitQuadr4To2Quadr4(q4s string, side uint)
}

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

type Measurementable interface {
	// Distance between 2 nodes
	// Distance between 2 parallel beam
	// Distance between 2 parallel plates
}

type Mesh interface {
	Viewable
	Addable
	Selectable
	Splitable
	Checkable
	Measurementable

	MoveCopyNodesDistance(nodes string, coordinates [3]string, copy bool)
	MoveCopyNodesN1N2(nodes, from, to string, copy bool)
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

var Operations = []Operation{{
	Group: Add,
	Name:  "Node by coordinate [X,Y,Z]",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List

		w, gt := Input3Float(
			[3]string{"X", "Y", "Z"},
			[3]string{"meter", "meter", "meter"},
		)
		list.Add(w)

		var b vl.Button
		b.SetText("Add")
		b.OnClick = func() {
			var vs [3]string
			for i := range vs {
				vs[i] = gt[i]()
			}
			m.AddNode(gt[0](), gt[1](), gt[2]())
		}
		list.Add(&b)
		return &list
	}}, {
	Group: Split,
	Name:  "Line2 by distance from node",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		s, sgt := Select("Select line", Single, m.SelectLines)
		list.Add(s)
		d, dgt := InputFloat("Distance", "meter")
		list.Add(d)

		var rg vl.RadioGroup
		rg.SetText([]string{"from line begin", "from line end"})
		list.Add(&rg)

		var bi vl.Button
		bi.SetText("Add")
		bi.OnClick = func() {
			m.AddNodeByDistance(sgt(), dgt(), rg.GetPos())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Split,
	Name:  "Line2 by ratio",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		s, sgt := Select("Select line", Single, m.SelectLines)
		list.Add(s)
		d, dgt := InputFloat("Ratio", "")
		list.Add(d)

		var rg vl.RadioGroup
		rg.SetText([]string{"from line begin", "from line end"})
		list.Add(&rg)

		var bi vl.Button
		bi.SetText("Split")
		bi.OnClick = func() {
			m.AddNodeByProportional(sgt(), dgt(), rg.GetPos())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Add,
	Name:  "Line2 by node numbers",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		b, bgt := Select("Select node1", Single, m.SelectNodes)
		list.Add(b)
		e, egt := Select("Select node2", Single, m.SelectNodes)
		list.Add(e)

		var bi vl.Button
		bi.SetText("Add")
		bi.OnClick = func() {
			m.AddLineByNodeNumber(bgt(), egt())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Add,
	Name:  "Triangle3 by node numbers",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		n1, n1gt := Select("Select node1", Single, m.SelectNodes)
		list.Add(n1)
		n2, n2gt := Select("Select node2", Single, m.SelectNodes)
		list.Add(n2)
		n3, n3gt := Select("Select node3", Single, m.SelectNodes)
		list.Add(n3)

		var bi vl.Button
		bi.SetText("Add")
		bi.OnClick = func() {
			m.AddTriangle3ByNodeNumber(n1gt(), n2gt(), n3gt())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Add,
	Name:  "Quadr4 by node numbers",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		n1, n1gt := Select("Select node1", Single, m.SelectNodes)
		list.Add(n1)
		n2, n2gt := Select("Select node2", Single, m.SelectNodes)
		list.Add(n2)
		n3, n3gt := Select("Select node3", Single, m.SelectNodes)
		list.Add(n3)
		n4, n4gt := Select("Select node4", Single, m.SelectNodes)
		list.Add(n4)

		var bi vl.Button
		bi.SetText("Add")
		bi.OnClick = func() {
			m.AddQuadr4ByNodeNumber(n1gt(), n2gt(), n3gt(), n4gt())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Add,
	Name:  "Elements by sequence of nodes",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		ns, nsgt := Select("Select sequence of nodes", Many, m.SelectNodes)
		list.Add(ns)

		var l2 vl.CheckBox
		l2.SetText("add lines")
		list.Add(&l2)

		var tr3 vl.CheckBox
		tr3.SetText("add triangles")
		list.Add(&tr3)

		var q4 vl.CheckBox
		q4.SetText("add quadr4")
		list.Add(&q4)

		var bi vl.Button
		bi.SetText("Add")
		bi.OnClick = func() {
			if !(l2.Checked || tr3.Checked || q4.Checked) {
				return
			}
			m.AddElementsByNodes(nsgt(), l2.Checked, tr3.Checked, q4.Checked)
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Split,
	Name:  "Line2 to equal parts",
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
	Group: Split,
	Name:  "Triangle3 to 3 Quadr4",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		ns, nsgt := Select("Select triangles3", Many, m.SelectTriangles)
		list.Add(ns)

		var bi vl.Button
		bi.SetText("Split")
		bi.OnClick = func() {
			m.SplitTri3To3Quadr4(nsgt())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Split,
	Name:  "Triangle3 to 2 Triangle3 by side",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		ns, nsgt := Select("Select triangles3", Many, m.SelectTriangles)
		list.Add(ns)

		var rg vl.RadioGroup
		rg.SetText([]string{"by side1", "by side2", "by side3"})
		list.Add(&rg)

		var bi vl.Button
		bi.SetText("Split")
		bi.OnClick = func() {
			m.SplitTri3To2Tri3(nsgt(), rg.GetPos())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Split,
	Name:  "Quadr4 to 2 equal Quadr4 by side",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List
		ns, nsgt := Select("Select quadr4", Many, m.SelectQuadr4)
		list.Add(ns)

		var rg vl.RadioGroup
		rg.SetText([]string{"by side1, side3", "by side2, side4"})
		list.Add(&rg)

		var bi vl.Button
		bi.SetText("Split")
		bi.OnClick = func() {
			m.SplitQuadr4To2Quadr4(nsgt(), rg.GetPos())
		}
		list.Add(&bi)

		return &list
	}}, {
	// Group: Split,
	// Name:  "Quadr4 to 4 Triangle3",
	// Part: func(m Mesh) (w vl.Widget) {
	// 	return vl.TextStatic("HOLD")
	// }}, {
	// Group: Split,
	// Name:  "Quadr4 to 4 Quadr4",
	// Part: func(m Mesh) (w vl.Widget) {
	// 	return vl.TextStatic("HOLD")
	// }}, {
	// Group: Split,
	// Name:  "Triangles3, Quadrs4 by Lines2",
	// Part: func(m Mesh) (w vl.Widget) {
	// 	return vl.TextStatic("HOLD")
	// }}, {

	Group: MoveCopy,
	Name:  "Move/Copy nodes by distance [dX,dY,dZ]",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List

		ns, ngt := Select("Select nodes", Many, m.SelectNodes)
		list.Add(ns)

		w, gt := Input3Float(
			[3]string{"dX", "dY", "dZ"},
			[3]string{"meter", "meter", "meter"},
		)
		list.Add(w)

		var rg vl.RadioGroup
		rg.SetText([]string{"Move", "Copy"})
		list.Add(&rg)

		var b vl.Button
		b.SetText("Move/Copy")
		b.OnClick = func() {
			var vs [3]string
			for i := range vs {
				vs[i] = gt[i]()
			}
			m.MoveCopyNodesDistance(ngt(), vs, rg.GetPos() == 1)
		}
		list.Add(&b)
		return &list
	}}, {
	Group: MoveCopy,
	Name:  "Move/Copy from node n1 to node n2",
	Part: func(m Mesh) (w vl.Widget) {
		var list vl.List

		ns, ngt := Select("Select nodes", Many, m.SelectNodes)
		list.Add(ns)

		nf, nfgt := Select("From node", Single, m.SelectNodes)
		list.Add(nf)

		nt, ntgt := Select("To node", Single, m.SelectNodes)
		list.Add(nt)

		var rg vl.RadioGroup
		rg.SetText([]string{"Move", "Copy"})
		list.Add(&rg)

		var b vl.Button
		b.SetText("Move/Copy")
		b.OnClick = func() {
			m.MoveCopyNodesN1N2(ngt(), nfgt(), ntgt(), rg.GetPos() == 1)
		}
		list.Add(&b)
		return &list
	}}, {
	Group: MoveCopy,
	Name:  "Move/Copy to specific plane",
	Part: func(m Mesh) (w vl.Widget) {
		// XOY
		// XOZ
		// YOZ
		return vl.TextStatic("HOLD")
	}}, {
	Group: MoveCopy,
	Name:  "Rotate",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: MoveCopy,
	Name:  "Mirror",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {

	Group: MoveCopy,
	Name:  "Copy by line path",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: MoveCopy,
	Name:  "Translational repeat",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: MoveCopy,
	Name:  "Circular repeat/Spiral",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {

	Group: Plate,
	Name:  "Triangulation by nodes",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Plate,
	Name:  "Triangulation exist plates by area",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Plate,
	Name:  "Smooth exist plates",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Scale,
	Name:  "By ratio and node number",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Scale,
	Name:  "By ratio and coordinate",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Scale,
	Name:  "By cylinder system coordinate",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Scale,
	Name:  "By direction on 2 nodes",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}},
}

func InputUnsigned(prefix, postfix string) (w vl.Widget, gettext func() string) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText("2")
	in.Filter(tf.UnsignedInteger)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, in.GetText
}

func InputFloat(prefix, postfix string) (w vl.Widget, gettext func() string) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText("0.000")
	in.Filter(tf.Float)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, in.GetText
}

func Input3Float(prefix, postfix [3]string) (w vl.Widget, gettext [3]func() string) {
	var list vl.List
	for i := 0; i < 3; i++ {
		w, gt := InputFloat(prefix[i], postfix[i])
		list.Add(w)
		gettext[i] = gt
	}
	return &list, gettext
}

func Select(name string, single bool, selector func(single bool) []uint) (
	w vl.Widget,
	gettext func() string,
) {
	var l vl.ListH
	l.Add(vl.TextStatic(name))
	// For avoid Inputbox
	var id vl.Text
	//
	// Base solution with Inputbox
	// 	var id vl.Inputbox
	// 	id.Filter(tf.UnsignedInteger)

	id.SetText("NONE")
	l.Add(&id)
	var b vl.Button
	b.SetText("Select")
	b.OnClick = func() {
		ids := selector(single)
		if len(ids) == 0 {
			return
		}
		if single && 1 < len(ids) {
			ids = ids[:1]
		}
		id.SetText(fmt.Sprintf("%v", ids))
	}
	l.Add(&b)
	return &l, id.GetText
}

var Debug []string

func UserInterface() (root vl.Widget, action chan func(), err error) {
	var m DebugMesh
	var (
		scroll vl.Scroll
		list   vl.List
	)
	root = &scroll
	scroll.Root = &list
	action = make(chan func())
	list.Add(vl.TextStatic(fmt.Sprintf("Amount operations: %d", len(Operations))))

	view := make([]bool, len(Operations))
	colHeader := make([]vl.CollapsingHeader, endGroup)
	for g := range colHeader {
		colHeader[g].SetText(GroupId(g).String())
		var sublist vl.List
		colHeader[g].Root = &sublist
		list.Add(&colHeader[g])
		//list.Add(new(vl.Separator))
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
				err = fmt.Errorf("Widget %02d is empty\n", i)
				return
			}
			r := part(m)
			c.Root = r
			colHeader[g].Root.(*vl.List).Add(&c)
			// colHeader[g].Root.(*vl.List).Add(new(vl.Separator))
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
			err = fmt.Errorf("Do not view next operations: %v", nums)
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

/*
// view
if (BeginMenu("Geometry")) {
    m("View selected elements");
    m("View all elements");

    if (BeginMenu("Select by list")) {
        m("Nodes");
        m("Lines");
        m("Plates");
        // TODO: select by specific properties (material, lenght, ...)
        EndMenu();
    }
    Separator();

    Checkbox("Show Node number", &show[0]);
    Checkbox("Show Beam number", &show[1]);
    Checkbox("Show Plate number", &show[2]);
    Checkbox("Show Local coordinate", &show[3]);
    Checkbox("Show Node point", &show[4]);
    Checkbox("Show plate secondary border", &show_second_plate_border);
    Separator();

    Checkbox("Cursor select node", &cursor[0]);
    Checkbox("Cursor select beams", &cursor[1]);
    Checkbox("Cursor select plates", &cursor[2]);
    Separator();

    Separator();

    m("Statistic");

    EndMenu();
}
// reports
if (BeginMenu("Reports")) {
    if (BeginMenu("Covering plates from direction")) {
        m("+X");
        m("-X");
        m("+Y");
        m("-Y");
        m("+Z");
        m("-Z");
        m("by points");
        EndMenu();
    }

		add(TypModels,
			"Cylinder",
			"Sphere",
			"Cone",
			"Disk",
			"Cube",
			"Pipe branch",
			"Frame",
			"Beam-beam connection",
			"Column-beam connection",
			"Column-column connection",
		)

		add(Plugin,
			"Beam intersection",
			"Merge nodes",
			"Merge beams",
			"Merge plates",
			"Plate intersection",
			"Chamfer plates",
			"Fillet plates",
			"Explode plates",
			"Lines offset by direction",
			// "Split plates by lines",
			"Split lines by plates",
			// "Convert triangles to rectangles",
			// "Convert rectangles to triangles",
			"Plate bending",
			"Triangulation",
			"2D offset",
			"Twist",
			"Extrude",
			"Hole circle, square, rectangle on direction",
			"Cutoff",
			"Bend plates",
			"Stamping by point",
			"Stiffening rib",
			"Weld",
		)
*/

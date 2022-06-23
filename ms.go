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
	Move
	Scale
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
	case Move:
		return "Move"
	case Scale:
		return "Scale"
	case Plugin:
		return "Plugin"
	}
	return fmt.Sprintf("Undefined:%02d", g)
}

type Mesh interface {
	InsertNode(X, Y, Z string)
	SelectLines(single bool) (ids []uint)
	SelectNodes(single bool) (ids []uint)
	InsertNodeByDistance(line, distance string, pos uint)
	InsertNodeByProportional(line, proportional string, pos uint)
	InsertLineByNodeNumber(n1, n2 string)
	InsertTriangle3ByNodeNumber(n1, n2, n3 string)
	InsertQuadr4ByNodeNumber(n1, n2, n3, n4 string)
	InsertElementsByNodes(ids string, l2, t3, q4 bool)
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
		// coordinates
		w, gt := Input3Float(
			[3]string{"X", "Y", "Z"},
			[3]string{"meter", "meter", "meter"},
		)
		list.Add(w)
		// button
		var b vl.Button
		b.SetText("Insert")
		b.OnClick = func() {
			var vs [3]string
			for i := range vs {
				vs[i] = gt[i]()
			}
			m.InsertNode(gt[0](), gt[1](), gt[2]())
		}
		list.Add(&b)
		return &list
	}}, {
	Group: Add,
	Name:  "Node at the line by distance",
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			m.InsertNodeByDistance(sgt(), dgt(), rg.GetPos())
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Add,
	Name:  "Node at the line2 by proportional",
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			m.InsertNodeByProportional(sgt(), dgt(), rg.GetPos())
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			m.InsertLineByNodeNumber(bgt(), egt())
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			m.InsertTriangle3ByNodeNumber(n1gt(), n2gt(), n3gt())
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			m.InsertQuadr4ByNodeNumber(n1gt(), n2gt(), n3gt(), n4gt())
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
		bi.SetText("Insert")
		bi.OnClick = func() {
			if !(l2.Checked || tr3.Checked || q4.Checked) {
				return
			}
			m.InsertElementsByNodes(nsgt(), l2.Checked, tr3.Checked, q4.Checked)
		}
		list.Add(&bi)

		return &list
	}}, {
	Group: Split,
	Name:  "Line2 to equal parts",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Split,
	Name:  "Triangle3 to 3 Quadr4",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Split,
	Name:  "Triangle3 to 2 Triangle3 by side",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Split,
	Name:  "Quadr4 to 2 equal Quadr4 by side",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Split,
	Name:  "Quadr4 to 4 Triangle3",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Move,
	Name:  "Move by distance [X,Y,Z]",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Move,
	Name:  "Move from node n1 to node n2",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Move,
	Name:  "Move to specific plate",
	Part: func(m Mesh) (w vl.Widget) {
		// XOY
		// XOZ
		// YOZ
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
	}},
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

func Select(name string, single bool, selector func(single bool) []uint) (w vl.Widget, gettext func() string) {
	var l vl.ListH
	l.Add(vl.TextStatic(name))
	var id vl.Inputbox
	id.Filter(tf.UnsignedInteger)
	id.SetText("0")
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
				err = fmt.Errorf("Widget %02d is empty\n", i)
				return
			}
			r := part(m)
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
			err = fmt.Errorf("Do not view next operations: %v", nums)
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

type DebugMesh struct{}

func (DebugMesh) InsertNode(X, Y, Z string) {
	Debug = append(Debug, fmt.Sprintln("InsertNode: ", X, Y, Z))
}

func (DebugMesh) SelectLines(single bool) (ids []uint) {
	ids = []uint{314, 567}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectLines: ", ids))
	return
}

func (DebugMesh) SelectNodes(single bool) (ids []uint) {
	ids = []uint{1, 23, 444}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectNodes: ", ids))
	return
}

func (DebugMesh) InsertNodeByDistance(line, distance string, pos uint) {
	Debug = append(Debug,
		fmt.Sprintln("InsertNodeByDistance: ", line, distance, pos))
}

func (DebugMesh) InsertNodeByProportional(line, proportional string, pos uint) {
	Debug = append(Debug,
		fmt.Sprintln("InsertNodeByProportional: ", line, proportional, pos))
}

func (DebugMesh) InsertLineByNodeNumber(n1, n2 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertLineByNodeNumber: ", n1, n2))
}

func (DebugMesh) InsertTriangle3ByNodeNumber(n1, n2, n3 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertTriangle3ByNodeNumber: ", n1, n2, n3))
}

func (DebugMesh) InsertQuadr4ByNodeNumber(n1, n2, n3, n4 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertQuadr4ByNodeNumber: ", n1, n2, n3, n4))
}

func (DebugMesh) InsertElementsByNodes(ids string, l2, t3, q4 bool) {
	Debug = append(Debug,
		fmt.Sprintln("InsertElementsByNodes: ", ids, l2, t3, q4))
}

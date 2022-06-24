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
	Copy
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
	case Move:
		return "Move"
	case Copy:
		return "Copy"
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

type Mesh interface {
	InsertNode(X, Y, Z string)
	SelectLines(single bool) (ids []uint)
	SelectNodes(single bool) (ids []uint)
	SelectTriangles(single bool) (ids []uint)
	SelectQuadr4(single bool) (ids []uint)
	InsertNodeByDistance(line, distance string, pos uint)
	InsertNodeByProportional(line, proportional string, pos uint)
	InsertLineByNodeNumber(n1, n2 string)
	InsertTriangle3ByNodeNumber(n1, n2, n3 string)
	InsertQuadr4ByNodeNumber(n1, n2, n3, n4 string)
	InsertElementsByNodes(ids string, l2, t3, q4 bool)
	SplitLinesByRatio(lines, ratio string)
	SplitTri3To3Quadr4(tris string)
	SplitTri3To2Tri3(tris string, side uint)
	SplitQuadr4To2Quadr4(q4s string, side uint)
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
		b.SetText("Add")
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
			m.InsertNodeByDistance(sgt(), dgt(), rg.GetPos())
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
		bi.SetText("Add")
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
		bi.SetText("Add")
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
		bi.SetText("Add")
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
		bi.SetText("Add")
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
		var list vl.List
		ns, nsgt := Select("Select lines", Many, m.SelectLines)
		list.Add(ns)

		r, rgt := InputFloat("Ratio", "")
		list.Add(r)

		var bi vl.Button
		bi.SetText("Split")
		bi.OnClick = func() {
			m.SplitLinesByRatio(nsgt(), rgt())
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
	Name:  "Move to specific plane",
	Part: func(m Mesh) (w vl.Widget) {
		// XOY
		// XOZ
		// YOZ
		return vl.TextStatic("HOLD")
	}}, {
	Group: Move,
	Name:  "Rotate",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Move,
	Name:  "Mirror",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {

	Group: Copy,
	Name:  "Copy by distance [X,Y,Z]",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Copy from node n1 to node n2",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Copy to specific plane",
	Part: func(m Mesh) (w vl.Widget) {
		// XOY
		// XOZ
		// YOZ
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Copy by line path",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Translational repeat",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Circular repeat/Spiral",
	Part: func(m Mesh) (w vl.Widget) {
		return vl.TextStatic("HOLD")
	}}, {
	Group: Copy,
	Name:  "Mirror",
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

func (DebugMesh) SelectTriangles(single bool) (ids []uint) {
	ids = []uint{333, 555, 777}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectTriangles: ", ids))
	return
}

func (DebugMesh) SelectQuadr4(single bool) (ids []uint) {
	ids = []uint{1111, 2222, 3333, 4444}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectQuadr4: ", ids))
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

func (DebugMesh) SplitLinesByRatio(lines, ratio string) {
	Debug = append(Debug,
		fmt.Sprintln("SplitLinesByRatio: ", lines, ratio))
}

func (DebugMesh) SplitTri3To3Quadr4(tris string) {
	Debug = append(Debug,
		fmt.Sprintln("SplitTri3To3Quadr4: ", tris))
}

func (DebugMesh) SplitTri3To2Tri3(tris string, side uint) {
	Debug = append(Debug,
		fmt.Sprintln("SplitTri3To2Tri3: ", tris, side))
}

func (DebugMesh) SplitQuadr4To2Quadr4(q4s string, side uint) {
	Debug = append(Debug,
		fmt.Sprintln("SplitQuadr4To2Quadr4: ", q4s, side))
}

/*
void menu()
{
    static bool show[5] = { false, false, false, false, true }; // Node num, Beam num, Plates num, LocalCoord, Node point
    static bool cursor[3] = { true, true, true }; // Node, Beam, Plates
    static bool show_second_plate_border = false;

    if (ImGui::BeginMainMenuBar()) {
        // file
        if (ImGui::BeginMenu("File")) {
            if (ImGui::MenuItem("New")) { }
            if (ImGui::MenuItem("Open", "Ctrl+O")) { }
            if (ImGui::BeginMenu("Open Recent")) {
                ImGui::MenuItem("fish_hat.c");
                ImGui::MenuItem("fish_hat.inl");
                ImGui::MenuItem("fish_hat.h");
                if (ImGui::BeginMenu("More..")) {
                    ImGui::MenuItem("Hello");
                    ImGui::MenuItem("Sailor");
                    ImGui::EndMenu();
                }
                ImGui::EndMenu();
            }
            if (ImGui::MenuItem("Save", "Ctrl+S")) { }
            if (ImGui::MenuItem("Save As..")) { }
            ImGui::EndMenu();
        }
        // edit
        if (ImGui::BeginMenu("Edit")) {
            if (ImGui::MenuItem("Undo", "CTRL+Z")) { }
            if (ImGui::MenuItem("Redo", "CTRL+Y", false, false)) { }
            ImGui::Separator();
            if (ImGui::MenuItem("Cut", "CTRL+X")) { }
            if (ImGui::MenuItem("Copy", "CTRL+C")) { }
            if (ImGui::MenuItem("Paste", "CTRL+V")) { }
            ImGui::EndMenu();
        }
        // view
        if (ImGui::BeginMenu("Geometry")) {
            templorary_debug_menu("View selected elements");
            templorary_debug_menu("View all elements");
            ImGui::Separator();

            // TODO: plates edges

            if (ImGui::BeginMenu("Select All")) {
                templorary_debug_menu("Nodes");
                templorary_debug_menu("Beams");
                templorary_debug_menu("Plates");
                templorary_debug_menu("Geomery");
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Select inverse")) {
                templorary_debug_menu("Nodes");
                templorary_debug_menu("Lines");
                templorary_debug_menu("Plates");
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Select by list")) {
                templorary_debug_menu("Nodes");
                templorary_debug_menu("Lines");
                templorary_debug_menu("Plates");
                // TODO: select by specific properties (material, lenght, ...)
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Select beams parallel")) {
                templorary_debug_menu("X");
                templorary_debug_menu("Y");
                templorary_debug_menu("Z");
                templorary_debug_menu("Direction");
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Select plares parallel")) {
                templorary_debug_menu("XY");
                templorary_debug_menu("YZ");
                templorary_debug_menu("XZ");
                templorary_debug_menu("Plane");
                // TODO: select by specific properties (material, lenght, ...)
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Measurement")) {
                templorary_debug_menu("Distance between 2 nodes");
                templorary_debug_menu("Distance between 2 parallel beam");
                templorary_debug_menu("Distance between 2 parallel plates");
                ImGui::EndMenu();
            }
            ImGui::Separator();

            if (ImGui::BeginMenu("Standard views")) {
                templorary_debug_menu("+X");
                templorary_debug_menu("-X");
                templorary_debug_menu("+Y");
                templorary_debug_menu("-Y");
                templorary_debug_menu("+Z");
                templorary_debug_menu("-Z");
                // TODO: isometric views
                ImGui::EndMenu();
            }
            ImGui::Separator();

            ImGui::Checkbox("Show Node number", &show[0]);
            ImGui::Checkbox("Show Beam number", &show[1]);
            ImGui::Checkbox("Show Plate number", &show[2]);
            ImGui::Checkbox("Show Local coordinate", &show[3]);
            ImGui::Checkbox("Show Node point", &show[4]);
            ImGui::Checkbox("Show plate secondary border", &show_second_plate_border);
            ImGui::Separator();

            ImGui::Checkbox("Cursor select node", &cursor[0]);
            ImGui::Checkbox("Cursor select beams", &cursor[1]);
            ImGui::Checkbox("Cursor select plates", &cursor[2]);
            ImGui::Separator();

            templorary_debug_menu("Wireframe mode");
            templorary_debug_menu("Solid mode");
            ImGui::Separator();

            templorary_debug_menu("Statistic");

            ImGui::EndMenu();
        }
        // modify
        if (ImGui::BeginMenu("Modify")) {

            templorary_debug_menu("Create group");
            templorary_debug_menu("Create cross section");


            ImGui::EndMenu();
        }
        // reports
        if (ImGui::BeginMenu("Reports")) {
            if (ImGui::BeginMenu("Covering plates from direction")) {
                templorary_debug_menu("+X");
                templorary_debug_menu("-X");
                templorary_debug_menu("+Y");
                templorary_debug_menu("-Y");
                templorary_debug_menu("+Z");
                templorary_debug_menu("-Z");
                templorary_debug_menu("by points");
                ImGui::EndMenu();
            }
            if (ImGui::BeginMenu("Location")) {
                templorary_debug_menu("X");
                templorary_debug_menu("Y");
                templorary_debug_menu("Z");
                ImGui::EndMenu();
            }
            //
            ImGui::EndMenu();
        }

        //
        ImGui::EndMainMenuBar();
    }
}
*/

func add(name GroupId, parts ...string) {
	for i := range parts {
		Operations = append(Operations, Operation{
			Group: name,
			Name:  fmt.Sprintf("%s", parts[i]),
			Part: func(m Mesh) (w vl.Widget) {
				return vl.TextStatic("HOLD")
			},
		})
	}
}

func init() {
	add(Check,
		"Multiple structures",
		"Node duplicate",
		"Beam duplicate",
		"Plate duplicate",
		"Zero length beam",
		"Zero length plates",
		"Overlapping Collinear beams",
		"Empty loads",
		"Empty combinations",
		"Not connected nodes",
		"Unused supports",
		"Unused beam properties",
		"All ortho elements",
	)

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
		// "Translational Repeat",
		// "Circular Repeat/Spiral",
		// "Mirror",
		// "Move",
		// "Move with remesh/smooth",
		// "Rotate",
		"Beam intersection",
		"Merge nodes",
		"Merge beams",
		"Merge plates",
		"Plate intersection",
		"Chamfer plates",
		"Fillet plates",
		"Explode plates",
		"Lines offset by direction",
		// "Copy by line path",
		// "Plates smooth",
		// "Split plates by lines",
		"Split lines by plates",
		// "Split triangles by side 1",
		// "Split triangles by side 2",
		// "Split triangles by side 3",
		// "Split triangles by center point",
		// "Split triangles to 3 rectangles",
		// "Split rectangles to 4 triangles",
		// "Convert triangles to rectangles",
		// "Convert rectangles to triangles",
		// "Move points",
		// "Move from node to node",
		// "Move on plate",
		"Plate bending",
		"Triangulation",
		"2D offset",
		"Twist",
		"Extrude",
		// "Scale global",
		// "Scale +X",
		// "Scale -X",
		// "Scale +Y",
		// "Scale -Y",
		// "Scale +Z",
		// "Scale -Z",
		// "Scale by direction",
		// "Scale on cylinder system coordinate",
		"Hole circle, square, rectangle on direction",
		"Cutoff",
		"Bend plates",
		"Stamping by point",
		"Stiffening rib",
		"Weld",
	)
}

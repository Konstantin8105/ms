package ms

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Konstantin8105/tf"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
)

type GroupID uint8

const (
	File GroupID = iota
	Edit
	View
	Selection
	AddRemove
	Ignore
	Hide
	MoveCopy
	// 	TypModels
	// 	Check
	Plugin
	endGroup
)

func (g GroupID) String() string {
	switch g {
	case File:
		return "File"
	case Edit:
		return "Edit"
	case View:
		return "View"
	case Selection:
		return "Select"
	case AddRemove:
		return "Add/Modify/Remove"
	case Ignore:
		return "Ignore"
	case Hide:
		return "Hide"
	case MoveCopy:
		return "Move/Copy"
		// 	case Check:
		// 		return "Check"
		// 	case TypModels:
		// 		return "Typical models"
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
	// Store all operations
	// View 3D model

	// 2D model: axesymm
	// Convert to 2d

	PartPresent() (id uint)
	PartsName() (names []string)
	PartChange(id uint)
	PartNew(str string)
	PartRename(id uint, str string)
}

func defaultPartName(id int) string {
	if id == 0 {
		return "base model"
	}
	return fmt.Sprintf("part %02d", id)
}

func init() {
	group := File
	ops := []Operation{{
		Name: "Name of actual model/part",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var pre vl.Text
			var name vl.Text
			var lh vl.ListH
			lh.Add(&pre)
			lh.Add(&name)
			list.Add(&lh)

			update := func() {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						AddInfo("Safety ignore panic: %s", r)
					}
				}()
				id := m.PartPresent()
				ns := m.PartsName()
				part := ns[id]
				if part == "" {
					part = defaultPartName(int(id))
				}
				prefix := "Base model"
				if 0 < id {
					prefix = "Submodel"
				}
				pre.SetText(fmt.Sprintf("№%02d. %s", id, prefix))
				name.SetText(part)
			}

			go func() {
				for {
					<-time.After(1 * time.Second)
					update()
				}
			}()
			return &list
		}}, {
		Name: "Choose model/part",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var rg vl.RadioGroup
			list.Add(&rg)

			update := func() {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						AddInfo("Safety ignore panic: %s", r)
					}
				}()
				ns := m.PartsName()
				for i := range ns {
					if ns[i] != "" {
						continue
					}
					if i == 0 {
						ns[i] = "base model"
						continue
					}
					ns[i] = fmt.Sprintf("part %02d", i)
				}
				rg.SetText(ns)
			}

			go func() {
				for {
					<-time.After(1 * time.Second)
					update()
				}
			}()

			var b vl.Button
			b.SetText("Choose")
			b.OnClick = func() {
				pos := rg.GetPos()
				m.PartChange(pos)
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Create a new part",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var listH vl.ListH
			listH.Add(vl.TextStatic("Name:"))

			var name vl.Inputbox
			listH.Add(&name)

			list.Add(&listH)

			var b vl.Button
			b.SetText("Create")
			b.OnClick = func() {
				m.PartNew(name.GetText())
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Rename model/part",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var listH vl.ListH
			listH.Add(vl.TextStatic("Name:"))

			var name vl.Inputbox
			listH.Add(&name)

			list.Add(&listH)

			lastID := uint(0)

			update := func() {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						AddInfo("Safety ignore panic: %s", r)
					}
				}()
				id := m.PartPresent()
				if lastID != id {
					ns := m.PartsName()
					name.SetText(ns[id])
					lastID = id
					return
				}
				m.PartRename(id, name.GetText())
			}

			go func() {
				for {
					<-time.After(2 * time.Second)
					update()
				}
			}()

			return &list
		}},
	}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

type Editable interface {
	Undo()
	// Redo() //  The redo command reverses the undo or advances the buffer to a more recent state.
}

func init() {
	group := Edit
	ops := []Operation{{
		Name: "Undo",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.ListH

			list.Add(vl.TextStatic("Undo operation for erase last change of model"))

			var b vl.Button
			b.SetText("Undo")
			b.OnClick = func() {
				m.Undo()
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

type SView uint8

const (
	StandardViewXOYpos SView = iota
	StandardViewYOZpos
	StandardViewXOZpos
	StandardViewXOYneg
	StandardViewYOZneg
	StandardViewXOZneg
	endStandardView
)

func (s SView) String() string {
	switch s {
	case StandardViewXOYpos:
		return "+XOY"
	case StandardViewYOZpos:
		return "+YOZ"
	case StandardViewXOZpos:
		return "+XOZ"
	case StandardViewXOYneg:
		return "-XOY"
	case StandardViewYOZneg:
		return "-YOZ"
	case StandardViewXOZneg:
		return "-XOZ"
	}
	return "Undefined view"
}

type Viewable interface {
	// Wireframe mode
	// Solid mode
	StandardView(view SView)
	ColorEdge(isColor bool)
	ViewAll(centerCorrection bool)
	// Isometric views
	// View node number
	// View line number
	// View element number
}

func init() {
	group := View
	name := group.String()
	ops := []Operation{{
		Name: "Standard View",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var names []string
			for i := 0; i < int(endStandardView); i++ {
				names = append(names, SView(i).String())
			}

			var rg vl.RadioGroup
			rg.SetText(names)
			list.Add(&rg)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				pos := rg.GetPos()
				if uint(endStandardView) <= pos {
					return
				}
				m.StandardView(SView(pos))
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Color edges of elements",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var rg vl.RadioGroup
			rg.SetText([]string{"Normal colors", "Edge colors of elements"})
			list.Add(&rg)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				pos := rg.GetPos()
				if uint(endStandardView) <= pos {
					return
				}
				m.ColorEdge(pos == 1)
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

type AddRemovable interface {
	AddNode(X, Y, Z float64) (id uint)
	AddLineByNodeNumber(n1, n2 uint) (id uint)
	AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint, ok bool)
	AddModel(m Model)

	AddLeftCursor(lc LeftCursor)
	// RemoveLeftCursor(nodes, lines, tria bool)

	// TODO REMOVE AddQuadr4ByNodeNumber(n1, n2, n3, n4 string)
	// TODO REMOVE AddElementsByNodes(ids string, l2, t3, q4 bool)
	// AddGroup
	// AddCrossSections

	// Lines offset by direction
	// 2D offset
	// Offset inside/outside triangle/triangles
	// Split triangle inside triangle

	SplitLinesByDistance(lines []uint, distance float64, atBegin bool)
	SplitLinesByRatio(lines []uint, proportional float64, atBegin bool)
	SplitLinesByEqualParts(lines []uint, parts uint)
	SplitTri3To3Tri3(tris []uint)
	// SplitTri3To4Tri3(tris []uint)
	// TODO REMOVE SplitTri3To3Quadr4(tris string)
	// SplitTri3To2Tri3(tris string, side uint)
	// SplitQuadr4To2Quadr4(q4s string, side uint)
	// Quadr4 to 4 Triangle3
	// Quadr4 to 4 Quadr4
	// Triangles3, Quadrs4 by Lines2

	Intersection(nodes, elements []uint)
	// Intersections outside of FE

	// Engineering change coordinates with precision 0.5 mm = 0.0005 meter

	MergeNodes(minDistance float64)
	// MergeLines()
	// MergeTriangles()
	// MergeMesh()

	// Triangulation by nodes
	// Triangulation exist plates by area
	// Smooth mesh

	// Parabola
	// Hyperbola
	// Shell roof

	// Scale by ratio [sX,sY,sZ] and node
	// Scale by cylinder system coordinate
	// Scale by direction on 2 nodes

	// Chamfer plates
	// Fillet plates
	// Explode plates

	Remove(nodes, elements []uint)

	RemoveSameCoordinates()
	RemoveZeroLines()
	RemoveZeroTriangles()

	// Remove Nodes without elements

	GetCoords() []Coordinate
	GetElements() []Element
}

func init() {
	group := AddRemove
	ops := []Operation{{
		Name: "Add node by coordinate [X,Y,Z]",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			w, gt := Input3Float(
				"Coordinate:",
				[3]string{"X", "Y", "Z"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)

			var b vl.Button
			b.SetText("Add")
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
		Name: "Add line2 by nodes",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			b, bgt := Select("Select node1", Single, m.SelectNodes)
			list.Add(b)
			e, egt := Select("Select node2", Single, m.SelectNodes)
			list.Add(e)

			var bi vl.Button
			bi.SetText("Add")
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
		Name: "Add triangle3 by nodes",
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
		}}, {
		Name: "Add by left cursor button",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var names []string
			for i := 0; i < endLC; i++ {
				names = append(names, LeftCursor(i).String())
			}

			var rg vl.RadioGroup
			rg.SetText(names)
			list.Add(&rg)

			var b vl.Button
			b.SetText("Change")
			b.OnClick = func() {
				m.AddLeftCursor(LeftCursor(rg.GetPos()))
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Split line2 by distance from node",
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
		Name: "Split Line2 by ratio",
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
		Name: "Split Line2 to equal parts",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List
			ns, nsgt := Select("Select lines", Many, m.SelectLines)
			list.Add(ns)

			r, rgt := InputUnsigned("Amount parts", "")
			list.Add(r)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				parts, ok := rgt()
				if !ok {
					return
				}
				m.SplitLinesByEqualParts(nsgt(), parts)
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Split Triangle3 to 3 Triangle3",
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
		}}, {
		Name: "Intersection between nodes and elements",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			var b vl.Button
			b.SetText("Intersect")
			b.OnClick = func() {
				cs := coordgt()
				es := elgt()
				if len(cs) == 0 && len(es) == 0 {
					return
				}
				m.Intersection(cs, es)
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Merge nodes",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			d, dgt := InputFloat("Minimal distance", "meter")
			list.Add(d)

			var b vl.Button
			b.SetText("Merge")
			b.OnClick = func() {
				d, ok := dgt()
				if !ok {
					return
				}
				if d <= 0.0 {
					return
				}
				m.MergeNodes(d)
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Remove nodes with same coordinates",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveSameCoordinates()
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Remove lines with zero lenght",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveZeroLines()
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Remove triangles with zero area",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveZeroTriangles()
			}
			list.Add(&bi)

			return &list
		}}, {
		Name: "Remove selected",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				cs := coordgt()
				es := elgt()
				if len(cs) == 0 && len(es) == 0 {
					return
				}
				m.Remove(cs, es)
			}
			list.Add(&bi)

			return &list
		}},
	}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

type Ignorable interface {
	IgnoreModelElements(ids []uint)
	IsIgnore(elID uint) bool
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

type Hidable interface {
	Hide(coordinates, elements []uint)
	UnhideAll()
}

func init() {
	group := Hide
	ops := []Operation{{
		Name: "Hide",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			var b vl.Button
			b.SetText("Hide")
			b.OnClick = func() {
				els := elgt()
				cs := coordgt()
				if len(els) == 0 && len(cs) == 0 {
					return
				}
				m.Hide(cs, els)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Unhide all",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var b vl.Button
			b.SetText("Unhide all")
			b.OnClick = func() {
				m.UnhideAll()
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

type Direction uint8

const (
	DirX Direction = iota
	DirY
	DirZ
)

func (d Direction) String() string {
	switch d {
	case DirX:
		return "X direction"
	case DirY:
		return "Y direction"
	case DirZ:
		return "Z direction"
	}
	return "Undefined direction"
}

type Selectable interface {
	// TODO select deeper or only first iteration

	SelectLeftCursor(nodes, lines, tria bool)

	SelectNodes(single bool) (ids []uint)
	SelectLines(single bool) (ids []uint)
	SelectTriangles(single bool) (ids []uint)
	SelectElements(single bool) (ids []uint)

	InvertSelect(nodes, lines, triangles bool)

	SelectLinesOrtho(x, y, z bool)
	SelectLinesOnPlane(xoy, xoz, yoz bool)
	SelectLinesParallel(lines []uint)
	SelectLinesByLenght(more bool, lenght float64)
	SelectLinesCylindrical(node uint, radiant, conc bool, axe Direction)
	// SelectLinesSpherical(node uint, radiant, conc bool)

	// SelectPlatesWithAngle
	// SelectPlatesParallel// XY, YZ, XZ
	// SelectPlatesByArea
	// SelectPlatesByAngle

	// Select Snow/Wind elements
	// SelectByGroup

	SelectAll(nodes, lines, tria bool)
	DeselectAll()

	SelectScreen(from, to [2]int32)

	// Zoom
}

func init() {
	group := Selection
	name := group.String()
	ops := []Operation{{
		Name: "Left cursor selection",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var nodes vl.CheckBox
			nodes.SetText("Nodes")
			list.Add(&nodes)

			var lines vl.CheckBox
			lines.SetText("Lines")
			list.Add(&lines)

			var tris vl.CheckBox
			tris.SetText("Triangles")
			list.Add(&tris)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.SelectLeftCursor(nodes.Checked, lines.Checked, tris.Checked)
			}
			list.Add(&b)
			return &list
		}}, {
		Name: "Invert selection",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var nodes vl.CheckBox
			nodes.SetText("Nodes")
			list.Add(&nodes)

			var lines vl.CheckBox
			lines.SetText("Lines")
			list.Add(&lines)

			var tris vl.CheckBox
			tris.SetText("Triangles")
			list.Add(&tris)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.InvertSelect(nodes.Checked, lines.Checked, tris.Checked)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select ortho lines parallel axes X, Y, Z",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var x vl.CheckBox
			x.SetText("X")
			list.Add(&x)

			var y vl.CheckBox
			y.SetText("Y")
			list.Add(&y)

			var z vl.CheckBox
			z.SetText("Z")
			list.Add(&z)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.SelectLinesOrtho(x.Checked, y.Checked, z.Checked)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select lines on plane XOY, YOZ, XOZ",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var xoy vl.CheckBox
			xoy.SetText("XOY")
			list.Add(&xoy)

			var yoz vl.CheckBox
			yoz.SetText("YOZ")
			list.Add(&yoz)

			var xoz vl.CheckBox
			xoz.SetText("XOZ")
			list.Add(&xoz)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.SelectLinesOnPlane(xoy.Checked, yoz.Checked, xoz.Checked)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select lines parallel lines",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			lf, lfgt := Select("Lines", Many, m.SelectLines)
			list.Add(lf)

			var b vl.Button
			b.SetText("Select")
			b.OnClick = func() {
				ls := lfgt()
				if len(ls) == 0 {
					return
				}
				m.SelectLinesParallel(ls)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select lines by lenght",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var rg vl.RadioGroup
			rg.SetText([]string{"More", "Less"})
			list.Add(&rg)

			d, dgt := InputFloat("Lenght", "meter")
			list.Add(d)

			var b vl.Button
			b.SetText("Select")
			b.OnClick = func() {
				l, ok := dgt()
				if !ok {
					return
				}
				if l <= 0 {
					return
				}
				m.SelectLinesByLenght(rg.GetPos() == 0, l)
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select lines in cylinder system coordinate",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			nf, nfgt := Select("Node", Single, m.SelectNodes)
			list.Add(nf)

			list.Add(vl.TextStatic("Lines:"))

			var radiant vl.CheckBox
			radiant.SetText("Radiant")
			list.Add(&radiant)

			var conc vl.CheckBox
			conc.SetText("Concentric lines")
			list.Add(&conc)

			list.Add(vl.TextStatic("Direction:"))
			var drg vl.RadioGroup
			drg.SetText([]string{DirX.String(), DirY.String(), DirZ.String()})
			list.Add(&drg)

			var b vl.Button
			b.SetText("Select")
			b.OnClick = func() {
				n := nfgt()
				if len(n) != 1 {
					return
				}
				m.SelectLinesCylindrical(
					n[0], radiant.Checked, conc.Checked,
					Direction(drg.GetPos()))
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Deselect all",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var b vl.Button
			b.SetText("Deselect all")
			b.OnClick = func() {
				m.DeselectAll()
			}
			list.Add(&b)

			return &list
		}}, {
		Name: "Select all",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var nodes vl.CheckBox
			nodes.SetText("Nodes")
			list.Add(&nodes)

			var lines vl.CheckBox
			lines.SetText("Lines")
			list.Add(&lines)

			var tris vl.CheckBox
			tris.SetText("Triangles")
			list.Add(&tris)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.SelectAll(nodes.Checked, lines.Checked, tris.Checked)
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

type MoveCopyble interface {
	MoveCopyDistance(nodes, elements []uint, coordinate [3]float64,
		intermediantParts uint,
		copy, addLines, addTri bool)
	MoveCopyN1N2(nodes, elements []uint, from, to uint,
		intermediantParts uint,
		copy, addLines, addTri bool)
	//	MoveCopyOnPlane(nodes, elements []uint, coordinate [3]float64,
	//		plane Plane,
	//		intermediantParts uint,
	//		copy, addLines, addTri bool)
	//	Rotate(nodes, elements []uint, center [3]float64,
	//		angle float64,
	//		intermediantParts uint,
	//		distance float64, direction Direction, // Twist/Spiral
	//		copy, addLines, addTri bool)
	//	Mirror(nodes, elements []uint, center [3]float64,
	//		plane Plane,
	//		intermediantParts uint,
	//		copy, addLines, addTri bool)
	//	CopyByPath(nodes, elements []uint,
	//		path []uint, // lines path
	//		withRotation bool,
	//		copy, addLines, addTri bool) // Copy by line path
	// Bend
	// Loft
	// Translational repeat
	// Circular repeat/Spiral
}

func init() {
	ops := []Operation{{
		Name: "Move/Copy by distance [dX,dY,dZ]",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			w, gt := Input3Float(
				"Coordinate different:",
				[3]string{"dX", "dY", "dZ"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)

			list.Add(vl.TextStatic("\nIntermediant elements:"))

			var chLines vl.CheckBox
			chLines.SetText("Add intermediant lines")
			list.Add(&chLines)

			var chTriangles vl.CheckBox
			chTriangles.SetText("Add intermediant triangles")
			list.Add(&chTriangles)

			r, rgt := InputUnsigned("Amount intermediant parts", "")
			list.Add(r)

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
				parts, ok := rgt()
				if !ok {
					return
				}
				m.MoveCopyDistance(coordgt(), elgt(), vs, parts, rg.GetPos() == 1,
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

			list.Add(vl.TextStatic("\nIntermediant elements:"))

			var chLines vl.CheckBox
			chLines.SetText("Add intermediant lines")
			list.Add(&chLines)

			var chTriangles vl.CheckBox
			chTriangles.SetText("Add intermediant triangles")
			list.Add(&chTriangles)

			r, rgt := InputUnsigned("Amount intermediant parts", "")
			list.Add(r)

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
				parts, ok := rgt()
				if !ok {
					return
				}
				m.MoveCopyN1N2(coordgt(), elgt(), f[0], t[0], parts, rg.GetPos() == 1,
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

type Checkable interface {
	// CheckSingleStructure()     // Multiple structures
	// CheckDuplicateNodes()      // Node duplicate
	// CheckDuplicateLines()      // Beam duplicate
	// CheckDuplicateTriangles()  // Plate duplicate
	// CheckZeroLenghtLine()      // Zero length beam
	// CheckZeroLenghtTriangles() // Zero length plates
	// CheckElementsIndexes()     // Check FE on Indexes lenght
	// CheckTriangleOnOneLine()   // Plates not valid FE
	// CheckLinesOverlapping()    // Overlapping Collinear beams
	// CheckFreeNodes()           // Not connected nodes
	// CheckValidCoordinates()    // no NaN, infinite
	// Empty loads
	// Empty combinations
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
	DemoSpiral()
	// Cylinder
	// Sphere
	// Cone
	// Disk
	// Cube
	// Torus
	// Pipe branch
	// Frame with different I-height
	// Nano aerodymanic
	// Beam-beam connection
	// Column-beam connection
	// Column-column connection
	// Split plates by lines
	// Split lines by plates
	// Convert triangles to rectangles
	// Convert rectangles to triangles
	// Plate bending
	// Twist
	// Extrude
	// Hole circle, square, rectangle on direction
	// Cutoff
	// Bend plates
	// Stamping by point
	// Stiffening rib
	// Weld
	// Group
}

func init() {
	group := Plugin
	ops := []Operation{{
		Name: "Demo: spiral",
		Part: func(m Mesh) (w vl.Widget) {
			var list vl.List

			var b vl.Button
			b.SetText("Add")
			b.OnClick = func() {
				m.DemoSpiral()
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

type Mesh interface {
	Filable
	Viewable
	AddRemovable
	Editable
	Ignorable
	Hidable
	Selectable
	MoveCopyble
	Checkable
	Pluginable
	Measurementable
}

const (
	Single = true
	Many   = false
)

type Operation struct {
	Group GroupID
	Name  string
	Part  func(m Mesh) (w vl.Widget)
}

var Operations []Operation

func InputUnsigned(prefix, postfix string) (w vl.Widget, gettext func() (_ uint, ok bool)) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText("2")
	in.Filter(tf.UnsignedInteger)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, func() (uint, bool) {
		text := in.GetText()
		value, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(value), true
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
		clearValue(&str)
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			in.SetText(Default)
			return
		}
		return v, true
	}
}

func Input3Float(header string, prefix, postfix [3]string) (
	w vl.Widget,
	gettext [3]func() (float64, bool),
) {
	var list vl.List
	list.Add(vl.TextStatic(header))
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
	var (
		verticalList vl.List
		l1           vl.ListH
		coords       vl.Text
		b            vl.Button
		l2           vl.ListH
		els          vl.Text
	)

	verticalList.Add(vl.TextStatic("Select:"))

	l1.Add(vl.TextStatic("Nodes:"))
	const Default = "NONE"

	coords.SetLinesLimit(3)
	coords.SetText(Default)
	l1.Add(&coords)

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
	l1.Add(&b)
	verticalList.Add(&l1)

	l2.Add(vl.TextStatic("Elements:"))

	els.SetLinesLimit(3)
	els.SetText(Default)
	l2.Add(&els)

	l2.Add(vl.TextStatic(""))
	verticalList.Add(&l2)

	return &verticalList, func() (ids []uint) {
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

	id.SetLinesLimit(3)
	if single {
		id.SetLinesLimit(1)
	}
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

var (
	Info []string
)

func AddInfo(format string, args ...interface{}) {
	Info = append(Info, fmt.Sprintf(format, args...))
}

func PrintInfo() string {
	var out string
	for i := range Info {
		out += Info[i] + "\n"
	}
	return out
}

///////////////////////////////////////////////////////////////////////////////

type Tui struct {
	root vl.Widget

	mesh   Mesh
	Change func(*Opengl)
}

func (tui *Tui) Run(quit <-chan struct{}) error {
	action := make(chan func())
	defer func() {
		close(action)
	}()
	// TODO remove key close
	return vl.Run(tui.root, action, quit, tcell.KeyCtrlC)
}

func NewTui(mesh Mesh) (tui *Tui, err error) {
	tui = new(Tui)
	tui.mesh = mesh

	{
		// widgets amount
		AddInfo(fmt.Sprintf("Amount widgets: %d", len(Operations)))
	}
	var (
		scroll vl.Scroll
		list   vl.List
	)
	tui.root = &scroll
	scroll.Root = &list

	view := make([]bool, len(Operations))
	colHeader := make([]vl.CollapsingHeader, endGroup)
	for g := range colHeader {
		colHeader[g].SetText(GroupID(g).String())
		var sublist vl.List
		colHeader[g].Root = &sublist
		list.Add(&colHeader[g])
	}
	for g := range colHeader {
		for i := range Operations {
			if Operations[i].Group != GroupID(g) {
				continue
			}
			var c vl.CollapsingHeader
			c.SetText(Operations[i].Name)
			part := Operations[i].Part
			if part == nil {
				err = fmt.Errorf("widget %02d is empty: %#v", i, Operations[i])
				return
			}
			r := part(tui.mesh)
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

	{
		// Logger
		var logList vl.List

		var b vl.Button
		b.SetText("Clear log")
		b.OnClick = func() {
			Info = nil
		}
		logList.Add(&b)

		var t vl.Button
		t.SetText("Add time to log")
		t.OnClick = func() {
			AddInfo("Time: %v", time.Now())
		}
		logList.Add(&t)

		var txt vl.Text
		txt.SetText("Logger")

		update := func() {
			defer func() {
				if r := recover(); r != nil {
					// safety ignore panic
					AddInfo("Safety ignore panic: %s", r)
				}
			}()
			txt.SetText(strings.Join(Info, "\n"))
		}

		go func() {
			for {
				<-time.After(time.Millisecond * 500)
				update()
			}
		}()
		logList.Add(&txt)

		var log vl.CollapsingHeader
		log.SetText("Log")
		log.Root = &logList
		list.Add(&log)
	}

	return
}

func isOne(f func() []uint) (value uint, ok bool) {
	b := f()
	if len(b) != 1 {
		return
	}
	return b[0], true
}

func clearValue(str *string) {
	*str = strings.ReplaceAll(*str, "[", " ")
	*str = strings.ReplaceAll(*str, "]", " ")
}

func convertUint(str string) (ids []uint) {
	clearValue(&str)
	if str == "NONE" {
		return
	}
	fs := strings.Fields(str)
	for i := range fs {
		u, err := strconv.ParseUint(fs[i], 10, 64)
		if err != nil {
			AddInfo("convertUint error: %v", err)
			continue
		}
		ids = append(ids, uint(u))
	}
	return
}

// compress
//
//	Example:
//	from: 1 2 3 4 5 6 7
//	to  : 1 TO 7
//
//	from: 1 2 3 4 6 7
//	to  : 1 TO 4 6 7
//
//	from: 1 3 5 7
//	to  : 1 3 5 7
// func compress(ids []uint) (res string) {
// 	ids := uniqUint(ids)
// 	for i, id := range ids {
// 		res += fmt.Sprintf(" %d ", id)
// 	}
// 	return
// }

func uniqUint(ids []uint) (res []uint) {
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	res = make([]uint, 0, len(ids))
	for i, id := range ids {
		if i == 0 {
			res = append(res, id)
			continue
		}
		if res[len(res)-1] == ids[i] {
			continue
		}
		res = append(res, id)
	}
	return
}

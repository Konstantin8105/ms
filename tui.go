package ms

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/tf"
	"github.com/Konstantin8105/vl"
	"github.com/ncruces/zenity"
)

type GroupID uint8

const (
	File GroupID = iota
	Edit
	View
	PartOperations
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

// TODO metadata (add,change,select): thickness, local axe, section
// TODO Array by line, circular
// TODO distance between 2 points if selected both
// TODO betta angle for repeat rotate Copy
// TODO check copy node on distance

func (g GroupID) String() string {
	switch g {
	case File:
		return "File"
	case Edit:
		return "Edit"
	case View:
		return "View"
	case PartOperations:
		return "Part operations"
	case Selection:
		return "Select"
	case AddRemove:
		return "Add/Modify/Remove"
	case Ignore:
		return "Ignore"
	case Hide:
		return "Hide"
	case MoveCopy:
		return "Move/Copy/Mirror"
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
	Open(name string) error
	IsChangedModel() bool
	GetPresentFilename() (name string)
	Save() error
	SaveAs(filename string) error
	Close()
	// Store all operations
	// Import from gmsh
	// Export to gmsh
	// View 3D model

	// 2D model: axesymm
	// Convert to 2d

}

func init() {
	group := File
	ops := []Operation{{
		Name: "Open file",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			list.Add(vl.TextStatic("Present file"))

			var txt vl.Text
			list.Add(&txt)

			var b vl.Button
			b.SetText("Open file")
			b.OnClick = func() {
				// name of file
				name := m.GetPresentFilename()
				if name == "" {
					name = "Undefined." + FileExtension
				}
				// check is saved
				if m.IsChangedModel() {
					_, err := zenity.SelectFileSave(
						zenity.ConfirmOverwrite(),
						zenity.Filename(name),
						zenity.FileFilters{
							{"ms files", []string{"*." + FileExtension}, false},
						})
					if err != nil {
						return
					}
				}
				// select file
				name, err := zenity.SelectFile(
					zenity.Filename("."),
					zenity.Title("Select file"),
					zenity.FileFilters{
						{"ms files", []string{"*." + FileExtension}, false},
					})
				if err != nil {
					err = nil
					return
				}
				m.Open(name)
				str := m.GetPresentFilename()
				txt.SetText(str)
			}
			list.Add(&b)

			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Save",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var res vl.Text

			var b vl.Button
			b.SetText("Save")
			b.OnClick = func() {
				err := m.Save()
				if err != nil {
					res.SetText(fmt.Sprintf("%v", err))
					return
				}
				res.SetText("")
			}
			list.Add(&b)
			list.Add(&res)

			return &list, func() {
				res.SetText("")
			}
		}}, {
		Name: "Save As",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Save as ...")
			b.OnClick = func() {
				// name of file
				name := m.GetPresentFilename()
				if name == "" {
					name = "Undefined." + FileExtension
				}
				// save in new file
				name, err := zenity.SelectFileSave(
					zenity.ConfirmOverwrite(),
					zenity.Filename(name),
					zenity.FileFilters{
						{"ms files", []string{"*." + FileExtension}, false},
					})
				if err != nil {
					return
				}
				m.SaveAs(name)
			}
			list.Add(&b)

			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Close",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Close")
			b.OnClick = func() {
				// name of file
				name := m.GetPresentFilename()
				if name == "" {
					name = "Undefined." + FileExtension
				}
				// check is saved
				if m.IsChangedModel() {
					_, err := zenity.SelectFileSave(
						zenity.ConfirmOverwrite(),
						zenity.Filename(name),
						zenity.FileFilters{
							{"ms files", []string{"*." + FileExtension}, false},
						})
					if err != nil {
						return
					}
				}
				m.Close()
			}
			list.Add(&b)

			return &list, func() {
				// do nothing
			}
		}},
	}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

type Partable interface {
	PartPresent() (id uint)
	PartsName() (names []string)
	PartChange(id uint)
	PartNew(str string)
	PartRename(id uint, str string)
	// Delete part
}

func defaultPartName(id int) string {
	if id == 0 {
		return "base model"
	}
	return fmt.Sprintf("part %02d", id)
}

func init() {
	group := PartOperations
	ops := []Operation{{
		Name: "Name of actual model/part",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var pre vl.Text
			var name vl.Text
			var lh vl.ListH
			lh.Add(&pre)
			lh.Add(&name)
			list.Add(&lh)

			update := func() (fus bool) {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
					}
				}()
				id := m.PartPresent()
				ns := m.PartsName()
				if len(ns) <= int(id) {
					return
				}
				part := ns[id]
				if part == "" {
					part = defaultPartName(int(id))
				}
				prefix := "Base model"
				if 0 < id {
					prefix = "Submodel"
				}
				pre.SetText(fmt.Sprintf("%02d. %s", id, prefix))
				name.SetText(part)
				return false
			}

			go func() {
				for {
					time.Sleep(time.Second)
					if *closedApp {
						break
					}
					*actions <- update
				}
			}()
			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Choose model/part",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var rg vl.RadioGroup
			list.Add(&rg)

			change := func() {
				pos := rg.GetPos()
				m.PartChange(pos)
			}

			update := func() (fus bool) {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
					}
				}()
				ns := m.PartsName()
				id := m.PartPresent()
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
				rg.Clear()
				rg.AddText(ns...)
				rg.SetPos(id)
				change()
				return false
			}

			rg.OnChange(change)

			go func() {
				for {
					time.Sleep(time.Second)
					if *closedApp {
						break
					}
					*actions <- update
				}
			}()

			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Create a new part",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var listH vl.ListH
			listH.Add(vl.TextStatic("Name:"))

			var name vl.Inputbox
			listH.Add(&name)

			list.Add(&listH)

			var b vl.Button
			b.SetText("Create")
			b.OnClick = func() {
				n := name.GetText()
				if len(n) == 0 {
					return
				}
				m.PartNew(n)
			}
			list.Add(&b)
			return &list, func() {
				name.SetText("")
			}
		}}, {
		Name: "Rename model/part",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var listH vl.ListH
			listH.Add(vl.TextStatic("Name:"))

			var name vl.Inputbox
			listH.Add(&name)

			list.Add(&listH)

			lastID := uint(0)

			update := func() (fus bool) {
				defer func() {
					if r := recover(); r != nil {
						// safety ignore panic
						logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
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
				return false
			}

			go func() {
				for {
					time.Sleep(time.Second)
					if *closedApp {
						break
					}
					*actions <- update
				}
			}()

			return &list, func() {
				// do nothing
			}
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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.ListH

			list.Add(vl.TextStatic("Undo operation for erase last change of model"))

			var b vl.Button
			b.SetText("Undo")
			b.OnClick = func() {
				m.Undo()
			}
			list.Add(&b)

			return &list, func() {
				// do nothing
			}
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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var names []string
			for i := 0; i < int(endStandardView); i++ {
				names = append(names, SView(i).String())
			}

			var rg vl.RadioGroup
			rg.AddText(names...)
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
			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Color edges of elements",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var rg vl.RadioGroup
			rg.AddText([]string{"Normal colors", "Edge colors of elements"}...)
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
			return &list, func() {
				// do nothing
			}
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
	// LeftCursor split triangle by edge

	Intersection(nodes, elements []uint)
	// Intersections outside of FE

	// Engineering change coordinates with precision 0.5 mm = 0.0005 meter

	MergeNodes(minDistance float64)
	MergeLines(lines []uint)
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

	// Convert 3D to 2D
	// Convert 2D to 3D

	// Connect 2 lines - find intersections

	// Chamfer plates
	// Fillet plates
	// Explode plates

	// create section by plane

	Remove(nodes, elements []uint)

	RemoveNodesWithoutElements()
	RemoveSameCoordinates()
	RemoveZeroLines()
	RemoveZeroTriangles()

	GetCoordByID(id uint) (c gog.Point3d, ok bool)

	GetCoords() []Coordinate
	GetElements() []Element
}

func init() {
	group := AddRemove
	ops := []Operation{{
		Name: "Add node by coordinate [X,Y,Z]",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			w, gt, initf := Input3Float(
				"Coordinate:",
				[3]string{"X", "Y", "Z"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)

			var b vl.Button
			b.SetText("Add")
			b.OnClick = func() {
				vs, ok := gt()
				if !ok {
					return
				}
				m.AddNode(vs[0], vs[1], vs[2])
			}
			list.Add(&b)
			return &list, func() {
				initf()
			}
		}}, {
		Name: "Add line2 by nodes",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			b, bgt, initb := Select("Select node1", Single, m.GetSelectNodes)
			list.Add(b)
			e, egt, inite := Select("Select node2", Single, m.GetSelectNodes)
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

			return &list, func() {
				initb()
				inite()
			}
		}}, {
		Name: "Add triangle3 by nodes",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			n1, n1gt, init1 := Select("Select node1", Single, m.GetSelectNodes)
			list.Add(n1)
			n2, n2gt, init2 := Select("Select node2", Single, m.GetSelectNodes)
			list.Add(n2)
			n3, n3gt, init3 := Select("Select node3", Single, m.GetSelectNodes)
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

			return &list, func() {
				init1()
				init2()
				init3()
			}
		}}, {
		Name: "Add by left cursor button",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var names []string
			for i := 0; i < int(endLC); i++ {
				names = append(names, LeftCursor(i).String())
			}

			var rg vl.RadioGroup
			rg.AddText(names...)
			list.Add(&rg)

			var b vl.Button
			b.SetText("Change")
			b.OnClick = func() {
				m.AddLeftCursor(LeftCursor(rg.GetPos()))
			}
			list.Add(&b)
			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Split line2 by distance from node",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			s, sgt, inits := Select("Select lines", Many, m.GetSelectLines)
			list.Add(s)
			d, dgt, initd := InputFloat("Distance", "meter")
			list.Add(d)

			var rg vl.RadioGroup
			rg.AddText([]string{"from line begin", "from line end"}...)
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

			return &list, func() {
				inits()
				initd()
			}
		}}, {
		Name: "Split Line2 by ratio",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			s, sgt, inits := Select("Select line", Many, m.GetSelectLines)
			list.Add(s)
			d, dgt, initd := InputFloat("Ratio", "")
			list.Add(d)

			var rg vl.RadioGroup
			rg.AddText([]string{"from line begin", "from line end"}...)
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

			return &list, func() {
				inits()
				initd()
			}
		}}, {
		Name: "Split Line2 to equal parts",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			ns, nsgt, inits := Select("Select lines", Many, m.GetSelectLines)
			list.Add(ns)

			r, rgt, initr := InputUnsigned("Amount parts", "")
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

			return &list, func() {
				inits()
				initr()
			}
		}}, {
		Name: "Split Triangle3 to 3 Triangle3",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List
			ns, nsgt, initn := Select("Select triangles3", Many, m.GetSelectTriangles)
			list.Add(ns)

			var bi vl.Button
			bi.SetText("Split")
			bi.OnClick = func() {
				m.SplitTri3To3Tri3(nsgt())
			}
			list.Add(&bi)

			return &list, func() {
				initn()
			}
		}}, {
		Name: "Intersection between nodes and elements",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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
			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Merge nodes",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			d, dgt, initd := InputFloat("Minimal distance", "meter")
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
			return &list, func() {
				initd()
				//TODO clean
			}
		}}, {
		Name: "Merge lines",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			s, sgt, inits := Select("Select lines", Many, m.GetSelectLines)
			list.Add(s)

			var b vl.Button
			b.SetText("Merge")
			b.OnClick = func() {
				m.MergeLines(sgt())
			}
			list.Add(&b)
			return &list, func() {
				inits()
				//TODO clean
			}
		}}, {
		Name: "Remove nodes with same coordinates",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveSameCoordinates()
			}
			list.Add(&bi)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Remove nodes without connection to element",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveNodesWithoutElements()
			}
			list.Add(&bi)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Remove lines with zero lenght",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveZeroLines()
			}
			list.Add(&bi)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Remove triangles with zero area",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var bi vl.Button
			bi.SetText("Remove")
			bi.OnClick = func() {
				m.RemoveZeroTriangles()
			}
			list.Add(&bi)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Remove selected",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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

			return &list, func() {
				//TODO clean
			}
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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			elf, elfgt, inite := Select("Select elements", Many, m.GetSelectElements)
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

			return &list, func() {
				inite()
				//TODO clean
			}
		}}, {
		Name: "Clear ignoring elements",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Clear")
			b.OnClick = func() {
				m.Unignore()
			}
			list.Add(&b)

			return &list, func() {
				//TODO clean
			}
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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Show only selected",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Show only selected")
			b.OnClick = func() {
				m.InvertSelect(true, true, true, true)
				ns := m.GetSelectNodes(Many)
				es := m.GetSelectElements(Many)
				m.DeselectAll()
				m.Hide(ns, es)
			}
			list.Add(&b)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Unhide all",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Unhide all")
			b.OnClick = func() {
				m.UnhideAll()
			}
			list.Add(&b)

			return &list, func() {
				//TODO clean
			}
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

type Plane uint8

const (
	PlaneXOY Plane = iota
	PlaneXOZ
	PlaneYOZ
)

func (p Plane) String() string {
	switch p {
	case PlaneXOY:
		return "XOY"
	case PlaneXOZ:
		return "XOZ"
	case PlaneYOZ:
		return "YOZ"
	}
	return "Undefined plane"
}

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

	SelectLeftCursor(nodes bool, elements []bool)

	GetSelectNodes(single bool) (ids []uint)
	GetSelectLines(single bool) (ids []uint)
	GetSelectTriangles(single bool) (ids []uint)
	GetSelectElements(single bool) (ids []uint)

	InvertSelect(nodes, lines, triangles, quards bool)

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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var nodes vl.CheckBox
			nodes.SetText("Nodes")
			list.Add(&nodes)

			els := make([]vl.CheckBox, lastElement)
			for e := Line2; e < lastElement; e++ {
				els[int(e)].SetText(ElType(e).String())
				list.Add(&els[int(e)])
			}

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				ls := make([]bool, len(els))
				for i := range els {
					ls[i] = els[i].Checked
				}
				m.SelectLeftCursor(nodes.Checked, ls)
			}
			list.Add(&b)
			return &list, func() {
				// do nothing
			}
		}}, {
		Name: "Invert selection",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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

			var quards vl.CheckBox
			quards.SetText("Triangles")
			list.Add(&quards)

			var b vl.Button
			b.SetText(name)
			b.OnClick = func() {
				m.InvertSelect(nodes.Checked, lines.Checked, tris.Checked, quards.Checked)
			}
			list.Add(&b)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Select ortho lines parallel axes X, Y, Z",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Select lines on plane XOY, YOZ, XOZ",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Select lines parallel lines",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			lf, lfgt, initl := Select("Lines", Many, m.GetSelectLines)
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

			return &list, func() {
				initl()
				//TODO clean
			}
		}}, {
		Name: "Select lines by lenght",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var rg vl.RadioGroup
			rg.AddText([]string{"More", "Less"}...)
			list.Add(&rg)

			d, dgt, initd := InputFloat("Lenght", "meter")
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

			return &list, func() {
				initd()
				//TODO clean
			}
		}}, {
		Name: "Select lines in cylinder system coordinate",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			nf, nfgt, initn := Select("Node", Single, m.GetSelectNodes)
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
			drg.AddText([]string{DirX.String(), DirY.String(), DirZ.String()}...)
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

			return &list, func() {
				initn()
				//TODO clean
			}
		}}, {
		Name: "Deselect all",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			var b vl.Button
			b.SetText("Deselect all")
			b.OnClick = func() {
				m.DeselectAll()
			}
			list.Add(&b)

			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Select all",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
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
			return &list, func() {
				//TODO clean
			}
		}},
	}
	for i := range ops {
		ops[i].Group = group
	}
	Operations = append(Operations, ops...)
}

// diffCoordinate is different of coordinate
//
//	0 - dX
//	1 - dY
//	2 - dZ
//	3 - angle around X
//	4 - angle around Y
//	5 - angle around Z
type diffCoordinate [6]float64

type MoveCopyble interface {
	Move(nodes, elements []uint,
		basePoint [3]float64,
		path diffCoordinate)
	Copy(nodes, elements []uint,
		basePoint [3]float64,
		paths []diffCoordinate,
		addLines, addTri bool)
	Mirror(nodes, elements []uint,
		basePoint [3][3]float64,
		copy bool,
		addLines, addTri bool)
	// TODO do not copy lines Collinear on copy direction
	//	MoveCopyOnPlane(nodes, elements []uint, coordinate [3]float64,
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
		Name: "Move/Rotate",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			type path struct {
				w    vl.Widget
				getC func() (basePoint [3]float64, dc diffCoordinate, ok bool)
			}
			var paths []path
			var inits []func()
			{ // from node to node
				var ch vl.CollapsingHeader
				ch.SetText("Move from node to node:")

				var list vl.List
				nf, nfgt, init1 := Select("From node", Single, m.GetSelectNodes)
				list.Add(nf)
				inits = append(inits, init1)
				nt, ntgt, init2 := Select("To node", Single, m.GetSelectNodes)
				list.Add(nt)
				inits = append(inits, init2)

				ch.Root = &list
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dc diffCoordinate, ok bool) {
						f := nfgt()
						if len(f) != 1 {
							return
						}
						t := ntgt()
						if len(t) != 1 {
							return
						}
						fc, ok := m.GetCoordByID(f[0])
						if !ok {
							return
						}
						tc, ok := m.GetCoordByID(t[0])
						if !ok {
							return
						}
						ok = true
						basePoint = fc
						dc[0] = tc[0] - fc[0]
						dc[1] = tc[1] - fc[1]
						dc[2] = tc[2] - fc[2]
						return
					},
				})
			}
			{ // different coordinates
				var ch vl.CollapsingHeader
				ch.SetText("Move by coordinate different [dX,dY,dZ]:")

				w, gt, initw := Input3Float(
					"",
					[3]string{"dX", "dY", "dZ"},
					[3]string{"meter", "meter", "meter"},
				)
				inits = append(inits, initw)

				ch.Root = w
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dc diffCoordinate, ok bool) {
						vs, vok := gt()
						if !vok {
							return
						}
						ok = true
						dc[0] = vs[0]
						dc[1] = vs[1]
						dc[2] = vs[2]
						return
					},
				})
			}
			{ // rotate
				var ch vl.CollapsingHeader
				ch.SetText("Rotate around node:")

				var list vl.List

				nt, ntgt, initn := Select("Center of rotation", Single, m.GetSelectNodes)
				list.Add(nt)
				inits = append(inits, initn)

				list.Add(new(vl.Separator))
				w, gt, initi := Input3Float(
					"Angle of rotation",
					[3]string{"around axe X", "around axe Y", "around axe Z"},
					[3]string{"degree", "degree", "degree"},
				)
				list.Add(w)
				inits = append(inits, initi)

				ch.Root = &list
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dc diffCoordinate, ok bool) {
						n := ntgt()
						if len(n) != 1 {
							return
						}
						as, aok := gt()
						if !aok {
							return
						}
						c, ok := m.GetCoordByID(n[0])
						if !ok {
							return
						}
						ok = true
						basePoint = c
						dc[3] = as[0]
						dc[4] = as[1]
						dc[5] = as[2]
						return
					},
				})
			}
			// radio group for paths
			list.Add(new(vl.Separator))
			list.Add(vl.TextStatic("Choose parameters:"))
			var param vl.RadioGroup
			for i := range paths {
				param.Add(paths[i].w)
			}
			list.Add(&param)
			// operation
			list.Add(new(vl.Separator))
			var b vl.Button
			b.SetText("Move/Rotate")
			b.OnClick = func() {
				pos := param.GetPos()
				bp, p, ok := paths[pos].getC()
				if !ok {
					return
				}
				m.Move(coordgt(), elgt(), bp, p)
			}
			list.Add(&b)
			return &list, func() {
				//TODO clean
			}
		}}, {
		Name: "Translational/Circular repeat/Spiral",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			type path struct {
				w    vl.Widget
				getC func() (basePoint [3]float64, dcs []diffCoordinate, ok bool)
			}
			var paths []path
			var inits []func()
			{ // from node to node with equal parts
				var ch vl.CollapsingHeader
				ch.SetText("Copy from node to node with equal parts:")

				var list vl.List
				nf, nfgt, initn1 := Select("From node", Single, m.GetSelectNodes)
				list.Add(nf)
				inits = append(inits, initn1)
				nt, ntgt, initn2 := Select("To node", Single, m.GetSelectNodes)
				list.Add(nt)
				inits = append(inits, initn2)

				parts, partsgt, initp := InputUnsigned("Amount equal parts", "items")
				list.Add(parts)
				inits = append(inits, initp)

				ch.Root = &list
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dcs []diffCoordinate, ok bool) {
						f := nfgt()
						if len(f) != 1 {
							return
						}
						t := ntgt()
						if len(t) != 1 {
							return
						}
						fc, ok := m.GetCoordByID(f[0])
						if !ok {
							return
						}
						tc, ok := m.GetCoordByID(t[0])
						if !ok {
							return
						}
						parts, pok := partsgt()
						if !pok {
							return
						}
						ok = true
						basePoint = fc
						for i := 0; i <= int(parts); i++ {
							dcs = append(dcs, diffCoordinate([6]float64{
								(tc[0] - fc[0]) / float64(parts+1),
								(tc[1] - fc[1]) / float64(parts+1),
								(tc[2] - fc[2]) / float64(parts+1),
							}))
						}
						return
					},
				})
			}
			{ // different coordinates

				var ch vl.CollapsingHeader
				ch.SetText("Copy by coordinate different [dX,dY,dZ] with equal parts:")

				var list vl.List
				w, gt, initw := Input3Float(
					"",
					[3]string{"dX", "dY", "dZ"},
					[3]string{"meter", "meter", "meter"},
				)
				list.Add(w)
				inits = append(inits, initw)

				parts, partsgt, initp := InputUnsigned("Amount equal parts", "items")
				list.Add(parts)
				inits = append(inits, initp)

				ch.Root = &list
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dcs []diffCoordinate, ok bool) {
						vs, vok := gt()
						if !vok {
							return
						}
						parts, pok := partsgt()
						if !pok {
							return
						}
						ok = true
						for i := 0; i <= int(parts); i++ {
							dcs = append(dcs, diffCoordinate([6]float64{
								vs[0] / float64(parts+1),
								vs[1] / float64(parts+1),
								vs[2] / float64(parts+1),
							}))
						}
						return
					},
				})
			}
			{ // Translational repeat
				var ch vl.CollapsingHeader
				ch.SetText("Triangulation repeat:")

				var list vl.List

				list.Add(vl.TextStatic("Direction of repeat:"))
				var dir vl.RadioGroup
				dir.AddText([]string{
					DirX.String(),
					DirY.String(),
					DirZ.String(),
				}...)
				list.Add(&dir)

				list.Add(new(vl.Separator))
				list.Add(vl.TextStatic("List of distances:"))

				var distances vl.List
				var dgt []func() (_ float64, ok bool)
				var bAdd vl.Button
				bAdd.SetText("Add distance")
				bAdd.OnClick = func() {
					w, gt, _ := InputFloat(
						fmt.Sprintf("%d", distances.Size()+1),
						"meter",
					)
					distances.Add(w)
					dgt = append(dgt, gt)
				}
				var bClear vl.Button
				bClear.SetText("Clear")
				bClear.OnClick = func() {
					distances.Clear()
					dgt = nil
					bAdd.OnClick() // one empty distance
				}
				bClear.OnClick() // default clear
				list.Add(&distances)
				{
					var h vl.ListH
					h.Add(&bAdd)
					h.Add(&bClear)
					list.Add(&h)
				}

				ch.Root = &list

				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dcs []diffCoordinate, ok bool) {
						ind := dir.GetPos()
						var vs []float64
						for i := range dgt {
							v, vok := dgt[i]()
							if !vok {
								return
							}
							vs = append(vs, v)
						}
						ok = true
						for i := range vs {
							var dc diffCoordinate
							dc[ind] = vs[i]
							dcs = append(dcs, dc)
						}
						return
					},
				})
			}
			{ // rotate
				var ch vl.CollapsingHeader
				ch.SetText("Circular repeat around node:")

				var list vl.List

				nt, ntgt, initn := Select("Center of rotation", Single, m.GetSelectNodes)
				list.Add(nt)
				inits = append(inits, initn)

				list.Add(new(vl.Separator))
				w, gt, initw := Input3Float(
					"Angle of rotation",
					[3]string{"around axe X", "around axe Y", "around axe Z"},
					[3]string{"degree", "degree", "degree"},
				)
				list.Add(w)
				inits = append(inits, initw)

				parts, partsgt, initp := InputUnsigned("Amount equal parts", "items")
				list.Add(parts)
				inits = append(inits, initp)

				ch.Root = &list
				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3]float64, dcs []diffCoordinate, ok bool) {
						n := ntgt()
						if len(n) != 1 {
							return
						}
						as, aok := gt()
						if !aok {
							return
						}
						ok = true
						c, ok := m.GetCoordByID(n[0])
						if !ok {
							return
						}
						parts, pok := partsgt()
						if !pok {
							return
						}
						basePoint = c
						for i := 0; i <= int(parts); i++ {
							dcs = append(dcs, diffCoordinate([6]float64{
								0, 0, 0,
								as[0] / float64(parts+1),
								as[1] / float64(parts+1),
								as[2] / float64(parts+1),
							}))
						}
						return
					},
				})
			}

			// TODO spiral

			// radio group for paths
			list.Add(new(vl.Separator))
			list.Add(vl.TextStatic("Choose parameters:"))
			var param vl.RadioGroup
			for i := range paths {
				param.Add(paths[i].w)
			}
			list.Add(&param)
			// intermediant
			var lines vl.CheckBox
			lines.SetText("Add intermediant lines")
			list.Add(&lines)
			var tris vl.CheckBox
			tris.SetText("Add intermediant triangles")
			list.Add(&tris)
			// operation
			list.Add(new(vl.Separator))
			var b vl.Button
			b.SetText("Repeat")
			b.OnClick = func() {
				pos := param.GetPos()
				bp, ps, ok := paths[pos].getC()
				if !ok {
					return
				}
				m.Copy(coordgt(), elgt(), bp, ps, lines.Checked, tris.Checked)
			}
			list.Add(&b)
			return &list, func() {
				for i := range inits {
					if f := inits[i]; f != nil {
						f()
					}
				}
				//TODO clean
			}
		}}, {
		Name: "Mirror",
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			ns, coordgt, elgt := SelectAll(m)
			list.Add(ns)

			type path struct {
				w    vl.Widget
				getC func() (basePoint [3][3]float64, ok bool)
			}
			var paths []path
			var inits []func()
			{
				var ch vl.CollapsingHeader
				ch.SetText("Node and plane:")

				var list vl.List

				n, ngt, initn := Select("Select node", Single, m.GetSelectNodes)
				list.Add(n)
				inits = append(inits, initn)

				list.Add(new(vl.Separator))
				list.Add(vl.TextStatic("Choose plane:"))
				var plane vl.RadioGroup
				plane.AddText([]string{
					PlaneXOY.String(),
					PlaneXOZ.String(),
					PlaneYOZ.String(),
				}...)

				list.Add(&plane)
				ch.Root = &list

				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3][3]float64, ok bool) {
						ns := ngt()
						if len(ns) != 1 {
							return
						}
						coord, ok := m.GetCoordByID(ns[0])
						if !ok {
							return
						}
						switch Plane(plane.GetPos()) {
						case PlaneXOY:
							basePoint[0] = coord
							basePoint[1] = [3]float64{
								coord[0] + 1, coord[1], coord[2],
							}
							basePoint[2] = [3]float64{
								coord[0], coord[1] + 1, coord[2],
							}
							ok = true
						case PlaneXOZ:
							basePoint[0] = coord
							basePoint[1] = [3]float64{
								coord[0] + 1, coord[1], coord[2],
							}
							basePoint[2] = [3]float64{
								coord[0], coord[1], coord[2] + 1,
							}
							ok = true
						case PlaneYOZ:
							basePoint[0] = coord
							basePoint[1] = [3]float64{
								coord[0], coord[1] + 1, coord[2],
							}
							basePoint[2] = [3]float64{
								coord[0], coord[1], coord[2] + 1,
							}
							ok = true
						}
						return
					},
				})
			}
			{
				var ch vl.CollapsingHeader
				ch.SetText("Plane by 3 points:")

				var list vl.List

				n0, ngt0, init0 := Select("Select node 1:", Single, m.GetSelectNodes)
				list.Add(n0)
				inits = append(inits, init0)
				n1, ngt1, init1 := Select("Select node 2:", Single, m.GetSelectNodes)
				list.Add(n1)
				inits = append(inits, init1)
				n2, ngt2, init2 := Select("Select node 3:", Single, m.GetSelectNodes)
				list.Add(n2)
				inits = append(inits, init2)

				ch.Root = &list

				paths = append(paths, path{
					w: &ch,
					getC: func() (basePoint [3][3]float64, ok bool) {
						n0 := ngt0()
						if len(n0) != 1 {
							return
						}
						c0, ok := m.GetCoordByID(n0[0])
						if !ok {
							return
						}

						n1 := ngt1()
						if len(n1) != 1 {
							return
						}
						c1, ok := m.GetCoordByID(n1[0])
						if !ok {
							return
						}

						n2 := ngt2()
						if len(n2) != 1 {
							return
						}
						c2, ok := m.GetCoordByID(n2[0])
						if !ok {
							return
						}

						basePoint[0] = c0
						basePoint[1] = c1
						basePoint[2] = c2
						ok = true
						return
					},
				})
			}

			// radio group for paths
			list.Add(new(vl.Separator))
			list.Add(vl.TextStatic("Choose parameters:"))
			var param vl.RadioGroup
			for i := range paths {
				param.Add(paths[i].w)
			}
			list.Add(&param)
			// intermediant

			// copy or move mirror
			var mir vl.RadioGroup
			mir.Add(vl.TextStatic("Move - no copy"))

			var cop vl.List
			var lines vl.CheckBox
			lines.SetText("Add intermediant lines")
			cop.Add(&lines)
			var tris vl.CheckBox
			tris.SetText("Add intermediant triangles")
			cop.Add(&tris)
			mir.Add(&cop)

			list.Add(&mir)

			// operation
			list.Add(new(vl.Separator))
			var b vl.Button
			b.SetText("Repeat")
			b.OnClick = func() {
				pos := param.GetPos()

				bp, ok := paths[pos].getC()
				if !ok {
					return
				}

				m.Mirror(coordgt(), elgt(),
					bp,
					mir.GetPos() == 1,
					lines.Checked, tris.Checked)
			}
			list.Add(&b)
			return &list, func() {
				for i := range inits {
					if f := inits[i]; f != nil {
						f()
					}
				}
				//TODO clean
			}
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
	DemoSpiral(levels uint)
	// Tubesheet inline and staggered
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
		Part: func(m Mesh, actions *chan ds.Action, closedApp *bool) (w vl.Widget, f func()) {
			var list vl.List

			r, rgt, initr := InputUnsigned("Amount levels", "")
			list.Add(r)

			var b vl.Button
			b.SetText("Add")
			b.OnClick = func() {
				n, ok := rgt()
				if !ok {
					return
				}
				if n < 1 {
					return
				}
				m.DemoSpiral(n)
			}
			list.Add(&b)
			return &list, func() {
				initr()
				//TODO clean
			}
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
	Partable
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
	Part  func(
		m Mesh,
		actions *chan ds.Action,
		closedApp *bool,
	) (
		w vl.Widget,
		initialization func(), // initialization values after open another file
	)
}

var Operations []Operation

func InputUnsigned(prefix, postfix string) (
	w vl.Widget,
	gettext func() (_ uint, ok bool),
	initialization func(),
) {
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
		}, func() {
			in.SetText("2")
		}
}

func InputFloat(prefix, postfix string) (
	w vl.Widget,
	gettext func() (_ float64, ok bool),
	initialization func(),
) {
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
		}, func() {
			in.SetText("")
		}
}

func Input3Float(header string, prefix, postfix [3]string) (
	w vl.Widget,
	gettext func() (_ [3]float64, ok bool),
	initialization func(),
) {
	var list vl.List
	list.Add(vl.TextStatic(header))
	var gt [3]func() (_ float64, ok bool)
	var inits [3]func()
	for i := 0; i < 3; i++ {
		w, wgt, init := InputFloat(prefix[i], postfix[i])
		list.Add(w)
		gt[i] = wgt
		inits[i] = init
	}
	return &list, func() (vs [3]float64, ok bool) {
			for i := 0; i < 3; i++ {
				v, vok := gt[i]()
				if !vok {
					return
				}
				vs[i] = v
			}
			ok = true
			return
		}, func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
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
		coordinates := m.GetSelectNodes(Many)
		elements := m.GetSelectElements(Many)
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
	initialization func(),
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
		}, func() {
			id.SetText("")
		}
}

///////////////////////////////////////////////////////////////////////////////

// var (
// 	Info []string
// )
//
// func ResetInfo() {
// 	Info = nil
// }
//
// func logger.Printf(format string, args ...interface{}) {
// 	Info = append(Info, fmt.Sprintf(format, args...))
// }
//
// func PrintInfo() string {
// 	var out string
// 	for i := range Info {
// 		out += Info[i] + "\n"
// 	}
// 	return out
// }

///////////////////////////////////////////////////////////////////////////////

// type Tui struct {
// 	root vl.Widget
//
// 	mesh Mesh
// 	// actions chan ds.Action
// 	Change func(*Opengl)
// 	quit   bool
// }

// func (tui *Tui) Run(quit <-chan struct{}) error {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			// safety ignore panic
// 			logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
// 		}
// 	}()
// 	defer func() {
// 		tui.quit = true
// 		// TODO: time.Sleep(5 * time.Second)
// 		// TODO: close(tui.actions)
// 	}()
// 	// TODO remove key close
// 	return vl.Run(tui.root, tui.actions, quit, tcell.KeyCtrlC)
// }

func NewTui(mesh Mesh, closedApp *bool, actions *chan ds.Action) (tui vl.Widget, initialization func(), err error) {
	defer func() {
		if r := recover(); r != nil {
			// safety ignore panic
			logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
		}
	}()
	// tui = new(Tui)
	// tui.mesh = mesh
	// tui.actions = make(chan ds.Action)

	{
		// widgets amount
		logger.Printf(fmt.Sprintf("Amount widgets: %d", len(Operations)))
	}
	var (
		scroll vl.Scroll
		list   vl.List
	)
	// tui.root = &scroll
	tui = &scroll
	scroll.Root = &list

	view := make([]bool, len(Operations))
	colHeader := make([]vl.CollapsingHeader, endGroup)
	for g := range colHeader {
		colHeader[g].SetText(GroupID(g).String())
		var sublist vl.List
		colHeader[g].Root = &sublist
		list.Add(&colHeader[g])
	}
	var is []func()
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
			// r := part(tui.mesh, tui.actions)
			r, inits := part(mesh, actions, closedApp)
			c.Root = r
			colHeader[g].Root.(*vl.List).Add(&c)
			view[i] = true
			is = append(is, inits)
		}
	}
	initialization = func() {
		for i := range is {
			if f := is[i]; f != nil {
				f()
			}
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

	// 	{
	// 		// Logger
	// 		var logList vl.List
	//
	// 		var b vl.Button
	// 		b.SetText("Clear log")
	// 		b.OnClick = func() {
	// 			ResetInfo()
	// 		}
	// 		logList.Add(&b)
	//
	// 		var t vl.Button
	// 		t.SetText("Add time to log")
	// 		t.OnClick = func() {
	// 			logger.Printf("Time: %v", time.Now())
	// 		}
	// 		logList.Add(&t)
	//
	// 		var txt vl.Text
	// 		txt.SetText("Logger")
	//
	// 		// 		update := func() {
	// 		// 			defer func() {
	// 		// 				if r := recover(); r != nil {
	// 		// 					// safety ignore panic
	// 		// 					logger.Printf("Safety ignore panic: %s\n%v", r, string(debug.Stack()))
	// 		// 				}
	// 		// 			}()
	// 		// 			txt.SetText(strings.Join(Info, "\n"))
	// 		// 		}
	// 		//
	// 		// 		go func() {
	// 		// 			for {
	// 		// 				time.Sleep(time.Millisecond * 500)
	// 		// 				if tui.quit {
	// 		// 					return
	// 		// 				}
	// 		// 				// tui.actions <- update
	// 		// 				actions <- update
	// 		// 			}
	// 		// 		}()
	// 		logList.Add(&txt)
	//
	// 		var log vl.CollapsingHeader
	// 		log.SetText("Log")
	// 		log.Root = &logList
	// 		list.Add(&log)
	// 	}

	return
}

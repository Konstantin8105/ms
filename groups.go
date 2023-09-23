package ms

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/vl"
)

// max uint16: 65535
type GroupIndex uint16

const (
	NamedIndex        GroupIndex = 80
	NamedListIndex               = 100
	NodeSupportsIndex            = 1000
	MetaIndex                    = 10000
)

func (gi GroupIndex) String() string {
	switch gi {
	case NamedIndex:
		return "Comments"
	case NamedListIndex:
		return "Named list"
	case NodeSupportsIndex:
		return "Node supports"
	case MetaIndex:
		return "Meta"
	}
	return fmt.Sprintf("Undefined name of group %d", uint16(gi))
}

func (gi GroupIndex) newInstance() (_ Group, ok bool) {
	switch gi {
	case NamedIndex:
		return new(Named), true
	case NamedListIndex:
		return new(NamedList), true
	case NodeSupportsIndex:
		return new(NodeSupports), true
	case MetaIndex:
		return new(Meta), true
	}
	return nil, false
}

//	type Model2 struct {
//		// ...
//		Data []struct {
//			Id          GroupIndex
//			bs []byte
//			group       Group
//		}
//		// ...
//	}
type Group interface {
	GetId() GroupIndex
	String() string                                                 // return short name
	Update(updating func(nodes, elements *[]uint))                  // update nodes, elements indexes
	GetWidget(mm Mesh, updateTree func(detail Group)) (_ vl.Widget) // return gui widget
}

type record struct {
	Index GroupIndex
	Data  string
}

// SaveGroup return stored information at json format
func SaveGroup(g Group) (bs []byte, err error) {
	switch m := g.(type) {
	case *Meta: // only for groups with slice of Groups
		var store struct {
			Name  string
			Datas []string
		}
		store.Name = m.Name
		for _, g := range m.Groups {
			var b []byte
			b, err = SaveGroup(g)
			if err != nil {
				return
			}
			store.Datas = append(store.Datas, string(b))
		}
		bs, err = json.Marshal(&store)
		if err != nil {
			return
		}
	default: // all other groups
		bs, err = json.Marshal(g)
		if err != nil {
			return
		}
	}
	r := record{Index: g.GetId(), Data: string(bs)}
	return json.Marshal(&r)
}

// ParseGroup return group from json
func ParseGroup(bs []byte) (g Group, err error) {
	var r record
	err = json.Unmarshal(bs, &r)
	if err != nil {
		return
	}
	Id := r.Index
	var ok bool
	g, ok = Id.newInstance()
	if !ok {
		err = fmt.Errorf("cannot create new instance for: %d", uint16(Id))
		return
	}
	switch m := g.(type) {
	case *Meta: // only for groups with slice of Groups
		var store struct {
			Name  string
			Datas []string
		}
		err = json.Unmarshal([]byte(r.Data), &store)
		if err != nil {
			return
		}
		m.Name = store.Name
		for _, d := range store.Datas {
			var inner Group
			inner, err = ParseGroup([]byte(d))
			if err != nil {
				return
			}
			m.Groups = append(m.Groups, inner)
		}

	default: // all other groups
		err = json.Unmarshal([]byte(r.Data), m)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

func treeNode(
	g Group, mesh Mesh,
	updateTree func(g Group),
	updateDetail func(w vl.Widget),
) (
	t vl.Tree,
) {
	if g == nil {
	}
	// prepare edit widget
	var list vl.List
	list.Add(vl.TextStatic(g.GetId().String() + ":"))
	var btn vl.Button
	btn.Compress = true
	btn.SetText(g.String())
	btn.OnClick = func() {
		edit := g.GetWidget(mesh, updateTree)
		if edit == nil {
			return
		}
		updateDetail(edit)
	}
	list.Add(&btn)
	t.Root = &list

	switch m := g.(type) {
	case *Meta:
		for i := range m.Groups {
			if m.Groups[i] == nil {
				logger.Printf("NOT ACCEPTABLE NIL")
			}
			w := treeNode(m.Groups[i], mesh, updateTree, updateDetail)
			t.Nodes = append(t.Nodes, w)
		}
		return
	}
	return
}

type modelTree struct {
	Tree          vl.Widget
	Detail        vl.Widget
	currentDetail vl.Widget
}

func NewGroupTree(mesh Mesh, closedApp *bool, actions *chan ds.Action) (gt vl.Widget, initialization func(), err error) {
	var list vl.List
	defer func() {
		gt = &list
	}()

	list.Add(vl.TextStatic("model tree"))
	list.Add(vl.TextStatic("detail of model tree nodes"))

	var updateDetail func(w vl.Widget)
	updateDetail = func(w vl.Widget) {
		if w == nil {
			updateDetail(vl.TextStatic("Choose node of model tree for modification"))
			return
		}
		if l1, ok := list.Get(1).(*vl.Scroll); ok {
			l1.Root = w
		} else {
			var scroll vl.Scroll
			scroll.Root = w
			list.Update(1, &scroll)
		}
	}

	var updateTree func(detail Group)
	updateTree = func(detail Group) {
		tree := treeNode(
			mesh.GetRootGroup(), mesh,
			updateTree,
			updateDetail,
		)
		if l0, ok := list.Get(0).(*vl.Scroll); ok {
			l0.Root = &tree
		} else {
			var scroll vl.Scroll
			scroll.Root = &tree
			list.Update(0, &scroll)
		}
		updateDetail(detail.GetWidget(mesh, updateTree))
	}

	// first update
	initialization = func() {
		updateTree(mesh.GetRootGroup())
	}

	return
}

///////////////////////////////////////////////////////////////////////////////

type Named struct{ Name string }

func (m Named) GetId() GroupIndex { return NamedIndex }
func (m Named) String() (name string) {
	if m.Name == "" {
		return "noname"
	}
	return strings.ToUpper(m.Name)
}
func (m *Named) Update(updating func(nodes, elements *[]uint)) { return }

func (m *Named) GetWidget(mesh Mesh, updateTree func(g Group)) (w vl.Widget) {
	var list vl.List
	list.IgnoreVerticalFix = true
	defer func() {
		w = &list
	}()
	{
		list.Add(vl.TextStatic("Rename:"))
		var name vl.Inputbox
		list.Add(&name)
		name.SetText(m.Name)
		var btn vl.Button
		btn.SetText("Rename")
		btn.OnClick = func() {
			m.Name = name.GetText()
			updateTree(m)
		}
		list.Add(&btn)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(Meta)

type Meta struct {
	Named
	Groups []Group
}

func (m Meta) GetId() GroupIndex { return MetaIndex }
func (m *Meta) Update(updating func(nodes, elements *[]uint)) {
	for _, gr := range m.Groups {
		if gr == nil {
			continue
		}
		gr.Update(updating)
	}
}
func (m *Meta) GetWidget(mm Mesh, updateTree func(g Group)) (w vl.Widget) {
	var list vl.List
	list.IgnoreVerticalFix = true
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(mm, func(_ Group) {
			updateTree(m)
		})
		list.Add(n)
	}
	list.Add(new(vl.Separator))
	{
		list.Add(vl.TextStatic("Add new group:"))
		var ids []GroupIndex
		var names []string
		for i := uint16(0); i < math.MaxUint16; i++ {
			gi := GroupIndex(i)
			_, ok := gi.newInstance()
			if !ok {
				continue
			}
			ids = append(ids, gi)
			names = append(names, gi.String())
		}
		var combo vl.Combobox
		combo.Add(names...)
		list.Add(&combo)
		if 0 < len(names) {
			combo.SetPos(0)
		}
		var btn vl.Button
		btn.SetText("Add group")
		btn.Compress = true
		btn.OnClick = func() {
			pos := combo.GetPos()
			gi := GroupIndex(ids[pos])
			g, ok := gi.newInstance()
			if !ok {
				return
			}
			m.Groups = append(m.Groups, g)
			updateTree(m)
		}
		list.Add(&btn)
	}
	list.Add(new(vl.Separator))
	{
		list.Add(vl.TextStatic("Remove group:"))
		var names []string
		var combo vl.Combobox
		list.Add(&combo)
		names = []string{"NONE"}
		for _, gr := range m.Groups {
			names = append(names, gr.GetId().String()+": "+gr.String())
		}
		combo.Add(names...)
		combo.SetPos(0)
		var btn vl.Button
		btn.SetText("Remove group")
		btn.Compress = true
		btn.OnClick = func() {
			pos := combo.GetPos()
			if pos == 0 {
				return // NONE selected group
			}
			pos -= 1
			m.Groups = append(m.Groups[:pos], m.Groups[pos+1:]...)
			updateTree(m)
		}
		list.Add(&btn)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NamedList)

type NamedList struct {
	Named
	Nodes, Elements []uint
}

func (m NamedList) String() (name string) {
	return fmt.Sprintf("%s: for %d nodes and %d elements",
		m.Named.String(), len(m.Nodes), len(m.Elements))
}
func (m NamedList) GetId() GroupIndex { return NamedListIndex }
func (m *NamedList) Update(updating func(nodes, elements *[]uint)) {
	updating(&m.Nodes, &m.Elements)
}
func (m *NamedList) GetWidget(mm Mesh, updateTree func(g Group)) (w vl.Widget) {
	var list vl.List
	list.IgnoreVerticalFix = true
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(mm, func(_ Group) {
			updateTree(m)
		})
		list.Add(n)
		list.Add(new(vl.Separator))
	}
	{
		var btn vl.Button
		btn.SetText("Select")
		btn.OnClick = func() {
			mm.Select(m.Nodes, m.Elements)
		}
		list.Add(&btn)
		list.Add(new(vl.Separator))
	}
	{
		change := Change(mm, true, true, &m.Nodes, &m.Elements, func() {
			updateTree(m)
		})
		list.Add(change)
		list.Add(new(vl.Separator))
		// TODO update screen
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

var dir = [6]string{"Dx", "Dy", "Dz", "Rx", "Ry", "Rz"}

var _ Group = new(NodeSupports)

type NodeSupports struct {
	Named
	Direction [6]bool
	Nodes     []uint
}

func (m NodeSupports) GetId() GroupIndex {
	return NodeSupportsIndex
}

func (m NodeSupports) String() (name string) {
	name += fmt.Sprintf("%s: ", m.Named.String())
	for i := range m.Direction {
		if !m.Direction[i] {
			continue
		}
		name += dir[i] + " "
	}
	name += fmt.Sprintf("for %d nodes", len(m.Nodes))
	return
}

func (m *NodeSupports) Update(updating func(nodes, elements *[]uint)) {
	updating(&m.Nodes, nil)
}

func (m *NodeSupports) GetWidget(mm Mesh, updateTree func(g Group)) (w vl.Widget) {
	var list vl.List
	list.IgnoreVerticalFix = true
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(mm, func(_ Group) {
			updateTree(m)
		})
		list.Add(n)
		list.Add(new(vl.Separator))
	}
	{
		var btn vl.Button
		btn.SetText("Select")
		btn.OnClick = func() {
			mm.Select(m.Nodes, nil)
		}
		list.Add(&btn)
		list.Add(new(vl.Separator))
	}
	{
		list.Add(vl.TextStatic("Fixed node support direction:"))
		for i := range m.Direction {
			i := i
			var ch vl.CheckBox
			ch.SetText(dir[i])
			ch.Checked = m.Direction[i]
			ch.OnChange = func() {
				m.Direction[i] = ch.Checked
				updateTree(m)
			}
			list.Add(&ch)
		}
		list.Add(new(vl.Separator))
	}
	{
		change := Change(mm, true, false, &m.Nodes, nil, func() {
			updateTree(m)
		})
		list.Add(change)
		list.Add(new(vl.Separator))
		// TODO update screen
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

func Change(m Mesh,
	nb, eb bool,
	nodes, elements *[]uint,
	update func(),
) (
	w vl.Widget,
) {
	var list vl.List
	list.IgnoreVerticalFix = true
	defer func() {
		w = &list
	}()

	if !nb && !eb {
		panic("no need widget")
	}

	var (
		l1     vl.ListH
		coords vl.Text
		b      vl.Button
		l2     vl.ListH
		els    vl.Text
	)

	if nb && eb {
		list.Add(vl.TextStatic("List of nodes and elements:"))
	} else if nb {
		list.Add(vl.TextStatic("List of nodes:"))
	} else {
		list.Add(vl.TextStatic("List of elements:"))
	}

	l1.Add(vl.TextStatic("Nodes:"))

	coords.SetLinesLimit(3)
	if nodes != nil {
		coords.SetText(fmt.Sprintf("%d", *nodes))
	}
	l1.Add(&coords)

	b.SetText("Change")
	b.OnClick = func() {
		newCoordinates := m.GetSelectNodes(Many)
		newElements := m.GetSelectElements(Many, nil)
		if len(newCoordinates) == 0 && len(newElements) == 0 {
			return
		}
		if nb {
			*nodes = newCoordinates
			coords.SetText(fmt.Sprintf("%v", newCoordinates))
		}
		if eb {
			*elements = newElements
			els.SetText(fmt.Sprintf("%v", newElements))
		}
		update()
	}
	l1.Add(&b)
	if nb {
		list.Add(&l1)
	}

	l2.Add(vl.TextStatic("Elements:"))

	els.SetLinesLimit(3)
	if elements != nil {
		els.SetText(fmt.Sprintf("%d", *elements))
	}
	l2.Add(&els)

	l2.Add(vl.TextStatic(""))
	if eb {
		list.Add(&l2)
	}

	return
}

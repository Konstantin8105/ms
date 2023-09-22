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
	NamedListIndex    GroupIndex = 100
	NodeSupportsIndex            = 1000
	MetaIndex                    = 10000
)

func (gi GroupIndex) String() string {
	switch gi {
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
	String() string                                         // return short name
	Update(updating func(nodes, elements *[]uint))          // update nodes, elements indexes
	GetWidget(mm Mesh) (_ vl.Widget, initialization func()) // return gui widget
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

func treeNode(g Group, mesh Mesh, w *vl.Widget) (t vl.Tree, initialization func()) {
	if g == nil {
		return
	}
	var inits []func()
	defer func() {
		initialization = func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
	}()

	// prepare edit widget
	var list vl.List
	list.Add(vl.TextStatic(g.GetId().String() + ":"))
	var btn vl.Button
	btn.Compress = true
	btn.SetText(g.String())
	btn.OnClick = func() {
		if w == nil {
			return
		}
		edit, init := g.GetWidget(mesh)
		inits = append(inits, init)
		if edit == nil {
			return
		}
		*w = edit
	}
	list.Add(&btn)
	t.Root = &list

	switch m := g.(type) {
	case *Meta:
		for i := range m.Groups {
			w, init := treeNode(m.Groups[i], mesh, w)
			inits = append(inits, init)
			t.Nodes = append(t.Nodes, w)
		}
		return
	}
	return
}

func NewGroupTree(mesh Mesh, closedApp *bool, actions *chan ds.Action) (gt vl.Widget, initialization func(), err error) {
	var list vl.List

	// default tree node widget
	var w vl.Widget = vl.TextStatic("Click on tree for modify")

	// group tree
	var tree vl.Tree
	tree, initialization = treeNode(mesh.GetRootGroup(), mesh, &w)
	list.Add(&tree)
	list.Add(w)

	gt = &list
	return
}

///////////////////////////////////////////////////////////////////////////////

type Named struct{ Name string }

func (m Named) String() (name string) {
	if m.Name == "" {
		return "noname"
	}
	return strings.ToUpper(m.Name)
}

func (m *Named) GetWidget(mesh Mesh) (w vl.Widget, initialization func()) {
	var list vl.List

	list.Add(vl.TextStatic("Rename:"))
	var name vl.Inputbox
	list.Add(&name)
	initialization = func() {
		name.SetText(m.Name)
	}
	var btn vl.Button
	btn.SetText("Rename")
	btn.OnClick = func() {
		m.Name = name.GetText()
	}
	list.Add(&btn)

	w = &list
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
func (m *Meta) GetWidget(mm Mesh) (w vl.Widget, initialization func()) {
	var inits []func()
	defer func() {
		initialization = func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
	}()
	var list vl.List
	defer func() {
		w = &list
	}()
	{
		n, ni := m.Named.GetWidget(mm)
		list.Add(n)
		inits = append(inits, ni)
		list.Add(new(vl.Separator))
	}
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
		inits = append(inits, func() {
			combo.SetPos(0)
		})
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
			// TODO update view
			// TODO update initialization funcs
		}
		list.Add(&btn)
		list.Add(new(vl.Separator))
	}
	{
		list.Add(vl.TextStatic("Remove group:"))
		var names []string
		var combo vl.Combobox
		list.Add(&combo)
		inits = append(inits, func() {
			combo.Clear()
			names = []string{"NONE"}
			for _, gr := range m.Groups {
				names = append(names, gr.GetId().String()+": "+gr.String())
			}
			combo.Add(names...)
			combo.SetPos(0)
		})
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
			// TODO update view
			// TODO update initialization funcs
		}
		list.Add(&btn)
		list.Add(new(vl.Separator))
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
func (m *NamedList) GetWidget(mm Mesh) (w vl.Widget, initialization func()) {
	var inits []func()
	defer func() {
		initialization = func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
	}()
	var list vl.List
	defer func() {
		w = &list
	}()
	{
		n, ni := m.Named.GetWidget(mm)
		list.Add(n)
		inits = append(inits, ni)
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
		change, init := Change(mm, true, true, &m.Nodes, &m.Elements)
		list.Add(change)
		inits = append(inits, init)
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

func (m *NodeSupports) GetWidget(mm Mesh) (w vl.Widget, initialization func()) {
	var inits []func()
	defer func() {
		initialization = func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
	}()
	var list vl.List
	defer func() {
		w = &list
	}()
	{
		n, ni := m.Named.GetWidget(mm)
		list.Add(n)
		inits = append(inits, ni)
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
				// TODO update screen
			}
			list.Add(&ch)
		}
		list.Add(new(vl.Separator))
	}
	{
		change, init := Change(mm, true, false, &m.Nodes, nil)
		list.Add(change)
		inits = append(inits, init)
		list.Add(new(vl.Separator))
		// TODO update screen
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

func Change(m Mesh,
	nb, eb bool,
	nodes, elements *[]uint,
) (
	w vl.Widget,
	initialization func(),
) {
	var inits []func()
	defer func() {
		initialization = func() {
			for i := range inits {
				if f := inits[i]; f != nil {
					f()
				}
			}
		}
	}()
	var list vl.List
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

	list.Add(vl.TextStatic("List of nodes and elements:"))

	l1.Add(vl.TextStatic("Nodes:"))

	coords.SetLinesLimit(3)
	inits = append(inits, func() {
		coords.SetText(fmt.Sprintf("%d", *nodes))
	})
	l1.Add(&coords)

	b.SetText("Change")
	b.OnClick = func() {
		coordinates := m.GetSelectNodes(Many)
		elements := m.GetSelectElements(Many, nil)
		if len(coordinates) == 0 && len(elements) == 0 {
			return
		}
		coords.SetText(fmt.Sprintf("%v", coordinates))
		els.SetText(fmt.Sprintf("%v", elements))
	}
	l1.Add(&b)
	if nb {
		list.Add(&l1)
	}

	l2.Add(vl.TextStatic("Elements:"))

	els.SetLinesLimit(3)
	inits = append(inits, func() {
		els.SetText(fmt.Sprintf("%d", *nodes))
	})
	l2.Add(&els)

	l2.Add(vl.TextStatic(""))
	if eb {
		list.Add(&l2)
	}

	return
}

package groups

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/vl"
)

///////////////////////////////////////////////////////////////////////////////

type Mesh interface {
	GetRootGroup() Group
	Update(nodes, elements *uint)

	Select(nodes, elements []uint)
	GetSelected() (nodes, elements []uint)
}

func FixMesh(mesh Mesh) {
	if mesh == nil {
		return
	}
	var maxId int = 1
	{
		// search maximal id
		compare := func(id int) {
			if maxId < id {
				maxId = id
			}
		}
		var walk func(gr Group)
		walk = func(gr Group) {
			if gr == nil {
				return
			}
			compare(gr.GetUniqueId())
			switch n := gr.(type) {
			case *Meta:
				for i := range n.Groups {
					walk(n.Groups[i])
				}
			}
		}
		walk(mesh.GetRootGroup())
	}
	// TODO problem - check with same ids

	// set id only if equal zero and root
	{
		set := func(gr Group) {
			if gr == nil {
				return
			}
			gr.SetRoot(mesh)
			if gr.GetUniqueId() == 0 {
				maxId++
				gr.SetUniqueId(maxId)
			}
		}
		var walk func(gr Group)
		walk = func(gr Group) {
			set(gr)
			switch n := gr.(type) {
			case *Meta:
				for i := range n.Groups {
					walk(n.Groups[i])
				}
			}
		}
		walk(mesh.GetRootGroup())
	}
}

///////////////////////////////////////////////////////////////////////////////

type GroupTest struct{ base Group }

func (g GroupTest) GetRootGroup() Group {
	return g.base
}
func (g GroupTest) Update(nodes, elements *uint)          {}
func (g GroupTest) Select(nodes, elements []uint)         {}
func (g GroupTest) GetSelected() (nodes, elements []uint) { return nil, nil }

///////////////////////////////////////////////////////////////////////////////

// max uint16: 65535
type GroupIndex uint16

const (
	NamedIndex        GroupIndex = 80
	NamedListIndex               = 100
	NodeSupportsIndex            = 1000
	MetaIndex                    = 10000
	CopyIndex                    = 10100
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
	case CopyIndex:
		return "Group copy"
	}
	return fmt.Sprintf("Undefined name of group %d", uint16(gi))
}

func (gi GroupIndex) newInstance(root Mesh) (gr Group, ok bool) {
	switch gi {
	case NamedIndex:
		gr, ok = new(Named), true
	case NamedListIndex:
		gr, ok = new(NamedList), true
	case NodeSupportsIndex:
		gr, ok = new(NodeSupports), true
	case MetaIndex:
		gr, ok = new(Meta), true
	case CopyIndex:
		gr, ok = new(Copy), true
	}
	if ok {
		gr.SetRoot(root)
		FixMesh(root)
	}
	return
}

type Group interface {
	// UniqueId of group, only for save/parse of group
	SetUniqueId(id int)
	GetUniqueId() int

	// update root group
	SetRoot(root Mesh)

	// operations with group
	GetGroupIndex() GroupIndex
	String() string
	Update(updating func(nodes, elements *[]uint))         // update nodes, elements indexes
	GetWidget(updateTree func(detail Group)) (_ vl.Widget) // return gui widget
}

///////////////////////////////////////////////////////////////////////////////

type record struct {
	Index GroupIndex
	Data  string
}

// SaveGroup return stored information at json format
func SaveGroup(gr Group) (bs []byte, err error) {
	var nodes []node

	add := func(gr Group) {
		n := node{
			Index: gr.GetGroupIndex(),
			Gr:    gr,
		}
		nodes = append(nodes, n)
	}
	var walk func(gr Group)
	walk = func(gr Group) {
		add(gr)
		switch n := gr.(type) {
		case *Meta:
			for i := range n.Groups {
				walk(n.Groups[i])
			}
		}
	}
	walk(gr)

	var records []record
	for i := range nodes {
		var local []byte
		switch m := nodes[i].Gr.(type) {
		case *Meta: // only for groups with slice of Groups
			var store struct {
				Name string
				ID   int
				Ids  []int
			}
			store.Name = m.Name
			store.ID = m.ID
			for _, gr := range m.Groups {
				store.Ids = append(store.Ids, gr.GetUniqueId())
			}
			local, err = json.Marshal(&store)
			if err != nil {
				return
			}
		default: // all other groups
			local, err = json.Marshal(m)
			if err != nil {
				return
			}
		}
		r := record{
			Index: nodes[i].Gr.GetGroupIndex(),
			Data:  string(local),
		}
		records = append(records, r)
	}
	return json.MarshalIndent(&records, "", "\t")
}

type node struct {
	Index GroupIndex
	Gr    Group
}

// ParseGroup return group from json
func ParseGroup(bs []byte) (gr Group, err error) {
	var records []record
	err = json.Unmarshal(bs, &records)
	if err != nil {
		return
	}
	var groups []Group
	for _, record := range records {
		gr, ok := record.Index.newInstance(nil)
		if !ok {
			err = fmt.Errorf("cannot create new instance for: %d", record.Index)
			return nil, err
		}
		if record.Index == MetaIndex {
			continue
		}
		err = json.Unmarshal([]byte(record.Data), gr)
		if err != nil {
			return nil, err
		}
		groups = append(groups, gr)
	}
	for i := len(records) - 1; 0 <= i; i-- {
		record := records[i]
		var ok bool
		_, ok = record.Index.newInstance(nil)
		if !ok {
			err = fmt.Errorf("cannot create new instance for: %d", record.Index)
			return
		}
		if record.Index != MetaIndex {
			continue
		}
		var store struct {
			Name string
			ID   int
			Ids  []int
		}
		err = json.Unmarshal([]byte(record.Data), &store)
		if err != nil {
			return
		}
		var m Meta
		m.Name = store.Name
		m.ID = store.ID

		for _, id := range store.Ids {
			found := false
			for ig := range groups {
				if id != groups[ig].GetUniqueId() {
					continue
				}
				found = true
				m.Groups = append(m.Groups, groups[ig])
				groups = append(groups[:ig], groups[ig+1:]...)
				break
			}
			if !found {
				fmt.Println("Not found:", id)
			}
		}
		groups = append(groups, &m)
	}
	if len(groups) == 1 {
		gr = groups[0]
	} else {
		err = fmt.Errorf("Not valid amounts groups: %d", len(groups))
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

func treeNode(
	gr Group, mesh Mesh,
	updateTree func(gr Group),
	updateDetail func(w vl.Widget),
) (
	t vl.Tree,
) {
	if gr == nil {
	}
	// prepare edit widget
	var list vl.List
	list.Add(vl.TextStatic(gr.GetGroupIndex().String() + ":"))
	var btn vl.Button
	btn.Compress()
	btn.SetText(gr.String())
	btn.OnClick = func() {
		edit := gr.GetWidget(updateTree)
		if edit == nil {
			return
		}
		updateDetail(edit)
	}
	list.Add(&btn)
	t.Root = &list

	switch m := gr.(type) {
	case *Meta:
		for i := range m.Groups {
			if m.Groups[i] == nil {
				// logger.Printf("NOT ACCEPTABLE NIL")
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
			l1.SetRoot(w)
		} else {
			var scroll vl.Scroll
			scroll.SetRoot(w)
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
			l0.SetRoot(&tree)
		} else {
			var scroll vl.Scroll
			scroll.SetRoot(&tree)
			list.Update(0, &scroll)
		}
		updateDetail(detail.GetWidget(updateTree))
	}

	// first update
	initialization = func() {
		updateTree(mesh.GetRootGroup())
	}

	return
}

///////////////////////////////////////////////////////////////////////////////

type rootBase struct {
	root Mesh
}

func (r *rootBase) SetRoot(root Mesh) {
	r.root = root
}

///////////////////////////////////////////////////////////////////////////////

type Idable struct{ ID int }

func (id *Idable) SetUniqueId(v int) { id.ID = v }
func (id Idable) GetUniqueId() int   { return id.ID }

///////////////////////////////////////////////////////////////////////////////

type Named struct {
	Idable
	rootBase
	Name string
}

func (m Named) GetGroupIndex() GroupIndex { return NamedIndex }
func (m Named) String() (name string) {
	if m.Name == "" {
		return "noname"
	}
	return strings.ToUpper(m.Name)
}
func (m *Named) Update(updating func(nodes, elements *[]uint)) { return }

func (m *Named) GetWidget(updateTree func(gr Group)) (w vl.Widget) {
	var list vl.List
	list.Compress()
	defer func() {
		w = &list
	}()
	{
		list.Add(vl.TextStatic("Rename:"))
		var name vl.InputBox
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
	Idable
	Named
	Groups []Group
}

func (m Meta) GetGroupIndex() GroupIndex { return MetaIndex }
func (m *Meta) Update(updating func(nodes, elements *[]uint)) {
	for _, gr := range m.Groups {
		if gr == nil {
			continue
		}
		gr.Update(updating)
	}
}
func (m *Meta) GetWidget(updateTree func(gr Group)) (w vl.Widget) {
	var list vl.List
	list.Compress()
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(func(_ Group) {
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
			_, ok := gi.newInstance(m.root)
			if !ok {
				continue
			}
			ids = append(ids, gi)
			names = append(names, gi.String())
		}
		var combo vl.ComboBox
		combo.Add(names...)
		list.Add(&combo)
		if 0 < len(names) {
			combo.SetPos(0)
		}
		var btn vl.Button
		btn.SetText("Add group")
		btn.Compress()
		btn.OnClick = func() {
			pos := combo.GetPos()
			gi := GroupIndex(ids[pos])
			g, ok := gi.newInstance(m.root)
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
		var combo vl.ComboBox
		list.Add(&combo)
		names = []string{"NONE"}
		for _, gr := range m.Groups {
			names = append(names, gr.GetGroupIndex().String()+": "+gr.String())
		}
		combo.Add(names...)
		combo.SetPos(0)
		var btn vl.Button
		btn.SetText("Remove group")
		btn.Compress()
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
	Idable
	Named
	Nodes, Elements []uint
}

func (m NamedList) String() (name string) {
	return fmt.Sprintf("%s: for %d nodes and %d elements",
		m.Named.String(), len(m.Nodes), len(m.Elements))
}
func (m NamedList) GetGroupIndex() GroupIndex { return NamedListIndex }
func (m *NamedList) Update(updating func(nodes, elements *[]uint)) {
	updating(&m.Nodes, &m.Elements)
}
func (m *NamedList) GetWidget(updateTree func(gr Group)) (w vl.Widget) {
	var list vl.List
	list.Compress()
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(func(_ Group) {
			updateTree(m)
		})
		list.Add(n)
		list.Add(new(vl.Separator))
	}
	{
		var btn vl.Button
		btn.SetText("Select")
		btn.OnClick = func() {
			m.root.Select(m.Nodes, m.Elements)
		}
		list.Add(&btn)
		list.Add(new(vl.Separator))
	}
	{
		change := Change(m.root, true, true, &m.Nodes, &m.Elements, func() {
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
	Idable
	Named
	Direction [6]bool
	Nodes     []uint
}

func (m NodeSupports) GetGroupIndex() GroupIndex {
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

func (m *NodeSupports) GetWidget(updateTree func(gr Group)) (w vl.Widget) {
	var list vl.List
	list.Compress()
	defer func() {
		w = &list
	}()
	{
		n := m.Named.GetWidget(func(_ Group) {
			updateTree(m)
		})
		list.Add(n)
		list.Add(new(vl.Separator))
	}
	{
		var btn vl.Button
		btn.SetText("Select")
		btn.OnClick = func() {
			m.root.Select(m.Nodes, nil)
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
		change := Change(m.root, true, false, &m.Nodes, nil, func() {
			updateTree(m)
		})
		list.Add(change)
		list.Add(new(vl.Separator))
		// TODO update screen
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

type Copy struct {
	Idable
	rootBase
	Link int
}

func getGroupById(id int, root Group) (group Group) {
	var walk func(gr Group)
	walk = func(gr Group) {
		if gr == nil {
			return
		}
		if gr.GetUniqueId() == id {
			group = gr
			return
		}
		switch n := gr.(type) {
		case *Meta:
			for i := range n.Groups {
				walk(n.Groups[i])
			}
		}
	}
	walk(root)
	return
}

func getNodes(
	root Group,
	filter func(gr Group) bool,
) (names []string, ids []int) {
	set := func(gr Group) {
		if filter != nil {
			if filter(gr) {
				return
			}
		}
		names = append(names, gr.String())
		ids = append(ids, gr.GetUniqueId())
	}
	var walk func(gr Group)
	walk = func(gr Group) {
		set(gr)
		switch n := gr.(type) {
		case *Meta:
			for i := range n.Groups {
				walk(n.Groups[i])
			}
		}
	}
	walk(root)
	return
}

func (c Copy) GetGroupIndex() GroupIndex { return CopyIndex }
func (c Copy) String() string {
	if c.Link == 0 {
		return "No copy"
	}
	if c.root == nil {
		return fmt.Sprintf("Link have not root")
	}
	gr := getGroupById(c.Link, c.root.GetRootGroup())
	if gr == nil {
		return fmt.Sprintf("Link to undefined group")
	}
	return fmt.Sprintf("Id:%d\n%s", c.Link, gr.String())
}
func (c Copy) Update(updating func(nodes, elements *[]uint)) { return }
func (c *Copy) GetWidget(updateTree func(detail Group)) (w vl.Widget) {
	var list vl.List
	list.Compress()
	defer func() {
		w = &list
	}()
	{
		list.Add(vl.TextStatic("Change link id:"))
		names, ids := getNodes(
			c.root.GetRootGroup(),
			func(gr Group) bool { // filter
				switch gr.(type) {
				case *Copy:
					return true
				}
				return false
			},
		)

		names = append([]string{"NONE"}, names...)
		ids = append([]int{-1}, ids...)

		var combo vl.ComboBox
		combo.Add(names...)
		list.Add(&combo)
		if 0 < len(names) {
			combo.SetPos(0)
		}
		var btn vl.Button
		btn.SetText("Link")
		btn.Compress()
		btn.OnClick = func() {
			pos := combo.GetPos()
			if pos == 0 {
				return
			}
			if len(ids) <= int(pos) {
				return
			}
			c.Link = ids[pos]
			updateTree(c)
		}
		list.Add(&btn)
	}
	return
}

///////////////////////////////////////////////////////////////////////////////

// TODO move nodes up and down
// TODO remove node

///////////////////////////////////////////////////////////////////////////////

func Change(m Mesh,
	nb, eb bool,
	nodes, elements *[]uint,
	update func(),
) (
	w vl.Widget,
) {
	var list vl.List
	list.Compress()
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
		newCoordinates, newElements := m.GetSelected()
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

///////////////////////////////////////////////////////////////////////////////

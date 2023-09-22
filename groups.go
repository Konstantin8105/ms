package ms

import (
	"encoding/json"
	"fmt"
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
	String() string                    // return short name
	Update() (nodes, elements *[]uint) // update nodes, elements indexes
	GetWidget() vl.Widget              // return gui widget
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

func treeNode(g Group) (t vl.Tree) {
	t.Root = vl.TextStatic(g.GetId().String() + ":\n" + g.String())
	switch m := g.(type) {
	case *Meta:
		for i := range m.Groups {
			t.Nodes = append(t.Nodes, treeNode(m.Groups[i]))
		}
		return
	}
	return
}

func NewGroupTree(mainGroup *Meta, closedApp *bool, actions *chan ds.Action) (gt vl.Widget, initialization func(), err error) {
	var list vl.List

	// group tree
	tree := treeNode(mainGroup)
	list.Add(&tree)

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
func (m Named) GetWidget() vl.Widget {
	return vl.TextStatic("Undefined widget:" + m.Name)
}

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(Meta)

type Meta struct {
	Named
	Groups []Group
}

func (m Meta) GetId() GroupIndex                  { return MetaIndex }
func (m *Meta) Update() (nodes, elements *[]uint) { return nil, nil }
func (m *Meta) Add(g Group) {
	if g == nil {
		return
	}
	m.Groups = append(m.Groups, g)
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
func (m NamedList) GetId() GroupIndex                  { return NamedListIndex }
func (m *NamedList) Update() (nodes, elements *[]uint) { return &m.Nodes, &m.Elements }

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NodeSupports)

type NodeSupports struct {
	Named
	Nodes     []uint
	Direction [6]bool
}

func (m NodeSupports) GetId() GroupIndex {
	return NodeSupportsIndex
}

func (m NodeSupports) String() (name string) {
	name += fmt.Sprintf("%s: ", m.Named.String())
	dir := [6]string{"Dx", "Dy", "Dz", "Rx", "Ry", "Rz"}
	for i := range m.Direction {
		if !m.Direction[i] {
			continue
		}
		name += dir[i] + " "
	}
	name += fmt.Sprintf("for %d nodes", len(m.Nodes))
	return
}

func (m *NodeSupports) Update() (nodes, elements *[]uint) {
	return &m.Nodes, nil
}

///////////////////////////////////////////////////////////////////////////////

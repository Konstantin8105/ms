package ms

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Konstantin8105/vl"
)

//	type Model2 struct {
//		// ...
//		Data []struct {
//			Id          uint64
//			Information []byte
//			group       Group
//		}
//		// ...
//	}
type Group interface {
	GetShortName() string              // return short name
	Update() (nodes, elements *[]uint) // update nodes, elements indexes
	GetWidget() vl.Widget              // return gui widget
	Save() (Id uint64, Information []byte, err error)
}

// list of all allowable groups
var groups = []struct {
	id          uint64
	name        string
	newInstance func() Group
}{
	{id: 1000, name: "Named list", newInstance: func() Group { return new(NamedList) }},
	{id: 10000, name: "Supports", newInstance: func() Group { return new(NodeSupports) }},
	{id: 100000, name: "Meta", newInstance: func() Group { return new(Meta) }},
	// {id: math.MaxUint64, name: "Empty group", newInstance: func() Group { return new(Empty) }},
}

// ParseGroup return group from json
func parseGroup(Id uint64, Information []byte) (group Group, err error) {
	for i := range groups {
		if Id != groups[i].id {
			continue
		}
		bs := []byte(Information)
		group = groups[i].newInstance()
		err = json.Unmarshal(bs, group)
		return
	}
	err = fmt.Errorf("cannot parse with Id: %d", Id)
	return
}

// SaveGroup return stored information at json format
func saveGroup(group Group) (Id uint64, Information []byte, err error) {
	for i := range groups {
		if reflect.TypeOf(groups[i].newInstance()) != reflect.TypeOf(group) {
			continue
		}
		Id = groups[i].id
	}
	b, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		return
	}
	Information = b
	return
}

///////////////////////////////////////////////////////////////////////////////

type Named struct {
	Name string
}

func (m Named) GetShortName() (name string) {
	if m.Name == "" {
		return "undefined"
	}
	return m.Name
}

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(Meta)

type Meta struct {
	Named
	Groups []Group
}

func (m *Meta) Update() (nodes, elements *[]uint) { return nil, nil }
func (m *Meta) GetWidget() vl.Widget              { return new(vl.Separator) }
func (m *Meta) Save() (Id uint64, Information []byte, err error) { return saveGroup(m) }

///////////////////////////////////////////////////////////////////////////////
//
// var _ Group = new(Empty)
//
// type Empty struct{}
//
// func (Empty) GetShortName() (name string)       { return "empty" }
// func (Empty) Update() (nodes, elements *[]uint) { return nil, nil }
// func (Empty) GetWidget() vl.Widget              { return new(vl.Separator) }
// func (Empty) Save() (Id uint64, Information []byte, err error) { return saveGroup(m) }
//
///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NamedList)

type NamedList struct {
	Named
	Nodes, Elements []uint
}

func (m *NamedList) Update() (nodes, elements *[]uint) { return &m.Nodes, &m.Elements }
func (m *NamedList) GetWidget() vl.Widget {
	var list vl.List
	list.Add(vl.TextStatic("Name"))
	list.Add(vl.TextStatic("Nodes: 1,4,6,8"))
	list.Add(vl.TextStatic("Elements: 12,42,63,85"))
	return &list
}
func (m *NamedList) Save() (Id uint64, Information []byte, err error) { return saveGroup(m) }

///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NodeSupports)

type NodeSupports struct {
	Nodes     []uint
	Direction [6]bool
}

func (m NodeSupports) GetShortName() (name string) {
	dir := [6]string{"Dx", "Dy", "Dz", "Rx", "Ry", "Rz"}
	for i := range m.Direction {
		if !m.Direction[i] {
			continue
		}
		name += dir[i] + " "
	}
	name += fmt.Sprintf("for %d nodes", m.Nodes)
	return
}

func (m *NodeSupports) Update() (nodes, elements *[]uint) {
	return &m.Nodes, nil
}

func (m *NodeSupports) GetWidget() vl.Widget {
	var list vl.List
	list.Add(vl.TextStatic("Node supports"))
	return &list
}

func (m *NodeSupports) Save() (Id uint64, Information []byte, err error) { return saveGroup(m) }

///////////////////////////////////////////////////////////////////////////////

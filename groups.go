package ms

import (
	"encoding/json"
	"fmt"

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

type Record struct {
	Index GroupIndex
	Data  []byte
}

// SaveGroup return stored information at json format
func (gi GroupIndex) Save(g Group) (_ []byte, err error) {
	b, err := json.Marshal(g)
	if err != nil {
		return
	}
	r := Record{Index: gi, Data: b}
	return json.Marshal(&r)
}

//	type Model2 struct {
//		// ...
//		Data []struct {
//			Id          GroupIndex
//			Information []byte
//			group       Group
//		}
//		// ...
//	}
type Group interface {
	GetId() GroupIndex
	GetShortName() string              // return short name
	Update() (nodes, elements *[]uint) // update nodes, elements indexes
	GetWidget() vl.Widget              // return gui widget
	Save() (Information []byte, err error)
}

// ParseGroup return group from json
func ParseGroup(Id GroupIndex, Information []byte) (group Group, err error) {
	var ok bool
	group, ok = Id.newInstance()
	if !ok {
		err = fmt.Errorf("cannot create new instance for: %d", uint16(Id))
		return
	}
	err = json.Unmarshal(Information, group)
	return
}

///////////////////////////////////////////////////////////////////////////////

type Named struct{ Name string }

func (m Named) GetShortName() (name string) {
	if m.Name == "" {
		return "undefined"
	}
	return m.Name
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
func (m *Meta) Save() (Information []byte, err error) {
	type store struct {
		Name    string
		Records []Record
	}
	var rs []Record
	for _, g := range m.Groups {
		r := Record{Index: g.GetId(), Data: m.Groups}
		rs = append(rs, r)
	}
	return json.Marshal(&rs)
}
func (m *Meta) Add(g Group) {
	if g == nil {
		return
	}
	m.Groups = append(m.Groups, g)
}

///////////////////////////////////////////////////////////////////////////////
//
// var _ Group = new(Empty)
//
// type Empty struct{}
//
// func (Empty) GetShortName() (name string)       { return "empty" }
// func (Empty) Update() (nodes, elements *[]uint) { return nil, nil }
// func (Empty) GetWidget() vl.Widget              { return new(vl.Separator) }
// func (Empty) Save() (Id uint16, Information []byte, err error) { return saveGroup(m) }
//
///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NamedList)

type NamedList struct {
	Named
	Nodes, Elements []uint
}

func (m NamedList) GetId() GroupIndex                      { return NamedListIndex }
func (m *NamedList) Update() (nodes, elements *[]uint)     { return &m.Nodes, &m.Elements }
func (m *NamedList) Save() (Information []byte, err error) { return m.GetId().Save(m) }

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

func (m *NodeSupports) Save() (Information []byte, err error) { return m.GetId().Save(m) }

///////////////////////////////////////////////////////////////////////////////

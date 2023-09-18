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
	Data  string
}

// SaveGroup return stored information at json format
func (gi GroupIndex) Save(g Group) (_ []byte, err error) {
	b, err := json.Marshal(g)
	if err != nil {
		return
	}
	r := Record{Index: gi, Data: string(b)}
	return json.Marshal(&r)
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
	GetShortName() string              // return short name
	Update() (nodes, elements *[]uint) // update nodes, elements indexes
	GetWidget() vl.Widget              // return gui widget
	Save() (bs []byte, err error)
	Parse(bs []byte) (err error)
}

// ParseGroup return group from json
func ParseGroup(bs []byte) (group Group, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("ParseGroup: %v", err)
		}
	}()
	var r Record
	err = json.Unmarshal(bs, &r)
	if err != nil {
		return
	}
	Id := r.Index
	var ok bool
	group, ok = Id.newInstance()
	if !ok {
		err = fmt.Errorf("cannot create new instance for: %d", uint16(Id))
		return
	}
	err = group.Parse(bs)
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
func (m *Meta) Save() (bs []byte, err error) {
	var store struct {
		Index GroupIndex
		Name  string
		Data  []string
	}
	store.Index = m.GetId()
	for _, g := range m.Groups {
		var b []byte
		b, err = g.Save()
		if err != nil {
			return
		}
		store.Data = append(store.Data, string(b))
	}
	return json.Marshal(&store)
}
func (m *Meta) Parse(bs []byte) (err error) {
	var store struct {
		Index GroupIndex
		Name  string
		Data  []string
	}
	err = json.Unmarshal(bs, &store)
	if err != nil {
		return
	}
	m.Name = store.Name
	m.Groups = nil
	for i := range store.Data {
		var g Group
		g, err = ParseGroup([]byte(store.Data[i]))
		if err != nil {
			return
		}
		m.Groups = append(m.Groups, g)
	}
	return
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
// func (Empty) Save() (Id uint16, bs []byte, err error) { return saveGroup(m) }
//
///////////////////////////////////////////////////////////////////////////////

var _ Group = new(NamedList)

type NamedList struct {
	Named
	Nodes, Elements []uint
}

func (m NamedList) GetId() GroupIndex                  { return NamedListIndex }
func (m *NamedList) Update() (nodes, elements *[]uint) { return &m.Nodes, &m.Elements }
func (m *NamedList) Save() (bs []byte, err error)      { return m.GetId().Save(m) }
func (m *NamedList) Parse(bs []byte) (err error) {
	var r Record
	err = json.Unmarshal(bs, &r)
	if err != nil {
		return
	}
	if r.Index != m.GetId() {
		err = fmt.Errorf("not valid group index: %d != %d", r.Index, m.GetId())
		return
	}
	err = json.Unmarshal([]byte(r.Data), m)
	return 
}

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

func (m *NodeSupports) Save() (bs []byte, err error) { return m.GetId().Save(m) }
func (m *NodeSupports) Parse(bs []byte) (err error) {
	var r Record
	err = json.Unmarshal(bs, &r)
	if err != nil {
		return
	}
	if r.Index != m.GetId() {
		err = fmt.Errorf("not valid group index: %d != %d", r.Index, m.GetId())
		return
	}
	err = json.Unmarshal([]byte(r.Data), m)
	return 
}

///////////////////////////////////////////////////////////////////////////////

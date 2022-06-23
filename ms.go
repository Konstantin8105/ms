package ms

import (
	"fmt"

	"github.com/Konstantin8105/vl"
)

type GroupId uint8

const (
	Add GroupId = iota
	Split
	Plate
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
		return "Plate"
	case Plugin:
		return "Plugin"
	}
	return fmt.Sprintf("Undefined:%02d", g)
}

type Operation struct {
	Group GroupId
	Name  string
	Part  func() (w vl.Widget, action chan func())
}

var Operations = []Operation{
	{
		Group: Add,
		Name:  "Node by coordinate [X,Y,Z]",
		Part: func() (w vl.Widget, action chan func()) {
			return
		},
	},
}

var Debug []string

func UserInterface() (root vl.Widget, action chan func()) {
	var (
		scroll vl.Scroll
		list   vl.List
	)
	root = &scroll
	scroll.Root = &list

	//view := make([]bool, len(Operations))
	colHeader := make([]vl.CollapsingHeader, endGroup)
	for g := range colHeader {
		colHeader[g].SetText(GroupId(g).String())
		var sublist vl.List 
		colHeader[g].Root = &sublist
		list.Add(&colHeader[g])
	}
	for g := range colHeader {
		for i := range Operations {
			if Operations[i].Group != GroupId(g) {
				continue
			}
			var c vl.CollapsingHeader
			c.SetText(Operations[i].Name)
			r,a := Operations[i].Part()
			_ = a
			c.Root=r
			colHeader[g].Root.(*vl.List).Add(&c)
		}
	}
	return
}

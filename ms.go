package ms

import (
	"fmt"

	"github.com/Konstantin8105/tf"
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
			var list vl.List
			// coordinates
			w, gt := Input3Float(
				[3]string{"X", "Y", "Z"},
				[3]string{"meter", "meter", "meter"},
			)
			list.Add(w)
			// button
			var b vl.Button
			b.SetText("Insert")
			b.OnClick = func() {
				var vs [3]string
				for i := range vs {
					vs[i] = gt[i]()
				}
				Debug = append(Debug, fmt.Sprintf("Insert: %v", vs))
			}
			list.Add(&b)
			return &list, nil
		},
	}, {
		Group: Add,
		Name:  "Node at the line by distance",
		Part: func() (w vl.Widget, action chan func()) {
			return
		},
	},
}

func InputFloat(prefix, postfix string) (w vl.Widget, gettext func() string) {
	var (
		list vl.ListH
		in   vl.Inputbox
	)
	list.Add(vl.TextStatic(prefix))
	in.SetText("0.000")
	in.Filter(tf.Float)
	list.Add(&in)
	list.Add(vl.TextStatic(postfix))
	return &list, in.GetText
}

func Input3Float(prefix, postfix [3]string) (w vl.Widget, gettext [3]func() string) {
	var list vl.List
	for i := 0; i < 3; i++ {
		w, gt := InputFloat(prefix[i], postfix[i])
		list.Add(w)
		gettext[i] = gt
	}
	return &list, gettext
}

var Debug []string

func UserInterface() (root vl.Widget, action chan func(), err error) {
	var (
		scroll vl.Scroll
		list   vl.List
	)
	root = &scroll
	scroll.Root = &list
	action = make(chan func())

	view := make([]bool, len(Operations))
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
			part := Operations[i].Part
			if part == nil {
				err = fmt.Errorf("Widget %02d is empty\n", i)
				return
			}
			r, a := part()
			go func() {
				for {
					select {
					case f := <-a:
						action <- f
					}
				}
			}()
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
			err = fmt.Errorf("Do not view next operations: %v", nums)
		}
	}
	return
}

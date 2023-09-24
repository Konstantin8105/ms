package groups

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/vl"
)

const (
	testdata = "../testdata"
)

func Test(t *testing.T) {
	type tc struct {
		name  string
		group Group
	}
	var tcs []tc
	for i := uint16(0); i < math.MaxUint16; i++ {
		gi := GroupIndex(i)
		g, ok := gi.newInstance(nil)
		if !ok {
			continue
		}
		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d", i),
			group: g,
		})
	}

	var inits []func()
	{
		// example
		var m Meta
		m.Name = "example of Meta"
		inits = append(inits, func() { m.ID = 0 })

		var n NamedList
		n.Name = "lug"
		n.ID = 100
		n.Nodes = []uint{1, 2, 32, 576, 90, 98, 345, 234, 456, 5678, 7689, 46, 6}
		n.Elements = []uint{34, 67, 23, 53465, 65, 68, 23, 657, 9, 143, 231, 124}
		m.Groups = append(m.Groups, &n)
		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d_example", n.GetGroupIndex()),
			group: &n,
		})

		var c Copy
		c.Link = n.ID
		m.Groups = append(m.Groups, &c)
		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d_example", c.GetGroupIndex()),
			group: &c,
		})

		var s NodeSupports
		s.Name = "base support"
		s.Nodes = []uint{23, 52, 12, 23, 34, 456, 57, 68, 79, 14, 25, 36, 47, 58, 69}
		s.Direction = [6]bool{true, true, false, true, false, false}
		m.Groups = append(m.Groups, &s)
		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d_example", s.GetGroupIndex()),
			group: &s,
		})
		inits = append(inits, func() { s.ID = 0 })
		{
			var sub Meta
			sub.Name = "Submodel"
			var n NamedList
			n.Name = "Hole"
			n.Nodes = []uint{1, 2, 46, 6}
			n.Elements = []uint{34, 67, 231, 124}
			sub.Groups = append(sub.Groups, &n)
			m.Groups = append(m.Groups, &sub)
			inits = append(inits, func() { sub.ID = 0 })
		}

		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d_example", m.GetGroupIndex()),
			group: &m,
		})
	}

	for i := range tcs {
		for _, f := range inits {
			f()
		}
		name := fmt.Sprintf("%s.group", tcs[i].name)
		t.Run(name, func(t *testing.T) {
			mesh1 := GroupTest{base: tcs[i].group}
			FixMesh(&mesh1)

			// save
			bs, err := SaveGroup(mesh1.GetRootGroup())
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s", string(bs))
			name := filepath.Join(testdata, name)
			compare.Test(t, name, bs)

			// save
			{
				bs2, err := saveGroupNew(mesh1.GetRootGroup())
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("%s", string(bs2))
			}

			// parse
			gr, err := ParseGroup(bs)
			if err != nil {
				t.Fatal(err)
			}
			mesh2 := GroupTest{base: gr}
			FixMesh(&mesh2)

			if s1, s2 := fmt.Sprintf("%s", mesh1.GetRootGroup()),
				fmt.Sprintf("%s", mesh2.GetRootGroup()); s1 != s2 {
				t.Fatalf("not same after parse:\n%s\n%s", s1, s2)
			}
			// visualize
			{
				tr := treeNode(gr, mesh2, nil, nil)
				var sc vl.Screen
				sc.Root = &tr
				var cells [][]vl.Cell
				sc.SetHeight(20)
				sc.GetContents(50, &cells)
				compare.Test(t, name+".view", []byte(vl.Convert(cells)))
			}
			{
				var sc vl.Screen
				sc.Root = gr.GetWidget(nil)
				var cells [][]vl.Cell
				sc.SetHeight(20)
				sc.GetContents(50, &cells)
				compare.Test(t, name+".widgets", []byte(vl.Convert(cells)))
			}
		})
	}
}

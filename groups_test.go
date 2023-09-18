package ms

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/compare"
)

func TestGroupsSave(t *testing.T) {
	type tc struct {
		name  string
		group Group
	}
	var tcs []tc
	for i := uint16(0); i < math.MaxUint16; i++ {
		gi := GroupIndex(i)
		g, ok := gi.newInstance()
		if !ok {
			continue
		}
		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d", i),
			group: g,
		})
	}

	{
		// example
		var m Meta
		m.Name = "example of Meta"

		var n NamedList
		n.Name = "lug"
		n.Nodes = []uint{1, 2, 46, 6}
		n.Elements = []uint{34, 67, 231, 124}
		m.Groups = append(m.Groups, &n)

		var s NodeSupports
		s.Name = "base support"
		s.Nodes = []uint{23, 52, 12}
		s.Direction = [6]bool{true, true, false, true, false, false}
		m.Groups = append(m.Groups, &s)

		tcs = append(tcs, tc{
			name:  fmt.Sprintf("%06d_example", m.GetId()),
			group: &m,
		})
	}

	for i := range tcs {
		name := fmt.Sprintf("%s.group", tcs[i].name)
		t.Run(name, func(t *testing.T) {
			e := tcs[i].group
			// save
			bs, err := SaveGroup(e)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(testdata, name)
			compare.Test(t, name, bs)
			// parse
			gr, err := ParseGroup(bs)
			if err != nil {
				t.Fatal(err)
			}
			if s1, s2 := fmt.Sprintf("%s", e), fmt.Sprintf("%s", gr); s1 != s2 {
				t.Fatalf("not same after parse:\n%s\n%s", s1, s2)
			}
		})
	}
}

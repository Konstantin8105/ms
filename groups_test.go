package ms

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/compare"
)

func TestGroupsSave(t *testing.T) {
	var tcs []Group
	for i := uint16(0); i < math.MaxUint16; i++ {
		gi := GroupIndex(i)
		g, ok := gi.newInstance()
		if !ok {
			continue
		}
		tcs = append(tcs, g)
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

		tcs = append(tcs, &m)
	}

	for i := range tcs {
		name := fmt.Sprintf("%04d.group", i)
		t.Run(name, func(t *testing.T) {
			e := tcs[i]
			// save
			bs, err := e.Save()
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
			if fmt.Sprintf("%#v", e) != fmt.Sprintf("%#v", gr) {
				t.Fatalf("not same after parse")
			}
		})
	}
}

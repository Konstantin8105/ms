package ms

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/compare"
)

func TestGroupsSave(t *testing.T) {
	for i := range groups {
		t.Run(groups[i].name, func(t *testing.T) {
			// save
			e := groups[i].newInstance()
			id, info, err := e.Save()
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(testdata, fmt.Sprintf("%09d.group", id))
			compare.Test(t, name, info)
			// parse
			gr, err := parseGroup(id, info)
			if err != nil {
				t.Fatal(err)
			}
			_ = gr
		})
	}
}

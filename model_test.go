package ms

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/ms/groups"
)

func TestUniqUint(t *testing.T) {
	tcs := []struct {
		input  []uint
		expect []uint
	}{{
		input:  []uint{1, 2, 4, 2, 1, 4},
		expect: []uint{1, 2, 4},
	}, {
		input:  []uint{6, 1, 1, 1, 1, 1},
		expect: []uint{1, 6},
	}}
	for i := range tcs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			act := uniqUint(tcs[i].input)
			if len(act) != len(tcs[i].expect) {
				t.Fatalf("not same")
			}
			for p := range act {
				if act[p] != tcs[i].expect[p] {
					t.Fatalf("not equal")
				}
			}
		})
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/Konstantin8105/ms
// cpu: Intel(R) Xeon(R) CPU E3-1240 V2 @ 3.40GHz
// BenchmarkIntersection-4     100	  10671180 ns/op	   45137 B/op	     193 allocs/op
func BenchmarkIntersection(b *testing.B) {
	var mm Model
	mm.DemoSpiral(50)
	for n := 0; n < b.N; n++ {
		mm.SelectAll(true, []bool{true, true, true, true})
		els := mm.GetSelectElements(false, nil)
		ns := mm.GetSelectNodes(false)
		mm.Intersection(ns, els)
	}
}

const (
	testdata = "testdata"
)

func TestModel(t *testing.T) {
	tcs := []struct {
		name string
		mm   func() Model
	}{{
		name: filepath.Join(testdata, "IntersectionPointTriangle.ms"),
		mm: func() (mm Model) {
			var (
				L0   = mm.AddNode(0, 0, 0)
				L2   = mm.AddNode(0, 2, 0)
				R2   = mm.AddNode(2, 2, 0)
				_    = mm.AddNode(0, 1, 0)
				_    = mm.AddNode(1, 2, 0)
				_    = mm.AddNode(1, 1, 0)
				_, _ = mm.AddTriangle3ByNodeNumber(L0, L2, R2)
			)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionPointTriangle2.ms"),
		mm: func() (mm Model) {
			// point
			mm.AddNode(0, 0, 0)
			// triangle
			var (
				a0 = mm.AddNode(-1, -1, 0)
				a1 = mm.AddNode(+1, -1, 0)
				a2 = mm.AddNode(+0, +5, 0)
			)
			mm.AddTriangle3ByNodeNumber(a0, a1, a2)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionTriangleTriangle.ms"),
		mm: func() (mm Model) {
			var (
				a0   = mm.AddNode(-2.0, 0, -1)
				a1   = mm.AddNode(2.00, 0, -1)
				a2   = mm.AddNode(0.00, 0, 1.)
				b0   = mm.AddNode(-1.0, -1, 0)
				b1   = mm.AddNode(1.00, -1, 0)
				b2   = mm.AddNode(0.00, 1., 0)
				_, _ = mm.AddTriangle3ByNodeNumber(a0, a1, a2)
				_, _ = mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionTriangleTriangle2.ms"),
		mm: func() (mm Model) {
			// triangle
			var (
				a0 = mm.AddNode(-2, -1, 0)
				a1 = mm.AddNode(+2, -1, 0)
				a2 = mm.AddNode(+0, +2, 0)
			)
			mm.AddTriangle3ByNodeNumber(a0, a1, a2)
			// triangle
			var (
				b0 = mm.AddNode(+0, +0, 0)
				b1 = mm.AddNode(+3, +5, 0)
				b2 = mm.AddNode(-3, +5, 0)
			)
			mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionTriangleTriangle3.ms"),
		mm: func() (mm Model) {
			// triangle
			var (
				a0 = mm.AddNode(-2, -1, 0)
				a1 = mm.AddNode(+2, -1, 0)
				a2 = mm.AddNode(+0, +2, 0)
			)
			mm.AddTriangle3ByNodeNumber(a0, a1, a2)
			// triangle
			var (
				b0 = mm.AddNode(-2, +1, 0)
				b1 = mm.AddNode(+2, +1, 0)
				b2 = mm.AddNode(+0, -2, 0)
			)
			mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionLineTriangle.ms"),
		mm: func() (mm Model) {
			// line
			var (
				a0 = mm.AddNode(-2.0, 0, 0)
				a1 = mm.AddNode(2.00, 0, 0)
			)
			mm.AddLineByNodeNumber(a0, a1)
			// triangle
			var (
				b0 = mm.AddNode(0, -1, -3)
				b1 = mm.AddNode(0, -1, +3)
				b2 = mm.AddNode(0, +3, +0)
			)
			mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionLineTriangle2.ms"),
		mm: func() (mm Model) {
			// line
			var (
				a0 = mm.AddNode(0, -5, -5)
				a1 = mm.AddNode(0, +5, +5)
			)
			mm.AddLineByNodeNumber(a0, a1)
			// triangle
			var (
				b0 = mm.AddNode(0, -1, -3)
				b1 = mm.AddNode(0, -1, +3)
				b2 = mm.AddNode(0, +3, +0)
			)
			mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionLineQuadr.ms"),
		mm: func() (mm Model) {
			// line
			var (
				a0 = mm.AddNode(-2.0, 0, 0)
				a1 = mm.AddNode(2.00, 0, 0)
			)
			mm.AddLineByNodeNumber(a0, a1)
			// Quadr
			var (
				b0 = mm.AddNode(0, +4, +4)
				b1 = mm.AddNode(0, +4, -4)
				b2 = mm.AddNode(0, -4, -4)
				b3 = mm.AddNode(0, -4, +4)
			)
			mm.AddQuadr4ByNodeNumber(b0, b1, b2, b3)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionSpiral.ms"),
		mm: func() (mm Model) {
			mm.DemoSpiral(3)
			mm.SelectAll(true, []bool{true, true, true, true})
			els := mm.GetSelectElements(false, nil)
			mm.DeselectAll()
			mm.SplitLinesByEqualParts(els, 4)
			return mm
		},
	}, {
		name: filepath.Join(testdata, "IntersectionInsideTriangle.ms"),
		mm: func() (mm Model) {
			var (
				a0   = mm.AddNode(-2.0, -1.0, 0)
				a1   = mm.AddNode(+2.0, -1.0, 0)
				a2   = mm.AddNode(+0.0, +2.0, 0)
				b0   = mm.AddNode(-1.0, -0.5, 0)
				b1   = mm.AddNode(+1.0, -0.5, 0)
				b2   = mm.AddNode(+0.0, +1.0, 0)
				_, _ = mm.AddTriangle3ByNodeNumber(a0, a1, a2)
				_, _ = mm.AddTriangle3ByNodeNumber(b0, b1, b2)
			)
			return mm
		},
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mm := tc.mm()
			{ // intersection for each
				mm.SelectAll(true, []bool{true, true, true, true})
				els := mm.GetSelectElements(false, nil)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
			}

			_, err := json.MarshalIndent(mm, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			mm.filename = tc.name
			if err = mm.Save(); err != nil {
				t.Fatal(err)
			}

			if name := mm.GetPresentFilename(); name != tc.name {
				t.Fatalf("filenames is not same")
			}

			var mesh groups.GroupTest
			var o Model
			if err = o.Open(&mesh, tc.name); err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%#v", mm) != fmt.Sprintf("%#v", o) {
				t.Fatalf("Save-Open operations not same")
			}
		})
	}
}

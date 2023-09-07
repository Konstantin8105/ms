package ms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/ds"
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

func TestIntegration(t *testing.T) {
	// defer func() {
	// 	t.Logf("%s", PrintInfo())
	// }()

	// tests movements
	defer func() {
		testCoverageFunc = nil
	}()
	testCoverageFunc = func(mm Mesh, ch *chan ds.Action, screenshot func(filename string)) {

		var counter int
		var wg sync.WaitGroup
		run := func(name string, f func()) {
			logger.Printf(fmt.Sprintf("begin of %s", name))
			wg.Add(1)
			*ch <- func() (fus bool) {
				f()
				wg.Done()
				return true
			}
			wg.Wait()
			// time.Sleep(time.Second)
			logger.Printf(fmt.Sprintf("end of %s", name))

			// screenshoot
			screenshot(filepath.Join(
				testdata,
				fmt.Sprintf("%03d-%s.png", counter, name),
			))
			counter++
		}

		run("DemoSpiral", func() { mm.DemoSpiral(26) })
		run("DemoSpiral again", func() { mm.DemoSpiral(27) })
		run("StandardView", func() { mm.StandardView(StandardViewXOYpos) })
		run("SelectLeftCursor", func() { mm.SelectLeftCursor(true, []bool{true, true, true, true}) })
		run("SelectScreen", func() { mm.SelectScreen([2]int32{0, 0}, [2]int32{400, 300}) })

		// SelectElements
		{
			logger.Printf("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() (fus bool) {
				els = mm.GetSelectElements(Many, nil)
				wg.Done()
				return true
			}
			wg.Wait()
			if len(els) == 0 {
				logger.Printf("Error: SelectElements is zero")
				mm.Close()
				t.Fatalf("after select screen")
			}
		}

		run("color edge", func() { mm.ColorEdge(true) })
		run("color edge false", func() { mm.ColorEdge(false) })
		run("deselect", func() { mm.DeselectAll() })
		run("SelectLinesOrtho", func() { mm.SelectLinesOrtho(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, []bool{true, true, true, true}) })

		// SelectElements
		{
			logger.Printf("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() (fus bool) {
				els = mm.GetSelectElements(Many, nil)
				wg.Done()
				return true
			}
			wg.Wait()
			if len(els) == 0 {
				logger.Printf("Error: SelectElements is zero")
				mm.Close()
				t.Fatalf("after select screen")
			}

			// run("IgnoreModelElements", func() { mm.IgnoreModelElements(els) })
			// run("Unignore", func() { mm.Unignore() })
			run("SelectLinesOrtho", func() { mm.SelectLinesOrtho(true, true, true) })
			run("ColorEdge", func() { mm.ColorEdge(true) })
		}

		// Select elements
		{
			logger.Printf("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() (fus bool) {
				els = mm.GetSelectElements(Many, nil)
				wg.Done()
				return true
			}
			wg.Wait()
			if len(els) == 0 {
				logger.Printf("Error: SelectElements is zero")
				mm.Close()
				t.Fatalf("after select screen")
			}

			run("SplitLinesByRatio", func() { mm.SplitLinesByRatio(els, 0.25, false) })
			run("SplitLinesByRatio", func() { mm.SplitLinesByRatio(els, 2.25, false) })
			run("SplitLinesByEqualParts", func() { mm.SplitLinesByEqualParts(els, 10) })
		}

		run("MergeNodes", func() { mm.MergeNodes(0.050) })
		run("DeselectAll", func() { mm.DeselectAll() })
		run("SelectLinesOnPlane", func() { mm.SelectLinesOnPlane(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, []bool{true, true, true, true}) })

		{
			logger.Printf("SelectTriangles")
			var tris []uint
			wg.Add(1)
			*ch <- func() (fus bool) {
				tris = mm.GetSelectElements(Many, func(t ElType) bool {
					return t == Triangle3
				})
				wg.Done()
				return true
			}
			wg.Wait()
			if len(tris) == 0 {
				mm.Close()
				t.Fatalf("No 3")
			}

			run("SplitTri3To3Tri3", func() { mm.SplitTri3To3Tri3(tris) })
		}

		run("DeselectAll", func() { mm.DeselectAll() })
		run("SelectLinesOnPlane", func() { mm.SelectLinesOnPlane(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, []bool{false, true, true, true}) })

		{
			logger.Printf("SelectElements")
			var tris []uint
			wg.Add(1)
			*ch <- func() (fus bool) {
				tris = mm.GetSelectElements(Many, nil)
				wg.Done()
				return true
			}
			wg.Wait()
			if len(tris) == 0 {
				mm.Close()
				t.Fatalf("No 4")
			}
			run("MoveCopyNodesDistance", func() {
				mm.Copy(nil, tris,
					[3]float64{4, 0, 0},
					[]diffCoordinate{[6]float64{4, 0, 0, 0, 0, 0}},
					true, true)
			})
			time.Sleep(500 * time.Millisecond)
		}

		run("StandardView", func() { mm.StandardView(StandardViewXOYpos) })
		run("StandardView", func() { mm.StandardView(StandardViewXOZpos) })
		run("StandardView", func() { mm.StandardView(StandardViewYOZpos) })

		run("Check", func() {
			if err := mm.Check(); err != nil {
				t.Errorf("%v", err)
			}
		})
		run("RemoveNodesWithoutElements", func() { mm.RemoveNodesWithoutElements() })

		run("ColorEdge", func() { mm.ColorEdge(false) })
		for i := 0; i < 15; i++ {
			run("Undo", func() { mm.Undo() })
		}
		run("StandardView", func() { mm.StandardView(StandardViewXOZpos) })
		mm.Close()
	}
	// create a new model
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
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
	{
		old := IntersectionThreads
		IntersectionThreads = 1
		defer func() {
			IntersectionThreads = old
		}()
	}
	tcs := []struct {
		name string
		mm   func() Model
	}{
		{
			name: filepath.Join(testdata, "IntersectionPointTriangle.ms"),
			mm: func() Model {
				var (
					mm   Model
					L0   = mm.AddNode(0, 0, 0)
					L2   = mm.AddNode(0, 2, 0)
					R2   = mm.AddNode(2, 2, 0)
					_    = mm.AddNode(0, 1, 0)
					_    = mm.AddNode(1, 2, 0)
					_    = mm.AddNode(1, 1, 0)
					_, _ = mm.AddTriangle3ByNodeNumber(L0, L2, R2)
				)
				mm.SelectAll(true, []bool{true, true, true, true})
				els := mm.GetSelectElements(false, nil)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
		{
			name: filepath.Join(testdata, "IntersectionTriangleTriangle.ms"),
			mm: func() Model {
				var (
					mm   Model
					a0   = mm.AddNode(-2.0, 0, -1)
					a1   = mm.AddNode(2.00, 0, -1)
					a2   = mm.AddNode(0.00, 0, 1.)
					b0   = mm.AddNode(-1.0, -1, 0)
					b1   = mm.AddNode(1.00, -1, 0)
					b2   = mm.AddNode(0.00, 1., 0)
					_, _ = mm.AddTriangle3ByNodeNumber(a0, a1, a2)
					_, _ = mm.AddTriangle3ByNodeNumber(b0, b1, b2)
				)
				mm.SelectAll(true, []bool{true, true, true, true})
				els := mm.GetSelectElements(false, nil)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
		{
			name: filepath.Join(testdata, "IntersectionSpiral.ms"),
			mm: func() Model {
				var mm Model
				mm.DemoSpiral(3)
				mm.SelectAll(true, []bool{true, true, true, true})
				els := mm.GetSelectElements(false, nil)
				mm.DeselectAll()
				mm.SplitLinesByEqualParts(els, 4)
				mm.SelectAll(true, []bool{true, true, true, true})
				els = mm.GetSelectElements(false, nil)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
		{
			name: filepath.Join(testdata, "IntersectionInsideTriangle.ms"),
			mm: func() Model {
				var (
					mm   Model
					a0   = mm.AddNode(-2.0, -1.0, 0)
					a1   = mm.AddNode(+2.0, -1.0, 0)
					a2   = mm.AddNode(+0.0, +2.0, 0)
					b0   = mm.AddNode(-1.0, -0.5, 0)
					b1   = mm.AddNode(+1.0, -0.5, 0)
					b2   = mm.AddNode(+0.0, +1.0, 0)
					_, _ = mm.AddTriangle3ByNodeNumber(a0, a1, a2)
					_, _ = mm.AddTriangle3ByNodeNumber(b0, b1, b2)
				)
				mm.SelectAll(true, []bool{true, true, true, true})
				els := mm.GetSelectElements(false, nil)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mm := tc.mm()

			b, err := json.MarshalIndent(mm, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			compare.Test(t, tc.name, b)

			mm.filename = tc.name
			if err = mm.Save(); err != nil {
				t.Fatal(err)
			}

			if name := mm.GetPresentFilename(); name != tc.name {
				t.Fatalf("filenames is not same")
			}

			var o Model
			if err = o.Open(tc.name); err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%#v", mm) != fmt.Sprintf("%#v", o) {
				t.Fatalf("Save-Open operations not same")
			}
		})
	}
}

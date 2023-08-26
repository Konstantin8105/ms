package ms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func Example() {
	ResetInfo()
	defer ResetInfo()
	var mm Model
	{
		var c Coordinate
		c.Point3d[0] = 0
		c.Point3d[1] = 0
		c.Point3d[2] = 0
		mm.Coords = append(mm.Coords, c)
	}
	{
		var c Coordinate
		c.Point3d[0] = math.Pi
		c.Point3d[1] = 2
		c.Point3d[2] = 1
		mm.Coords = append(mm.Coords, c)
	}
	{
		var c Coordinate
		c.Point3d[0] = 6
		c.Point3d[1] = 5
		c.Point3d[2] = 2
		mm.Coords = append(mm.Coords, c)
	}
	mm.Elements = append(mm.Elements,
		Element{ElementType: Line2, Indexes: []int{0, 1}},
	)
	mm.IgnoreElements = append(mm.IgnoreElements, true)

	var p Part
	p.IgnoreElements = append(p.IgnoreElements, true, true)
	mm.Parts = append(mm.Parts, p)

	b, err := json.MarshalIndent(mm, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	// test
	var actual Model
	if err := json.Unmarshal(b, &actual); err != nil {
		fmt.Println(err)
		return
	}
	b2, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	if !bytes.Equal(b, b2) {
		fmt.Println("results are not same")
		return
	}

	// Output:
	// {
	//   "Name": "",
	//   "IgnoreElements": [
	//     true
	//   ],
	//   "Elements": [
	//     {
	//       "ElementType": 1,
	//       "Indexes": [
	//         0,
	//         1
	//       ]
	//     }
	//   ],
	//   "Coords": [
	//     {
	//       "Point3d": [
	//         0,
	//         0,
	//         0
	//       ],
	//       "Removed": false
	//     },
	//     {
	//       "Point3d": [
	//         3.141592653589793,
	//         2,
	//         1
	//       ],
	//       "Removed": false
	//     },
	//     {
	//       "Point3d": [
	//         6,
	//         5,
	//         2
	//       ],
	//       "Removed": false
	//     }
	//   ],
	//   "Parts": [
	//     {
	//       "Name": "",
	//       "IgnoreElements": [
	//         true,
	//         true
	//       ]
	//     }
	//   ]
	// }
}

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
	defer func() {
		t.Logf("%s", PrintInfo())
	}()

	AddInfo := func(v string) {
		t.Logf("%s", v)
		AddInfo(v)
	}

	// tests movements
	quit := make(chan struct{})
	testCoverageFunc = func(mm Mesh, ch *chan func()) {

		var wg sync.WaitGroup
		run := func(name string, f func()) {
			AddInfo(fmt.Sprintf("begin of %s", name))
			wg.Add(1)
			*ch <- func() {
				f()
				wg.Done()
			}
			wg.Wait()
			AddInfo(fmt.Sprintf("end of %s", name))
		}

		run("DemoSpiral", func() { mm.DemoSpiral(26) })
		run("DemoSpiral again", func() { mm.DemoSpiral(27) })
		run("SelectLeftCursor", func() { mm.SelectLeftCursor(true, true, true) })
		run("SelectScreen", func() { mm.SelectScreen([2]int32{0, 0}, [2]int32{400, 300}) })

		// SelectElements
		{
			AddInfo("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() {
				els = mm.GetSelectElements(Many)
				wg.Done()
			}
			wg.Wait()
			if len(els) == 0 {
				AddInfo("Error: SelectElements is zero")
				close(quit)
				t.Fatalf("after select screen")
			}
		}

		run("color edge", func() { mm.ColorEdge(true) })
		run("color edge false", func() { mm.ColorEdge(false) })
		run("deselect", func() { mm.DeselectAll() })
		run("SelectLinesOrtho", func() { mm.SelectLinesOrtho(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, true, true) })

		// SelectElements
		{
			AddInfo("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() {
				els = mm.GetSelectElements(Many)
				wg.Done()
			}
			wg.Wait()
			if len(els) == 0 {
				AddInfo("Error: SelectElements is zero")
				close(quit)
				t.Fatalf("after select screen")
			}

			run("IgnoreModelElements", func() { mm.IgnoreModelElements(els) })
			run("Unignore", func() { mm.Unignore() })
			run("SelectLinesOrtho", func() { mm.SelectLinesOrtho(true, true, true) })
			run("ColorEdge", func() { mm.ColorEdge(true) })
		}

		// Select elements
		{
			AddInfo("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() {
				els = mm.GetSelectElements(Many)
				wg.Done()
			}
			wg.Wait()
			if len(els) == 0 {
				AddInfo("Error: SelectElements is zero")
				close(quit)
				t.Fatalf("after select screen")
			}

			run("SplitLinesByRatio", func() { mm.SplitLinesByRatio(els, 0.25, false) })
			run("SplitLinesByRatio", func() { mm.SplitLinesByRatio(els, 2.25, false) })
			run("SplitLinesByEqualParts", func() { mm.SplitLinesByEqualParts(els, 10) })
		}

		run("MergeNodes", func() { mm.MergeNodes(0.050) })
		run("DeselectAll", func() { mm.DeselectAll() })
		run("SelectLinesOnPlane", func() { mm.SelectLinesOnPlane(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, true, true) })

		{
			AddInfo("SelectTriangles")
			var tris []uint
			wg.Add(1)
			*ch <- func() {
				tris = mm.GetSelectTriangles(Many)
				wg.Done()
			}
			wg.Wait()
			if len(tris) == 0 {
				close(quit)
				t.Fatalf("No 3")
			}

			run("SplitTri3To3Tri3", func() { mm.SplitTri3To3Tri3(tris) })
		}

		run("DeselectAll", func() { mm.DeselectAll() })
		run("SelectLinesOnPlane", func() { mm.SelectLinesOnPlane(true, true, true) })
		run("InvertSelect", func() { mm.InvertSelect(true, false, true) })

		{
			AddInfo("SelectElements")
			var tris []uint
			wg.Add(1)
			*ch <- func() {
				tris = mm.GetSelectElements(Many)
				wg.Done()
			}
			wg.Wait()
			if len(tris) == 0 {
				close(quit)
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

		run("ColorEdge", func() { mm.ColorEdge(false) })
		for i := 0; i < 15; i++ {
			run("Undo", func() { mm.Undo() })
		}
		run("StandardView", func() { mm.StandardView(StandardViewXOZpos) })
		close(quit)
	}
	// create a new model
	if err := Run("", quit); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func TestAddInfo(t *testing.T) {
	// defer func() {
	// 	t.Logf("%s", PrintInfo())
	// }()

	// tests movements
	quit := make(chan struct{})

	var wg sync.WaitGroup
	for i, size := 0, 1000; i < size; i++ {
		wg.Add(1)
		go func(cp int) {
			defer wg.Done()
			for i := 0; i < size; i++ {
				AddInfo(fmt.Sprintf("StandardView %02d.%02d", cp, i))
			}
		}(i)
	}

	testCoverageFunc = func(mm Mesh, ch *chan func()) {
		wg.Wait()
		// quit
		close(quit)
	}

	// create a new model
	if err := Run("", quit); err != nil {
		t.Error(err)
	}
}

func BenchmarkIntersection(b *testing.B) {
	var mm Model
	mm.DemoSpiral(50)
	for n := 0; n < b.N; n++ {
		mm.SelectAll(true, true, true)
		els := mm.GetSelectElements(false)
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
			name: "IntersectionPointTriangle",
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
				mm.SelectAll(true, true, true)
				els := mm.GetSelectElements(false)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
		{
			name: "IntersectionTriangleTriangle",
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
				mm.SelectAll(true, true, true)
				els := mm.GetSelectElements(false)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
		{
			name: "IntersectionSpiral",
			mm: func() Model {
				var mm Model
				mm.DemoSpiral(3)
				mm.SelectAll(true, true, true)
				els := mm.GetSelectElements(false)
				mm.DeselectAll()
				mm.SplitLinesByEqualParts(els, 4)
				mm.SelectAll(true, true, true)
				els = mm.GetSelectElements(false)
				ns := mm.GetSelectNodes(false)
				mm.Intersection(ns, els)
				return mm
			},
		},
	}

	compare := func(name string, actual []byte) {
		var (
			filename = filepath.Join(testdata, name)
		)
		// for update test screens run in console:
		// UPDATE=true go test
		if os.Getenv("UPDATE") == "true" {
			if err := ioutil.WriteFile(filename, actual, 0644); err != nil {
				t.Fatalf("Cannot write snapshot to file: %v", err)
			}
		}
		// get expect result
		expect, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("Cannot read snapshot file: %v", err)
		}
		// compare
		if !bytes.Equal(actual, expect) {
			f2 := filename + ".new"
			if err := ioutil.WriteFile(f2, actual, 0644); err != nil {
				t.Fatalf("Cannot write snapshot to file new: %v", err)
			}
			size := 1000
			if size < len(actual) {
				actual = actual[:size]
			}
			if size < len(expect) {
				expect = expect[:size]
			}
			t.Errorf("Snapshots is not same:\nActual:\n%s\nExpect:\n%s\nmeld %s %s",
				actual,
				expect,
				filename, f2,
			)
		}
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ResetInfo()
			defer ResetInfo()
			mm := tc.mm()

			b, err := json.MarshalIndent(mm, "", "  ")
			if err != nil {
				fmt.Println(err)
				return
			}

			t.Logf("%s\n", PrintInfo())
			compare(tc.name, b)
		})
	}
}

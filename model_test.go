package ms

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/gog"
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

func TestIntegration(t *testing.T) {
	// defer func() {
	// 	t.Logf("%s", PrintInfo())
	// }()

	const reset = "Reset"

	// tests movements
	defer func() {
		testCoverageFunc = nil
	}()
	testCoverageFunc = func(mm Mesh, ch *chan ds.Action, screenshot func(func(image.Image))) {

		var counter int
		var testCounter int
		testHeader := func() {
			logger.Printf("============== TEST %02d ==============", testCounter)
			testCounter++
			counter = 0
		}

		var wg sync.WaitGroup
		run := func(name string, f func()) {
			if name != reset {
				logger.Printf(fmt.Sprintf("begin of %s", name))
			}
			wg.Add(1)
			*ch <- func() (fus bool) {
				f()
				wg.Done()
				return true
			}
			wg.Wait()
			// time.Sleep(time.Second)
			if name != reset {
				logger.Printf(fmt.Sprintf("end of %s", name))
			}

			// screenshoot
			if name == reset {
				return
			}
			name = strings.ReplaceAll(name, " ", "_")
			name = filepath.Join(
				testdata,
				fmt.Sprintf("Test%03d-%03d-%s.png", testCounter, counter, name),
			)
			screenshot(
				func(actual image.Image) {
					compare.TestPng(t, name, actual)
				},
			)
			counter++
		}

		clean := func() {
			run(reset, func() { mm.ColorEdge(false) })
			for i := 0; i < 45; i++ {
				run(reset, func() { mm.Undo() })
			}
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
		}

		// test
		testHeader()
		run("DemoSpiral", func() { mm.DemoSpiral(26) })
		run(reset, func() { mm.StandardView(StandardViewXOYpos) })
		run("DemoSpiral again", func() { mm.DemoSpiral(27) })
		run(reset, func() { mm.StandardView(StandardViewXOYpos) })
		run("StandardView", func() { mm.StandardView(StandardViewXOYpos) })
		run("SelectLeftCursor", func() { mm.SelectLeftCursor(true, []bool{true, true, true, true}) })
		run("SelectScreen", func() { mm.SelectScreen([2]int32{0, 0}, [2]int32{400, 300}) })
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
		run("SelectElementsOnPlane", func() { mm.SelectElementsOnPlane(true, true, true, []bool{true, true, true, true}) })
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
		run("SelectElementsOnPlane", func() { mm.SelectElementsOnPlane(true, true, true, []bool{true, true, true, true}) })
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

		// test
		clean()
		{
			testHeader()
			run("DemoSpiral", func() { mm.DemoSpiral(20) })
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("SelectAll", func() { mm.SelectAll(true, []bool{false, false, false, false, false}) })

			logger.Printf("SelectNodes")
			wg.Add(1)
			var ns []uint
			*ch <- func() (fus bool) {
				ns = mm.GetSelectNodes(Many)
				wg.Done()
				return true
			}
			wg.Wait()
			if len(ns) == 0 {
				logger.Printf("Error: GetSelectNodes is zero")
				mm.Close()
				t.Fatalf("after select screen")
			}

			run("DeselectAll", func() { mm.DeselectAll() })
			run("Move half points", func() {
				mm.Move(
					ns[len(ns)/2:], nil,
					[3]float64{0, 0, 0},
					[6]float64{1, 1}, // diffCoordinate
				)
			})
			run("Remove half points", func() {
				mm.Remove(ns[:len(ns)/2], nil)
			})
			run("RemoveSameCoordinates", func() { mm.RemoveSameCoordinates() })
			run("Copy", func() {
				var diff diffCoordinate = [6]float64{-1, -1, -1}
				mm.Copy(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[]diffCoordinate{diff},
					false, false,
				)
			})
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("Move", func() {
				mm.Move(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[6]float64{1, 1, 1},
				)
			})
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("RemoveSameCoordinates", func() { mm.RemoveSameCoordinates() })

			run("Copy with intermediant parts", func() {
				var diff diffCoordinate = [6]float64{-1, -1, -1}
				mm.Copy(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[]diffCoordinate{diff},
					true, true,
				)
			})
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("Move", func() {
				mm.Move(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[6]float64{1, 1, 1},
				)
			})
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("RemoveSameCoordinates", func() { mm.RemoveSameCoordinates() })
		}

		// test
		clean()
		{
			testHeader()
			run("DemoSpiral", func() { mm.DemoSpiral(10) })
			run("XOYpos", func() { mm.StandardView(StandardViewXOYpos) })
			run("YOZpos", func() { mm.StandardView(StandardViewYOZpos) })
			run("XOZpos", func() { mm.StandardView(StandardViewXOZpos) })
			run("XOYneg", func() { mm.StandardView(StandardViewXOYneg) })
			run("YOZneg", func() { mm.StandardView(StandardViewYOZneg) })
			run("XOZneg", func() { mm.StandardView(StandardViewXOZneg) })
		}

		// test
		clean()
		{
			testHeader()
			run(reset, func() { mm.DemoSpiral(10) })
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("ViewAll", func() { mm.ViewAll() })
			var id uint
			run("AddNode", func() { id = mm.AddNode(0, 0, 0) })
			run("SelectLinesCylindrical - radiant", func() {
				mm.SelectLinesCylindrical(
					id,
					true, false,
					DirY,
				)
			})
			var ns, es []uint
			run(reset, func() {
				ns = mm.GetSelectNodes(false)
				es = mm.GetSelectElements(false, nil)
			})
			run("Hide", func() { mm.Hide(ns, es) })
			run("SelectLinesCylindrical - conc", func() {
				mm.SelectLinesCylindrical(
					id,
					false, true,
					DirY,
				)
			})
			run(reset, func() {
				ns = mm.GetSelectNodes(false)
				es = mm.GetSelectElements(false, nil)
			})
			run("Hide", func() { mm.Hide(ns, es) })
			run("Unhide", func() { mm.UnhideAll() })
		}
		// test
		clean()
		{
			testHeader()
			run(reset, func() { mm.DemoSpiral(10) })
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("ViewAll", func() { mm.ViewAll() })
			run("SelectLinesParallel", func() {
				var lines []uint
				els := mm.GetElements()
				ns := mm.GetCoords()
				for i := len(els) - 1; 0 <= i; i-- {
					if els[i].ElementType != Line2 {
						continue
					}
					if 1e-3 < math.Abs(ns[els[i].Indexes[0]].Point3d[1]-ns[els[i].Indexes[1]].Point3d[1]) {
						continue
					}
					lines = append(lines, uint(i))
					break
				}
				mm.SelectLinesParallel(lines)
			})
			run("InvertSelect", func() {
				mm.InvertSelect(true, []bool{true, true, true, true})
			})
			var ns, es []uint
			run(reset, func() {
				ns = mm.GetSelectNodes(false)
				es = mm.GetSelectElements(false, nil)
			})
			run("Hide", func() { mm.Hide(ns, es) })
			run("SelectAll", func() {
				mm.SelectAll(true, []bool{true, true, true, true})
			})

			run(reset, func() {
				es = mm.GetSelectElements(false, func(e ElType) bool {
					return e == Line2
				})
			})
			run("SplitLinesByEqualParts", func() {
				mm.SplitLinesByEqualParts(es, 10)
				if err := mm.Check(); err != nil {
					t.Fatal(err)
				}
			})

			run("SelectAll", func() {
				mm.SelectAll(true, []bool{true, true, true, true})
			})
			run(reset, func() {
				es = mm.GetSelectElements(false, func(e ElType) bool {
					return e == Line2
				})
			})
			run("DeselectAll", func() { mm.DeselectAll() })
			run("MergeLines", func() { mm.MergeLines(es) })
			run("RemoveNodesWithoutElements", func() { mm.RemoveNodesWithoutElements() })
		}

		// test
		clean()
		{
			testHeader()
			run(reset, func() { mm.DemoSpiral(10) })
			run(reset, func() { mm.StandardView(StandardViewXOYpos) })
			run("ViewAll", func() { mm.ViewAll() })
			run("SelectAll", func() {
				mm.SelectAll(true, []bool{true, true, true, true})
			})
			var ns, es []uint
			run(reset, func() {
				ns = mm.GetSelectNodes(false)
				es = mm.GetSelectElements(false, nil)
			})

			run("Mirror-copy-all", func() {
				mm.Mirror(ns, es,
					[3]gog.Point3d{
						[3]float64{0, 0, 0},
						[3]float64{1, 0, 0},
						[3]float64{0, 0, 1},
					},
					true,
					true, true,
				)
			})
			run("ViewAll", func() { mm.ViewAll() })
			run("Undo", func() { mm.Undo() })

			run("Mirror-copy-only tri", func() {
				mm.Mirror(ns, es,
					[3]gog.Point3d{
						[3]float64{0, 0, 0},
						[3]float64{1, 0, 0},
						[3]float64{0, 0, 1},
					},
					true,
					false, true,
				)
			})
			run("ViewAll", func() { mm.ViewAll() })
			run("Undo", func() { mm.Undo() })

			run("Mirror-copy-only lines", func() {
				mm.Mirror(ns, es,
					[3]gog.Point3d{
						[3]float64{0, 0, 0},
						[3]float64{1, 0, 0},
						[3]float64{0, 0, 1},
					},
					true,
					true, false,
				)
			})
			run("ViewAll", func() { mm.ViewAll() })
			run("Undo", func() { mm.Undo() })

			run("Mirror-move", func() {
				mm.Mirror(ns, es,
					[3]gog.Point3d{
						[3]float64{0, 0, 0},
						[3]float64{1, 0, 0},
						[3]float64{0, 0, 1},
					},
					false,
					false, false,
				)
			})
			run("ViewAll", func() { mm.ViewAll() })

			run("Mirror-copy-all-any-plane", func() {
				mm.Mirror(ns, es,
					[3]gog.Point3d{
						[3]float64{0, 1, 0},
						[3]float64{1, 3, 0},
						[3]float64{0, 2, 1},
					},
					true,
					false, true,
				)
			})
			run("ViewAll", func() { mm.ViewAll() })
			run("Undo", func() { mm.Undo() })
		}
		// MergeLines
		// IsChangedModel
		// GetPresentFilename
		// Save
		// AddNode
		// AddLineByNodeNumber
		// AddTriangle3ByNodeNumber
		// GetCoordByID
		// AddLeftCursor

		// close model
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

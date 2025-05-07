//go:build ignore

package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/ms"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "logger: ", log.Ltime|log.Llongfile)
}

const (
	testdata = "testdata"
)

type checker struct {
	iserror bool
	err     error
}

func (c *checker) Errorf(format string, args ...any) {
	c.iserror = true
	c.err = fmt.Errorf(format, args...)
	panic(c.err)
}

func (c *checker) Fatalf(format string, args ...any) {
	c.Errorf(format, args...)
}

func main() {
	t := new(checker)

	// defer func() {
	// 	t.Logf("%s", PrintInfo())
	// }()

	const reset = "Reset"

	// tests movements
	defer func() {
		ms.TestCoverageFunc = nil
	}()
	ms.TestCoverageFunc = func(mm ms.Mesh, ch *chan ds.Action, screenshot func(func(image.Image))) {

		var counter int
		var testCounter, subTestCounter int
		testHeader := func() {
			logger.Printf("=== TEST %02d ===================", testCounter)
			testCounter++
			subTestCounter = 0
			counter = 0
		}

		var wg sync.WaitGroup
		run := func(name string, f func()) {
			logger.Printf("=== TEST %02d. SUB TEST %02d. ===", testCounter, subTestCounter)
			subTestCounter++
			if name != reset {
				logger.Printf("begin of %s", name)
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
				logger.Printf("end of %s", name)
			} else {
				// screenshoot
				logger.Printf("reset")
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
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
		}

		// test
		testHeader()
		run("DemoSpiral", func() { mm.DemoSpiral(26) })
		run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
		run("DemoSpiral again", func() { mm.DemoSpiral(27) })
		run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
		run("StandardView", func() { mm.StandardView(ms.StandardViewXOYpos) })
		run("SelectLeftCursor", func() { mm.SelectLeftCursor(true, []bool{true, true, true, true}) })
		run("SelectScreen", func() { mm.SelectScreen([2]int32{0, 0}, [2]int32{400, 300}) })
		{
			logger.Printf("SelectElements")
			wg.Add(1)
			var els []uint
			*ch <- func() (fus bool) {
				els = mm.GetSelectElements(ms.Many, nil)
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
				els = mm.GetSelectElements(ms.Many, nil)
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
				els = mm.GetSelectElements(ms.Many, nil)
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
				tris = mm.GetSelectElements(ms.Many, func(t ms.ElType) bool {
					return t == ms.Triangle3
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
				tris = mm.GetSelectElements(ms.Many, nil)
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
					[]ms.DiffCoordinate{[6]float64{4, 0, 0, 0, 0, 0}},
					true, true)
			})
			time.Sleep(500 * time.Millisecond)
		}

		run("StandardView", func() { mm.StandardView(ms.StandardViewXOYpos) })
		run("StandardView", func() { mm.StandardView(ms.StandardViewXOZpos) })
		run("StandardView", func() { mm.StandardView(ms.StandardViewYOZpos) })

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
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("SelectAll", func() { mm.SelectAll(true, []bool{false, false, false, false, false}) })

			logger.Printf("SelectNodes")
			wg.Add(1)
			var ns []uint
			*ch <- func() (fus bool) {
				ns = mm.GetSelectNodes(ms.Many)
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
				var diff ms.DiffCoordinate = [6]float64{-1, -1, -1}
				mm.Copy(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[]ms.DiffCoordinate{diff},
					false, false,
				)
			})
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("Move", func() {
				mm.Move(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[6]float64{1, 1, 1},
				)
			})
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("RemoveSameCoordinates", func() { mm.RemoveSameCoordinates() })

			run("Copy with intermediant parts", func() {
				var diff ms.DiffCoordinate = [6]float64{-1, -1, -1}
				mm.Copy(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[]ms.DiffCoordinate{diff},
					true, true,
				)
			})
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("Move", func() {
				mm.Move(
					nil, []uint{uint(len(mm.GetElements()) - 1)},
					[3]float64{0, 0, 0},
					[6]float64{1, 1, 1},
				)
			})
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("RemoveSameCoordinates", func() { mm.RemoveSameCoordinates() })
		}

		// test
		clean()
		{
			testHeader()
			run("DemoSpiral", func() { mm.DemoSpiral(10) })
			run("XOYpos", func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("YOZpos", func() { mm.StandardView(ms.StandardViewYOZpos) })
			run("XOZpos", func() { mm.StandardView(ms.StandardViewXOZpos) })
			run("XOYneg", func() { mm.StandardView(ms.StandardViewXOYneg) })
			run("YOZneg", func() { mm.StandardView(ms.StandardViewYOZneg) })
			run("XOZneg", func() { mm.StandardView(ms.StandardViewXOZneg) })
		}

		// test
		clean()
		{
			testHeader()
			run(reset, func() { mm.DemoSpiral(10) })
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("ViewAll", func() { mm.ViewAll() })
			var id uint
			run("AddNode", func() { id = mm.AddNode(0, 0, 0) })
			run("SelectLinesCylindrical - radiant", func() {
				mm.SelectLinesCylindrical(
					id,
					true, false,
					ms.DirY,
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
					ms.DirY,
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
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
			run("ViewAll", func() { mm.ViewAll() })
			run("SelectLinesParallel", func() {
				var lines []uint
				els := mm.GetElements()
				ns := mm.GetCoords()
				for i := len(els) - 1; 0 <= i; i-- {
					if els[i].ElementType != ms.Line2 {
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
				es = mm.GetSelectElements(false, func(e ms.ElType) bool {
					return e == ms.Line2
				})
			})
			run("SplitLinesByEqualParts", func() {
				mm.SplitLinesByEqualParts(es, 10)
				if err := mm.Check(); err != nil {
					t.Fatalf("%v", err)
				}
			})

			run("SelectAll", func() {
				mm.SelectAll(true, []bool{true, true, true, true})
			})
			run(reset, func() {
				es = mm.GetSelectElements(false, func(e ms.ElType) bool {
					return e == ms.Line2
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
			run(reset, func() { mm.StandardView(ms.StandardViewXOYpos) })
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
	if err := ms.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

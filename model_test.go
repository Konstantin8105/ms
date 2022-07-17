package ms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

func Example() {
	var mm Model
	mm.Coords = append(mm.Coords,
		Coordinate{X: 0, Y: 0, Z: 0},
		Coordinate{X: math.Pi, Y: 2, Z: 1},
		Coordinate{X: 6, Y: 5, Z: 4},
	)
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
	//       "Removed": false,
	//       "X": 0,
	//       "Y": 0,
	//       "Z": 0
	//     },
	//     {
	//       "Removed": false,
	//       "X": 3.141592653589793,
	//       "Y": 2,
	//       "Z": 1
	//     },
	//     {
	//       "Removed": false,
	//       "X": 6,
	//       "Y": 5,
	//       "Z": 4
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
		for i := range Debug {
			t.Logf("%s", Debug[i])
		}
	}()

	print := func(args ...interface{}) {
		Debug = append(Debug, fmt.Sprintln(args...))
	}

	// tests movements
	quit := make(chan struct{})
	testCoverageFunc = func(mm Mesh) {
		// draw spiral
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		print("add DemoSpiral")
		// draw clone spiral
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		print("add DemoSpiral again")
		// select
		<-time.After(300 * time.Millisecond)
		mm.SelectLeftCursor(true, true, true)
		print("change SelectLeftCursor")
		<-time.After(300 * time.Millisecond)
		print("SelectScreen")
		mm.SelectScreen([2]int32{0, 0}, [2]int32{600, 400})
		// {
		// 	els := mm.SelectElements(Many)
		// 	print("SelectElements")
		// 	if len(els) == 0 {
		// 		close(quit)
		// 		t.Fatalf("after select screen")
		// 	}
		// }
		<-time.After(1 * time.Second)
		print("select screen")
		// color change
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(true)
		print("add colors")
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(false)
		print("false colors")
		// deselect
		<-time.After(1 * time.Second)
		mm.DeselectAll()
		print("deselect")
		// select ortho
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOrtho(true, true, true)
		print("SelectLinesOrtho")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, true, true)
		print("InvertSelect")
		<-time.After(300 * time.Millisecond)
		els := mm.SelectElements(Many)
		print("SelectElements")
		if len(els) == 0 {
			close(quit)
			t.Fatalf("No 1")
		}
		// IgnoreElements
		<-time.After(300 * time.Millisecond)
		mm.IgnoreModelElements(els)
		print("IgnoreModelElements")
		<-time.After(300 * time.Millisecond)
		mm.Unignore()
		print("Unignore")
		// split lines
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOrtho(true, true, true)
		print("SelectLinesOrtho")
		<-time.After(300 * time.Millisecond)
		els = mm.SelectElements(Many)
		print("SelectElements")
		if len(els) == 0 {
			close(quit)
			t.Fatalf("No 2")
		}
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(true) // color
		print("add colors")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByRatio(els, 0.25, false)
		print("SplitLinesByRatio")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByRatio(els, 2.25, false)
		print("SplitLinesByRatio")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByEqualParts(els, 10)
		print("SplitLinesByEqualParts")
		// merge
		<-time.After(300 * time.Millisecond)
		mm.MergeNodes(0.050)
		print("MergeNodes")
		// deselect
		<-time.After(300 * time.Millisecond)
		mm.DeselectAll()
		print("DeselectAll")

		// select
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOnPlane(true, true, true)
		print("SelectLinesOnPlane")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, true, true)
		print("InvertSelect")
		<-time.After(300 * time.Millisecond)
		tris := mm.SelectTriangles(Many)
		print("SelectTriangles")
		if len(tris) == 0 {
			close(quit)
			t.Fatalf("No 3")
		}
		// selectObjects.fromAdd = false
		<-time.After(300 * time.Millisecond)
		mm.SplitTri3To3Tri3(tris)
		print("SplitTri3To3Tri3")

		// deselect
		<-time.After(300 * time.Millisecond)
		mm.DeselectAll()
		print("DeselectAll")
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOnPlane(true, true, true)
		print("SelectLinesOnPlane")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, false, true)
		print("InvertSelect")
		<-time.After(300 * time.Millisecond)
		tris = mm.SelectElements(Many)
		print("SelectElements")
		if len(tris) == 0 {
			close(quit)
			t.Fatalf("No 4")
		}
		<-time.After(300 * time.Millisecond)
		mm.MoveCopyNodesDistance(nil, tris, [3]float64{4, 0, 0},
			true, true, true)
		print("MoveCopyNodesDistance")
		// view
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewXOYpos)
		print("StandardView")
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewXOZpos)
		print("StandardView")
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewYOZpos)
		print("StandardView")
		// undo
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(false)
		<-time.After(300 * time.Millisecond)
		for i := 0; i < 15; i++ {
			mm.Undo()
			print("Undo")
			<-time.After(2 * time.Second)
		}
		// view
		mm.StandardView(StandardViewXOZpos)
		print("StandardView")
		// quit
		<-time.After(2 * time.Second)
		close(quit)
	}
	// create a new model
	if err := Run(quit); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

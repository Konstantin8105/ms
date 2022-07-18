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
		t.Logf("%s", PrintInfo())
	}()

	// tests movements
	quit := make(chan struct{})
	testCoverageFunc = func(mm Mesh) {
		// draw spiral
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		AddInfo("add DemoSpiral")
		// draw clone spiral
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		AddInfo("add DemoSpiral again")
		// select
		<-time.After(300 * time.Millisecond)
		mm.SelectLeftCursor(true, true, true)
		AddInfo("change SelectLeftCursor")
		<-time.After(300 * time.Millisecond)
		AddInfo("SelectScreen")
		mm.SelectScreen([2]int32{0, 0}, [2]int32{400, 300})
		<-time.After(2 * time.Second)
		{
			els := mm.SelectElements(Many)
			AddInfo("SelectElements")
			if len(els) == 0 {
				AddInfo("Error: SelectElements is zero")
				close(quit)
				t.Fatalf("after select screen")
			}
		}
		<-time.After(1 * time.Second)
		AddInfo("select screen")
		// color change
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(true)
		AddInfo("add colors")
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(false)
		AddInfo("false colors")
		// deselect
		<-time.After(1 * time.Second)
		mm.DeselectAll()
		AddInfo("deselect")
		// select ortho
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOrtho(true, true, true)
		AddInfo("SelectLinesOrtho")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, true, true)
		AddInfo("InvertSelect")
		<-time.After(300 * time.Millisecond)
		els := mm.SelectElements(Many)
		AddInfo("SelectElements")
		if len(els) == 0 {
			close(quit)
			t.Fatalf("No 1")
		}
		// IgnoreElements
		<-time.After(300 * time.Millisecond)
		mm.IgnoreModelElements(els)
		AddInfo("IgnoreModelElements")
		<-time.After(300 * time.Millisecond)
		mm.Unignore()
		AddInfo("Unignore")
		// split lines
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOrtho(true, true, true)
		AddInfo("SelectLinesOrtho")
		<-time.After(300 * time.Millisecond)
		els = mm.SelectElements(Many)
		AddInfo("SelectElements")
		if len(els) == 0 {
			close(quit)
			t.Fatalf("No 2")
		}
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(true) // color
		AddInfo("add colors")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByRatio(els, 0.25, false)
		AddInfo("SplitLinesByRatio")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByRatio(els, 2.25, false)
		AddInfo("SplitLinesByRatio")
		<-time.After(300 * time.Millisecond)
		mm.SplitLinesByEqualParts(els, 10)
		AddInfo("SplitLinesByEqualParts")
		// merge
		<-time.After(300 * time.Millisecond)
		mm.MergeNodes(0.050)
		AddInfo("MergeNodes")
		// deselect
		<-time.After(300 * time.Millisecond)
		mm.DeselectAll()
		AddInfo("DeselectAll")

		// select
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOnPlane(true, true, true)
		AddInfo("SelectLinesOnPlane")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, true, true)
		AddInfo("InvertSelect")
		<-time.After(300 * time.Millisecond)
		tris := mm.SelectTriangles(Many)
		AddInfo("SelectTriangles")
		if len(tris) == 0 {
			close(quit)
			t.Fatalf("No 3")
		}
		// selectObjects.fromAdd = false
		<-time.After(300 * time.Millisecond)
		mm.SplitTri3To3Tri3(tris)
		AddInfo("SplitTri3To3Tri3")

		// deselect
		<-time.After(300 * time.Millisecond)
		mm.DeselectAll()
		AddInfo("DeselectAll")
		<-time.After(300 * time.Millisecond)
		mm.SelectLinesOnPlane(true, true, true)
		AddInfo("SelectLinesOnPlane")
		<-time.After(300 * time.Millisecond)
		mm.InvertSelect(true, false, true)
		AddInfo("InvertSelect")
		<-time.After(300 * time.Millisecond)
		tris = mm.SelectElements(Many)
		AddInfo("SelectElements")
		if len(tris) == 0 {
			close(quit)
			t.Fatalf("No 4")
		}
		<-time.After(300 * time.Millisecond)
		mm.MoveCopyNodesDistance(nil, tris, [3]float64{4, 0, 0},
			true, true, true)
		AddInfo("MoveCopyNodesDistance")
		// view
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewXOYpos)
		AddInfo("StandardView")
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewXOZpos)
		AddInfo("StandardView")
		<-time.After(1 * time.Second)
		mm.StandardView(StandardViewYOZpos)
		AddInfo("StandardView")
		// undo
		<-time.After(300 * time.Millisecond)
		mm.ColorEdge(false)
		<-time.After(300 * time.Millisecond)
		for i := 0; i < 15; i++ {
			mm.Undo()
			AddInfo("Undo")
			<-time.After(2 * time.Second)
		}
		// view
		mm.StandardView(StandardViewXOZpos)
		AddInfo("StandardView")
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

package ms

import (
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
			fmt.Println(Debug[i])
		}
	}()
	// Model
	var mm Model
	// tests movements
	quit := make(chan struct{})
	go func() {
		// draw spiral
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		// draw clone spiral
		<-time.After(100 * time.Millisecond)
		mm.DemoSpiral()
		// select
		<-time.After(100 * time.Millisecond)
		selectObjects.xFrom = 0
		selectObjects.yFrom = 0
		selectObjects.fromAdd = true
		selectObjects.xFrom = 600
		selectObjects.yFrom = 400
		selectObjects.toUpdate = true
		selectObjects.toAdd = true
		// color change
		<-time.After(100 * time.Millisecond)
		mm.ColorEdge(true)
		<-time.After(100 * time.Millisecond)
		mm.ColorEdge(false)
		// deselect
		<-time.After(100 * time.Millisecond)
		mm.DeselectAll()
		// quit
		<-time.After(2 * time.Second)
		close(quit)
	}()
	// create a new model
	if err := mm.Run(quit); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

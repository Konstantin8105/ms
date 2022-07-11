package ms

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
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
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
			debug.PrintStack()
		}
		for i := range Debug {
			fmt.Println(Debug[i])
		}
	}()

	root, action, err := UserInterface()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	go func() {
		if err := M3(); err != nil {
			fmt.Fprintf(os.Stderr, "M3: %v", err)
		}
	}()

	quit := make(chan struct{})

	go func() {
		<-time.After(500 * time.Millisecond)
		mm.DemoSpiral()
		<-time.After(2 * time.Second)
		close(quit)
	}()

	err = vl.Run(root, action, quit, tcell.KeyCtrlC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

package ms

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Konstantin8105/pow"
)

// 3D model variables
type object3d struct {
	selected bool
	hided    bool
}

type ElType uint8

const (
	Line2 ElType = iota + 1
	Triangle3
)

// Element is typical element for FEM. Examples:
//
//	Line o======o                                             //
//	ElType : 2                                                //
//	Indexes: 2 (amount indexes of coordinates)                //
//
//	Triangle o======o                                         //
//	          \    /                                          //
//	           \  /                                           //
//	            o                                             //
//	ElType : 3                                                //
//	Indexes: 3 (amount indexes of coordinates)                //
//
//	Quadr4 o======o                                           //
//	       |      |                                           //
//	       |      |                                           //
//	       o======o                                           //
//	ElType : 4                                                //
//	Indexes: 4 (amount indexes of coordinates)                //
//
type Element struct {
	object3d
	ElementType ElType
	Indexes     []int // index of coordinate
}

// valid matrix element constants
var valid = [...][2]int{{2, 2}, {3, 3}, {4, 4}}

func (e Element) Check() error {
	for i := range valid {
		if int(e.ElementType) == valid[i][0] && len(e.Indexes) != valid[i][1] {
			return fmt.Errorf("unacceptable element: %v", e)
		}
	}
	return fmt.Errorf("undefined element: %v", e)
}

// Coordinate store coordinate of points
type Coordinate struct {
	object3d
	X, Y, Z float64
}

// Named intermediant named structure
type Named struct{ Name string }
type Ignored struct{ IgnoreElements []bool }

var mm MultiModel

type MultiModel struct {
	// actual = 0, then change Model
	// 0 < actual, then change Parts[actual - 1]
	actual int

	Named
	Ignored
	Elements []Element
	Coords   []Coordinate

	Parts []Part
}

type Part struct {
	Named
	Ignored
}

func init() { // TODO remove
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 256
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)

	for i := 0; i < levels; i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0
		mm.Coords = append(mm.Coords,
			Coordinate{X: Ri * math.Sin(angle), Y: float64(i) * dy, Z: Ri * math.Cos(angle)},
			Coordinate{X: Ro * math.Sin(angle), Y: float64(i) * dy, Z: Ro * math.Cos(angle)},
		)
		mm.Elements = append(mm.Elements, Element{ElementType: Line2,
			Indexes: []int{2 * i, 2*i + 1},
		})
		if 0 < i {
			mm.Elements = append(mm.Elements,
				Element{ElementType: Line2,
					Indexes: []int{2 * (i - 1), 2 * i},
				}, Element{ElementType: Line2,
					Indexes: []int{2*(i-1) + 1, 2*i + 1},
				})
			mm.Elements = append(mm.Elements,
				Element{ElementType: Triangle3,
					Indexes: []int{2 * (i - 1), 2 * i, 2*(i-1) + 1},
				}, Element{ElementType: Triangle3,
					Indexes: []int{2 * i, 2*(i-1) + 1, 2*i + 1},
				})
		}
	}
	updateModel = true // TODO  remove
}

const distanceError = 1e-6

func clearValue(str string) string {
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")
	return str
}

func (mm *MultiModel) AddNode(X, Y, Z string) (err error) {
	// clear
	X = clearValue(X)
	Y = clearValue(Y)
	Z = clearValue(Z)
	// parse
	x, err := strconv.ParseFloat(X, 64)
	if err != nil {
		return
	}
	y, err := strconv.ParseFloat(Y, 64)
	if err != nil {
		return
	}
	z, err := strconv.ParseFloat(Z, 64)
	if err != nil {
		return
	}
	// check is this coordinate exist?
	for i := range mm.Coords {
		distance := math.Sqrt(pow.E2(mm.Coords[i].X-x) +
			pow.E2(mm.Coords[i].Y-y) +
			pow.E2(mm.Coords[i].Z-z))
		if distance < distanceError {
			return
		}
	}
	// append
	mm.Coords = append(mm.Coords, Coordinate{X: x, Y: y, Z: z})
	updateModel = true // Update camera parameter
	return
}

func (mm *MultiModel) AddLineByNodeNumber(N1, N2 string) (err error) {
	// clear
	N1 = clearValue(N1)
	N2 = clearValue(N2)
	// parse
	// n1
	n1, err := strconv.ParseUint(N1, 10, 64)
	if err != nil {
		return
	}
	// n2
	n2, err := strconv.ParseUint(N2, 10, 64)
	if err != nil {
		return
	}
	// type convection
	ni1 := int(n1)
	ni2 := int(n2)
	// check is this coordinate exist?
	for _, el := range mm.Elements {
		if el.ElementType != Line2 {
			continue
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni2 {
			return
		}
		if el.Indexes[1] == ni1 && el.Indexes[0] == ni2 {
			return
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Line2,
		Indexes:     []int{ni1, ni2},
	})
	updateModel = true // Update camera parameter
	return
}

func (mm *MultiModel) AddTriangle3ByNodeNumber(N1, N2, N3 string) (err error) {
	// clear
	N1 = clearValue(N1)
	N2 = clearValue(N2)
	N3 = clearValue(N3)
	// TODO:
	updateModel = true // Update camera parameter
	return
}

func (mm *MultiModel) SelectNodes(single bool) (ids []uint) {
	for i := range mm.Coords {
		if !mm.Coords[i].selected {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *MultiModel) SelectLines(single bool) (ids []uint)     { return }
func (mm *MultiModel) SelectTriangles(single bool) (ids []uint) { return }

func (mm *MultiModel) SplitLinesByDistance(line, distance string, atBegin bool) {}
func (mm *MultiModel) SplitLinesByRatio(line, proportional string, pos uint)    {}
func (mm *MultiModel) SplitLinesByEqualParts(lines, parts string)               {}
func (mm *MultiModel) SplitTri3To3Tri3(tris string)                             {}

func (mm *MultiModel) MoveCopyNodesDistance(nodes string, coordinates [3]string, copy, addLines, addTri bool) {
}
func (mm *MultiModel) MoveCopyNodesN1N2(nodes, from, to string, copy, addLines, addTri bool) {}

//
// Approach is not aurogenerate model, but approach is
// fast create model.
//
// Union of 2 models
//	   m1                                                     //
//	+------------+                                            //
//	|            |                                            //
//	|            |                                            //
//	+------------+                                            //
//	                                                          //
//	               m2                m3                       //
//	        +-----------+         +--------+                  //
//	        |           |         |        |                  //
//	        |           |         |        |                  //
//	        +-----------+         +--------+                  //
//	                                                          //
//	   m1     +    m2                                         //
//	+-------+----+------+                                     //
//	|xxxxxxx|xxxx|xxxxxx|                                     //
//	|xxxxxxx|xxxx|xxxxxx|                                     //
//	+-------+----+------+                                     //
//	                                                          //
//	   m1                +           m3                       //
//	+------------+                +--------+                  //
//	|xxxxxxxxxxxx|                |xxxxxxxx|                  //
//	|xxxxxxxxxxxx|                |xxxxxxxx|                  //
//	+------------+                +--------+                  //
//
// Both model must based on single mesh
// type Union struct {
// 	Models []int
// TODO:
// problem with union points.
// How to connect 2 models?
// }
//
// FilterModel
//	   m1                                                     //
//	+------------+                                            //
//	|            |                                            //
//	|            |                                            //
//	+------------+                                            //
//	                                                          //
//	          m2                                              //
//	        +----+                                            //
//	        |    |                                            //
//	        |    |                                            //
//	        +----+                                            //
//	                                                          //
//	   m1   - m2                                              //
//	+-------+----+                                            //
//	|xxxxxxx|    |                                            //
//	|xxxxxxx|    |                                            //
//	+-------+----+                                            //
//
// Both models may based on different mesh
//
// Examples of use:
//	* Erection cases
//
// type FilterModel struct {
// 	Base          int
// 	IgnoreElement []bool
// }
//
// Modificator from `base` model to `after` model after run Updater.
// Amount coordinates, elements are `after` model equal `base` model.
//
// Examples of use:
//	* Coords based on buckling imperfections.
//	* Coords based on deformation imperfections.
//	* Move/Rotate base model.
//
// type Modificator struct {
// Update from `base` model to `after` model.
// Specific code for update models.
// Examples of code:
//	// move and rotate models №0
//	Clone 0
//	Move 0,0,2000
//	Rotate 0,90,0
//	// add imperfection buckling shape №1 at 0.750 and №2 at 0.250
//	Clone 0
//	Imperfection Buckl 1 0.750 2 0.250
//	// union models 0 and 1
//	Union 0 1 TODO: ????? MERGE POINTS
// 	Updater string
// Model after run function Update().
// If `after` is not valid, then use `base` model value.
// 	After int
// }

//	Erection case:
//	"Model based on Model with hided some parts"
//
//	Buckling imperfection:
//	"Model based on Model with modifications in according to buckling shape"
//
//	Submodel:
//	"Model based on Model with hided some parts and modifications on both Models"
//
//	Building:
//	"Model create by copy of one Model with modification of elevation"
//
//	Combine:
//	"Combine 2 models with new supports/loads"
//
//	 MODEL                  MODEL                                //
//	 |                      |                                    //
//	 |          --- X ----  |                                    //
//	 |          --------->  |                                    //
//	 +-- MESH   <-------->  +-- MESH                             //
//	 |          +-------->  |                                    //
//	 |          |  +----->  |                                    //
//	 |          |  |        |                                    //
//	 +-- LOADS  |  | -----  +-- LOADS                            //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- COMB   |  | -----  +-- COMB                             //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- SUP    |  | -----  +-- SUPPORTS                         //
//	 |          |  | --X--  |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 |          |  |        |                                    //
//	 +-- BUCKLING  |        +--                                  //
//	 |             |                                             //
//	 |             |                                             //
//	 |             |                                             //
//	 +--------- FREQUENCY                                        //
//	                                                             //

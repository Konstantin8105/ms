package ms

import (
	"fmt"
	"math"

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

// TODO : type MultiModel struct { Models []Model}
var mm Model // TODO : remove

type Model struct {
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

func (mm *Model) AddNode(X, Y, Z float64) (err error) {
	// check is this coordinate exist?
	for i := range mm.Coords {
		distance := math.Sqrt(pow.E2(mm.Coords[i].X-X) +
			pow.E2(mm.Coords[i].Y-Y) +
			pow.E2(mm.Coords[i].Z-Z))
		if distance < distanceError {
			return
		}
	}
	// append
	mm.Coords = append(mm.Coords, Coordinate{X: X, Y: Y, Z: Z})
	updateModel = true // Update camera parameter
	return
}

func (mm *Model) AddLineByNodeNumber(n1, n2 uint) (err error) {
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

func (mm *Model) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (err error) {
	// type convection
	ni1 := int(n1)
	ni2 := int(n2)
	ni3 := int(n3)
	// check is this coordinate exist?
	for _, el := range mm.Elements {
		if el.ElementType != Triangle3 {
			continue
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni2 && el.Indexes[2] == ni3 {
			return
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni3 && el.Indexes[2] == ni1 {
			return
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni1 && el.Indexes[2] == ni2 {
			return
		}
		if el.Indexes[0] == ni3 && el.Indexes[1] == ni2 && el.Indexes[2] == ni1 {
			return
		}
		if el.Indexes[0] == ni2 && el.Indexes[1] == ni1 && el.Indexes[2] == ni3 {
			return
		}
		if el.Indexes[0] == ni1 && el.Indexes[1] == ni3 && el.Indexes[2] == ni2 {
			return
		}
	}
	// append
	mm.Elements = append(mm.Elements, Element{
		ElementType: Triangle3,
		Indexes:     []int{ni1, ni2, ni3},
	})
	updateModel = true // Update camera parameter
	return
}

func (mm *Model) IgnoreElements(ids []uint) {
	ignore := &mm.IgnoreElements
	if 0 < mm.actual {
		ignore = &mm.Parts[mm.actual-1].IgnoreElements
	}
	*ignore = append(*ignore, ids...)
}

func (mm *Model) Unignore() {
	ignore := &mm.IgnoreElements
	if 0 < mm.actual {
		ignore = &mm.Parts[mm.actual-1].IgnoreElements
	}
	*ignore = nil
}

func (mm *Model) SelectNodes(single bool) (ids []uint) {
	for i := range mm.Coords {
		if !mm.Coords[i].selected {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectLines(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		if !el.selected {
			continue
		}
		if el.ElementType != Line2 {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectTriangles(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		if !el.selected {
			continue
		}
		if el.ElementType != Triangle3 {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SelectElements(single bool) (ids []uint) {
	for i, el := range mm.Elements {
		if !el.selected {
			continue
		}
		ids = append(ids, uint(i))
	}
	return
}

func (mm *Model) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	// split point on line corner
	// split point inside line
	// split point outside line
	// TODO
}

func (mm *Model) SplitLinesByRatio(lines []uint, proportional float64, pos uint) {
	// split point on line corner
	// split point inside line
	// split point outside line
	// TODO
}

func (mm *Model) SplitLinesByEqualParts(lines []uint, parts uint) {
	// TODO
}

func (mm *Model) SplitTri3To3Tri3(tris []uint) {
	// TODO
}

func (mm *Model) MoveCopyNodesDistance(nodes, elements []uint, coordinates [3]float64, copy, addLines, addTri bool) {
	// TODO
}
func (mm *Model) MoveCopyNodesN1N2(nodes, elements []uint, from, to uint, copy, addLines, addTri bool) {
	// TODO
}

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

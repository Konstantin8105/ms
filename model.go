package ms

import (
	"fmt"
	"math"
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
			return fmt.Errorf("Unacceptable element: %v", e)
		}
	}
	return fmt.Errorf("Undefined element: %v", e)
}

// Coordinate store coordinate of points
type Coordinate struct {
	object3d
	X, Y, Z float64
}

// Named intermediant named structure
type Named struct{ Name string }
type Ignored struct{ Elements []bool }

type MultiModel struct {
	actual int // 0 ... len(Models) ... len(Models)+len(Parts)
	Models []Model
	Parts  []Part
}

type Model struct {
	Named
	Ignored
	Coordinates []Coordinate
	Elements    []Element
}

type Part struct {
	Named
	Ignored
	Base int
}

var model Model // TODO change to MultiModel

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
		model.Coordinates = append(model.Coordinates,
			Coordinate{X: Ri * math.Sin(angle), Y: float64(i) * dy, Z: Ri * math.Cos(angle)},
			Coordinate{X: Ro * math.Sin(angle), Y: float64(i) * dy, Z: Ro * math.Cos(angle)},
		)
		model.Elements = append(model.Elements, Element{
			ElementType: Line2,
			Indexes:     []int{2 * i, 2*i + 1},
		})
		if 0 < i {
			model.Elements = append(model.Elements,
				Element{ElementType: Line2,
					Indexes: []int{2 * (i - 1), 2 * i},
				}, Element{ElementType: Line2,
					Indexes: []int{2*(i-1) + 1, 2*i + 1},
				})
			model.Elements = append(model.Elements,
				Element{ElementType: Triangle3,
					Indexes: []int{2 * (i - 1), 2 * i, 2*(i-1) + 1},
				}, Element{ElementType: Triangle3,
					Indexes: []int{2 * i, 2*(i-1) + 1, 2*i + 1},
				})
		}
	}
	updateModel = true // TODO  remove
}

func (m *Model) AddNode(X, Y, Z float64) {
	m.Coordinates = append(m.Coordinates, Coordinate{X: X, Y: Y, Z: Z})
	updateModel = true // Update camera parameter
}

// func (m *Model) AddNodeByDistance(line, distance string, atBegin bool) {
// }
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
//	* Coordinates based on buckling imperfections.
//	* Coordinates based on deformation imperfections.
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

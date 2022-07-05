package ms

import "fmt"

var counter uint

func getId() uint {
	counter++
	return counter
}

type DebugMesh struct{}

func (DebugMesh) AddNode(X, Y, Z string) {
	Debug = append(Debug, fmt.Sprintln("InsertNode: ", X, Y, Z))
}

func (DebugMesh) SelectLines(single bool) (ids []uint) {
	ids = []uint{314, 567, getId()}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectLines: ", ids))
	return
}

func (DebugMesh) SelectNodes(single bool) (ids []uint) {
	ids = []uint{getId(), 1, 23, 444}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectNodes: ", ids))
	return
}

func (DebugMesh) SelectTriangles(single bool) (ids []uint) {
	ids = []uint{getId(), 333, 555, 777, 888, 999, 111, 222, 123, 345}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectTriangles: ", ids))
	return
}

func (DebugMesh) SelectQuadr4(single bool) (ids []uint) {
	ids = []uint{getId(), 1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectQuadr4: ", ids))
	return
}

func (DebugMesh) SplitLinesByDistance (line, distance string, atBegin bool) {
	Debug = append(Debug,
		fmt.Sprintln("SplitLinesByDistance: ", line, distance, atBegin))
}

func (DebugMesh) SplitLinesByRatio (line, proportional string, pos uint) {
	Debug = append(Debug,
		fmt.Sprintln("SplitLinesByRatio: ", line, proportional, pos))
}

func (DebugMesh) AddLineByNodeNumber(n1, n2 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertLineByNodeNumber: ", n1, n2))
}

func (DebugMesh) AddTriangle3ByNodeNumber(n1, n2, n3 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertTriangle3ByNodeNumber: ", n1, n2, n3))
}

func (DebugMesh) AddQuadr4ByNodeNumber(n1, n2, n3, n4 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertQuadr4ByNodeNumber: ", n1, n2, n3, n4))
}

func (DebugMesh) AddElementsByNodes(ids string, l2, t3, q4 bool) {
	Debug = append(Debug,
		fmt.Sprintln("InsertElementsByNodes: ", ids, l2, t3, q4))
}

func (DebugMesh) SplitTri3To3Quadr4(tris string) {
	Debug = append(Debug,
		fmt.Sprintln("SplitTri3To3Quadr4: ", tris))
}

func (DebugMesh) SplitTri3To2Tri3(tris string, side uint) {
	Debug = append(Debug,
		fmt.Sprintln("SplitTri3To2Tri3: ", tris, side))
}

func (DebugMesh) SplitQuadr4To2Quadr4(q4s string, side uint) {
	Debug = append(Debug,
		fmt.Sprintln("SplitQuadr4To2Quadr4: ", q4s, side))
}

func (DebugMesh) SplitLinesByEqualParts(lines, parts string) {
	Debug = append(Debug,
		fmt.Sprintln("SplitLinesByEqualParts: ", lines, parts))
}

func (DebugMesh) MoveCopyNodesDistance(nodes string, coordinates [3]string, copy bool) {
	Debug = append(Debug,
		fmt.Sprintln("MoveCopyNodesDistance: ", nodes, coordinates, copy))
}

func (DebugMesh) MoveCopyNodesN1N2(nodes, from, to string, copy bool) {
	Debug = append(Debug,
		fmt.Sprintln("MoveCopyNodesN1N2: ", nodes, from, to, copy))
}

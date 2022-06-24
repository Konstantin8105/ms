package ms

import "fmt"

type DebugMesh struct{}

func (DebugMesh) InsertNode(X, Y, Z string) {
	Debug = append(Debug, fmt.Sprintln("InsertNode: ", X, Y, Z))
}

func (DebugMesh) SelectLines(single bool) (ids []uint) {
	ids = []uint{314, 567}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectLines: ", ids))
	return
}

func (DebugMesh) SelectNodes(single bool) (ids []uint) {
	ids = []uint{1, 23, 444}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectNodes: ", ids))
	return
}

func (DebugMesh) SelectTriangles(single bool) (ids []uint) {
	ids = []uint{333, 555, 777}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectTriangles: ", ids))
	return
}

func (DebugMesh) SelectQuadr4(single bool) (ids []uint) {
	ids = []uint{1111, 2222, 3333, 4444}
	if single {
		ids = ids[:1]
	}
	Debug = append(Debug,
		fmt.Sprintln("SelectQuadr4: ", ids))
	return
}

func (DebugMesh) InsertNodeByDistance(line, distance string, pos uint) {
	Debug = append(Debug,
		fmt.Sprintln("InsertNodeByDistance: ", line, distance, pos))
}

func (DebugMesh) InsertNodeByProportional(line, proportional string, pos uint) {
	Debug = append(Debug,
		fmt.Sprintln("InsertNodeByProportional: ", line, proportional, pos))
}

func (DebugMesh) InsertLineByNodeNumber(n1, n2 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertLineByNodeNumber: ", n1, n2))
}

func (DebugMesh) InsertTriangle3ByNodeNumber(n1, n2, n3 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertTriangle3ByNodeNumber: ", n1, n2, n3))
}

func (DebugMesh) InsertQuadr4ByNodeNumber(n1, n2, n3, n4 string) {
	Debug = append(Debug,
		fmt.Sprintln("InsertQuadr4ByNodeNumber: ", n1, n2, n3, n4))
}

func (DebugMesh) InsertElementsByNodes(ids string, l2, t3, q4 bool) {
	Debug = append(Debug,
		fmt.Sprintln("InsertElementsByNodes: ", ids, l2, t3, q4))
}

func (DebugMesh) SplitLinesByRatio(lines, ratio string) {
	Debug = append(Debug,
		fmt.Sprintln("SplitLinesByRatio: ", lines, ratio))
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

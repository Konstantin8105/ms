package ms

import (
	"fmt"
)

type Undo struct {
	Model
	// 	sync.Mutex
	// 	list *list.List
}

func (u *Undo) addToUndo() {
	// 	b, err := json.Marshal(u)
	// 	if err != nil {
	// 		Debug = append(Debug, fmt.Sprintf("addToUndo: %v", err))
	// 		return
	// 	}
	// 	if u.list == nil {
	// 		u.list = list.New()
	// 	}
	// 	u.list.PushBack(b)
	// Debug = append(Debug, fmt.Sprintf("List size: %d", u.list.Len()))
}

func (u *Undo) Undo() {
	// 	u.Lock()
	// 	defer u.Unlock()

	Debug = append(Debug, fmt.Sprintf("Undo action"))

	// 	u.DeselectAll()
	// 	el := u.list.Back()
	// 	if el == nil {
	// 		return
	// 	}
	// 	var last Model
	// 	b := el.Value.([]byte)
	// 	if err := json.Unmarshal(b, &last); err != nil {
	// 		Debug = append(Debug, fmt.Sprintf("Undo: %v", err))
	// 		return
	// 	}
	//	u.Model = &last
	// 	u.init()
	// 	u.updateModel = true
	// 	u.list.Remove(el)
	// 	Debug = append(Debug, fmt.Sprintf("List size: %d", u.list.Len()))
}

func (u *Undo) AddNode(X, Y, Z float64) (id uint) {
	defer u.addToUndo() // store
	return u.AddNode(X, Y, Z)
}

func (u *Undo) AddLineByNodeNumber(n1, n2 uint) (id uint) {
	defer u.addToUndo() // store
	return u.AddLineByNodeNumber(n1, n2)
}

func (u *Undo) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint) {
	defer u.addToUndo() // store
	return u.AddTriangle3ByNodeNumber(n1, n2, n3)
}

func (u *Undo) IgnoreModelElements(ids []uint) {
	defer u.addToUndo() // store
	u.IgnoreModelElements(ids)
}

func (u *Undo) Unignore() {
	defer u.addToUndo() // store
	u.Unignore()
}

func (u *Undo) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	defer u.addToUndo() // store
	u.SplitLinesByDistance(lines, distance, atBegin)
}

func (u *Undo) SplitLinesByRatio(lines []uint, proportional float64, atBegin bool) {
	defer u.addToUndo() // store
	u.SplitLinesByRatio(lines, proportional, atBegin)
}

func (u *Undo) SplitLinesByEqualParts(lines []uint, parts uint) {
	defer u.addToUndo() // store
	u.SplitLinesByEqualParts(lines, parts)
}

func (u *Undo) SplitTri3To3Tri3(tris []uint) {
	defer u.addToUndo() // store
	u.SplitTri3To3Tri3(tris)
}

func (u *Undo) MergeNodes(minDistance float64) {
	defer u.addToUndo() // store
	u.MergeNodes(minDistance)
}

func (u *Undo) MoveCopyNodesDistance(nodes, elements []uint, coordinates [3]float64, copy, addLines, addTri bool) {
	defer u.addToUndo() // store
	u.MoveCopyNodesDistance(nodes, elements, coordinates, copy, addLines, addTri)
}

func (u *Undo) MoveCopyNodesN1N2(nodes, elements []uint, from, to uint, copy, addLines, addTri bool) {
	defer u.addToUndo() // store
	u.MoveCopyNodesN1N2(nodes, elements, from, to, copy, addLines, addTri)
}

func (u *Undo) DemoSpiral() {
	defer u.addToUndo() // store
	u.DemoSpiral()
}

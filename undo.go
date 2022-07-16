package ms

import (
	"container/list"
	"encoding/json"
	"fmt"
)

type Undo struct {
	list  *list.List
	model *Model
}

func (u *Undo) addToUndo() {
	b, err := json.Marshal(u.model)
	if err != nil {
		Debug = append(Debug, fmt.Sprintf("addToUndo: %v", err))
		return
	}
	if u.list == nil {
		u.list = list.New()
	}
	u.list.PushBack(b)
}

func (u *Undo) Undo() {
	u.model.DeselectAll()
	el := u.list.Back()
	if el == nil {
		return
	}
	var last Model
	b := el.Value.([]byte)
	if err := json.Unmarshal(b, &last); err != nil {
		Debug = append(Debug, fmt.Sprintf("Undo: %v", err))
		return
	}
	// swap models

	// opengl
	last.op = u.model.op
	last.op.ChangeModel(&last)
	// tui
	last.tui = u.model.tui
	last.tui.ChangeModel(&last)
	// undo model
	u.model = &last

	// remove
	u.list.Remove(el)
}

func (u *Undo) PartPresent() (id uint) {
	return u.model.PartPresent()
}

func (u *Undo) PartsName() (names []string) {
	return u.model.PartsName()
}

func (u *Undo) PartChange(id uint) {
	u.model.PartChange(id)
}

func (u *Undo) PartNew(str string) {
	u.model.PartNew(str)
}

func (u *Undo) PartRename(id uint, str string) {
	u.model.PartRename(id, str)
}

func (u *Undo) StandardView(view SView) {
	u.model.StandardView(view)
}

func (u *Undo) ColorEdge(isColor bool) {
	u.model.ColorEdge(isColor)
}

func (u *Undo) AddNode(X, Y, Z float64) (id uint) {
	defer u.addToUndo() // store
	return u.model.AddNode(X, Y, Z)
}

func (u *Undo) AddLineByNodeNumber(n1, n2 uint) (id uint) {
	defer u.addToUndo() // store
	return u.model.AddLineByNodeNumber(n1, n2)
}

func (u *Undo) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint) {
	defer u.addToUndo() // store
	return u.model.AddTriangle3ByNodeNumber(n1, n2, n3)
}

func (u *Undo) IgnoreModelElements(ids []uint) {
	defer u.addToUndo() // store
	u.model.IgnoreModelElements(ids)
}

func (u *Undo) Unignore() {
	defer u.addToUndo() // store
	u.model.Unignore()
}

func (u *Undo) Hide(coordinates, elements []uint) {
	u.model.Hide(coordinates, elements)
}

func (u *Undo) UnhideAll() {
	u.model.UnhideAll()
}

func (u *Undo) SelectLeftCursor(nodes, lines, tria bool) {
	u.model.SelectLeftCursor(nodes, lines, tria)
}

func (u *Undo) SelectNodes(single bool) (ids []uint) {
	return u.model.SelectNodes(single)
}

func (u *Undo) SelectLines(single bool) (ids []uint) {
	return u.model.SelectLines(single)
}

func (u *Undo) SelectTriangles(single bool) (ids []uint) {
	return u.model.SelectTriangles(single)
}

func (u *Undo) SelectElements(single bool) (ids []uint) {
	return u.model.SelectElements(single)
}

func (u *Undo) InvertSelect(nodes, lines, triangles bool) {
	u.model.InvertSelect(nodes, lines, triangles)
}

func (u *Undo) SelectLinesOrtho(x, y, z bool) {
	u.model.SelectLinesOrtho(x, y, z)
}

func (u *Undo) SelectLinesOnPlane(xoy, xoz, yoz bool) {
	u.model.SelectLinesOnPlane(xoy, xoz, yoz)
}

func (u *Undo) DeselectAll() {
	u.model.DeselectAll()
}

func (u *Undo) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	defer u.addToUndo() // store
	u.model.SplitLinesByDistance(lines, distance, atBegin)
}

func (u *Undo) SplitLinesByRatio(lines []uint, proportional float64, atBegin bool) {
	defer u.addToUndo() // store
	u.model.SplitLinesByRatio(lines, proportional, atBegin)
}

func (u *Undo) SplitLinesByEqualParts(lines []uint, parts uint) {
	defer u.addToUndo() // store
	u.model.SplitLinesByEqualParts(lines, parts)
}

func (u *Undo) SplitTri3To3Tri3(tris []uint) {
	defer u.addToUndo() // store
	u.model.SplitTri3To3Tri3(tris)
}

func (u *Undo) MergeNodes(minDistance float64) {
	defer u.addToUndo() // store
	u.model.MergeNodes(minDistance)
}

func (u *Undo) MoveCopyNodesDistance(nodes, elements []uint, coordinates [3]float64, copy, addLines, addTri bool) {
	defer u.addToUndo() // store
	u.model.MoveCopyNodesDistance(nodes, elements, coordinates, copy, addLines, addTri)
}

func (u *Undo) MoveCopyNodesN1N2(nodes, elements []uint, from, to uint, copy, addLines, addTri bool) {
	defer u.addToUndo() // store
	u.model.MoveCopyNodesN1N2(nodes, elements, from, to, copy, addLines, addTri)
}

func (u *Undo) DemoSpiral() {
	defer u.addToUndo() // store
	u.model.DemoSpiral()
}

func (u *Undo) Remove(nodes, elements []uint) {
	defer u.addToUndo() // store
	u.model.Remove(nodes, elements)
}

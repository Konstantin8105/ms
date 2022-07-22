package ms

import (
	"container/list"
	"encoding/json"
	"sync"
)

type Undo struct {
	list *list.List

	mu sync.Mutex

	model *Model  // actual model
	op    *Opengl // for 3d
	tui   *Tui    // for terminal ui
}

func (u *Undo) sync(isUndo bool) (pre, post func()) {
	// no need opengl lock, because used panic-free model
	return func() {
			// Lock/Unlock model for avoid concurrency problems
			// with Opengl drawing
			u.mu.Lock() // mutex lock evethink
			if !isUndo {
				u.addToUndo() // store model in undo list
			}
		}, func() {
			u.op.UpdateModel() // update camera view
			u.mu.Unlock()      // mutex unlock everythink
		}
}

func (u *Undo) addToUndo() {
	b, err := json.Marshal(u.model)
	if err != nil {
		AddInfo("addToUndo: %v", err)
		return
	}
	if u.list == nil {
		u.list = list.New()
		u.addToUndo() // store
	}
	u.list.PushBack(b)
}

func (u *Undo) Undo() {
	// sync
	pre, post := u.sync(true)
	pre()
	defer post()
	// action
	el := u.list.Back()
	if el == nil {
		return
	}
	var last Model
	b := el.Value.([]byte)
	if err := json.Unmarshal(b, &last); err != nil {
		AddInfo("Undo: %v", err)
		return
	}
	// swap models

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
	u.op.StandardView(view)
}

func (u *Undo) ColorEdge(isColor bool) {
	u.op.ColorEdge(isColor)
}

func (u *Undo) ViewAll(centerCorrection bool) {
	u.op.ViewAll(centerCorrection)
}

func (u *Undo) AddNode(X, Y, Z float64) (id uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddNode(X, Y, Z)
}

func (u *Undo) AddLineByNodeNumber(n1, n2 uint) (id uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddLineByNodeNumber(n1, n2)
}

func (u *Undo) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddTriangle3ByNodeNumber(n1, n2, n3)
}

func (u *Undo) GetCoords() []Coordinate {
	return u.model.GetCoords()
}

func (u *Undo) GetElements() []Element {
	return u.model.GetElements()
}

func (u *Undo) IgnoreModelElements(ids []uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.IgnoreModelElements(ids)
}

func (u *Undo) Unignore() {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Unignore()
}

func (u *Undo) IsIgnore(elID uint) bool {
	return u.model.IsIgnore(elID)
}

func (u *Undo) Hide(coordinates, elements []uint) {
	u.model.Hide(coordinates, elements)
}

func (u *Undo) UnhideAll() {
	u.model.UnhideAll()
}

func (u *Undo) AddLeftCursor(lc LeftCursor) {
	u.op.AddLeftCursor(lc)
}

func (u *Undo) SelectLeftCursor(nodes, lines, tria bool) {
	u.op.SelectLeftCursor(nodes, lines, tria)
}

func (u *Undo) AddModel(m Model) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.AddModel(m)
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

func (u *Undo) SelectLinesParallel(lines []uint) {
	u.model.SelectLinesParallel(lines)
}

func (u *Undo) SelectLinesByLenght(more bool, lenght float64) {
	u.model.SelectLinesByLenght(more, lenght)
}

func (u *Undo) SelectLinesCylindrical(node uint, radiant, conc bool, axe Direction) {
	u.model.SelectLinesCylindrical(node, radiant, conc, axe)
}

func (u *Undo) SelectScreen(from, to [2]int32) {
	u.op.SelectScreen(from, to)
}

func (u *Undo) DeselectAll() {
	u.model.DeselectAll()
}

func (u *Undo) SelectAll(nodes, lines, triangles bool) {
	u.model.SelectAll(nodes, lines, triangles)
}

func (u *Undo) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByDistance(lines, distance, atBegin)
}

func (u *Undo) SplitLinesByRatio(lines []uint, proportional float64, atBegin bool) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByRatio(lines, proportional, atBegin)
}

func (u *Undo) SplitLinesByEqualParts(lines []uint, parts uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByEqualParts(lines, parts)
}

func (u *Undo) SplitTri3To3Tri3(tris []uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitTri3To3Tri3(tris)
}

func (u *Undo) MergeNodes(minDistance float64) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.MergeNodes(minDistance)
}

func (u *Undo) MoveCopyDistance(nodes, elements []uint, coordinate [3]float64,
	intermediantParts uint,
	copy, addLines, addTri bool) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.MoveCopyDistance(nodes, elements, coordinate,
		intermediantParts,
		copy, addLines, addTri)
}

func (u *Undo) MoveCopyN1N2(nodes, elements []uint, from, to uint,
	intermediantParts uint,
	copy, addLines, addTri bool) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.MoveCopyN1N2(nodes, elements, from, to,
		intermediantParts,
		copy, addLines, addTri)
}

func (u *Undo) DemoSpiral() {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.DemoSpiral()
}

func (u *Undo) Remove(nodes, elements []uint) {
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Remove(nodes, elements)
}

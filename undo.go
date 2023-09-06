package ms

import (
	"container/list"
	"encoding/json"

	"github.com/Konstantin8105/gog"
)

type Undo struct {
	list *list.List

	//mu sync.Mutex

	model *Model  // actual model
	op    *Opengl // for 3d
	// tui   *Tui    // for terminal ui
	changed        bool
	quit           *chan struct{}
	initialization func()
}

func (u *Undo) addTuiInitialization(f func()) {
	u.initialization = f
}

func (u *Undo) sync(isUndo bool) (pre, post func()) {
	// no need opengl lock, because used panic-free model
	return func() {
			// Lock/Unlock model for avoid concurrency problems
			// with Opengl drawing
			// u.mu.Lock() // mutex lock evethink
			if !isUndo {
				u.addToUndo() // store model in undo list
			}
		}, func() {
			u.changed = true
			// u.op.UpdateModel() // update camera view
			// u.mu.Unlock()      // mutex unlock everythink
		}
}

func (u *Undo) addToUndo() {
	b, err := json.Marshal(u.model)
	if err != nil {
		logger.Printf("addTo%v", err)
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
		logger.Printf("%v", err)
		return
	}
	// swap models

	// undo model
	u.model = &last

	// remove
	u.list.Remove(el)
}

func (u *Undo) Open(name string) (err error) {
	logger.Print("Open")
	u.list = list.New()
	err = u.model.Open(name)
	if err != nil {
		return
	}
	u.StandardView(StandardViewXOYpos)
	u.changed = false
	if f := u.initialization; f != nil {
		f()
	}
	return
}

func (u *Undo) Save() error {
	logger.Print("Save")
	if err := u.model.Save(); err != nil {
		logger.Printf("Save: %v", err)
		return err
	}
	u.changed = false
	return nil
}

func (u *Undo) SaveAs(filename string) error {
	logger.Print("SaveAs")
	if err := u.model.SaveAs(filename); err != nil {
		logger.Printf("Save: %v", err)
		return err
	}
	u.changed = false
	return nil
}

func (u *Undo) Close() {
	logger.Print("Close")
	*u.op.actions <- func() (fus bool) {
		close(*u.quit)
		return false
	}
}

func (u *Undo) IsChangedModel() bool {
	logger.Print("IsChangedModel")
	return u.changed
}

func (u *Undo) GetPresentFilename() (name string) {
	logger.Print("GetPresentFilename")
	return u.model.GetPresentFilename()
}

// func (u *Undo) PartPresent() (id uint) {
// 	// too many : logger.Print("PartPresent")
// 	return u.model.PartPresent()
// }
//
// func (u *Undo) PartsName() (names []string) {
// 	// too many : logger.Print("PartsName")
// 	return u.model.PartsName()
// }
//
// func (u *Undo) PartChange(id uint) {
// 	// too many : logger.Print("PartChange")
// 	u.model.PartChange(id)
// }
//
// func (u *Undo) PartNew(str string) {
// 	logger.Print("PartNew")
// 	u.model.PartNew(str)
// }
//
// func (u *Undo) PartRename(id uint, str string) {
// 	// too many: logger.Print("PartRename")
// 	u.model.PartRename(id, str)
// }

func (u *Undo) StandardView(view SView) {
	logger.Print("StandardView")
	u.op.StandardView(view)
}

func (u *Undo) ColorEdge(isColor bool) {
	logger.Print("ColorEdge")
	u.op.ColorEdge(isColor)
}

func (u *Undo) ViewAll(centerCorrection bool) {
	logger.Print("ViewAll")
	u.op.ViewAll(centerCorrection)
}

func (u *Undo) AddNode(X, Y, Z float64) (id uint) {
	logger.Print("AddNode")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddNode(X, Y, Z)
}

func (u *Undo) AddLineByNodeNumber(n1, n2 uint) (id uint) {
	logger.Print("AddLineByNodeNumber")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddLineByNodeNumber(n1, n2)
}

func (u *Undo) AddTriangle3ByNodeNumber(n1, n2, n3 uint) (id uint, ok bool) {
	logger.Print("AddTriangle3ByNodeNumber")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.AddTriangle3ByNodeNumber(n1, n2, n3)
}

func (u *Undo) GetCoordByID(id uint) (_ gog.Point3d, ok bool) {
	logger.Print("GetCoordByID")
	return u.model.GetCoordByID(id)
}

func (u *Undo) GetCoords() []Coordinate {
	// too many : logger.Print("GetCoords")
	return u.model.GetCoords()
}

func (u *Undo) GetElements() []Element {
	// too many : logger.Print("GetElements")
	return u.model.GetElements()
}

// func (u *Undo) IgnoreModelElements(ids []uint) {
// 	logger.Print("IgnoreModelElements")
// 	// sync
// 	pre, post := u.sync(false)
// 	pre()
// 	defer post()
// 	// action
// 	u.model.IgnoreModelElements(ids)
// }
//
// func (u *Undo) Unignore() {
// 	logger.Print("Unignore")
// 	// sync
// 	pre, post := u.sync(false)
// 	pre()
// 	defer post()
// 	// action
// 	u.model.Unignore()
// }
//
// func (u *Undo) IsIgnore(elID uint) bool {
// 	// too many logs: logger.Print("IsIgnore")
// 	return u.model.IsIgnore(elID)
// }

func (u *Undo) Hide(coordinates, elements []uint) {
	logger.Print("Hide")
	u.model.Hide(coordinates, elements)
}

func (u *Undo) UnhideAll() {
	logger.Print("UnhideAll")
	u.model.UnhideAll()
}

func (u *Undo) AddLeftCursor(lc LeftCursor) {
	logger.Print("AddLeftCursor")
	u.op.AddLeftCursor(lc)
}

func (u *Undo) SelectLeftCursor(nodes bool, elements []bool) {
	logger.Print("SelectLeftCursor")
	u.op.SelectLeftCursor(nodes, elements)
}

func (u *Undo) AddModel(m Model) {
	logger.Print("AddModel")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.AddModel(m)
}

func (u *Undo) GetSelectNodes(single bool) (ids []uint) {
	logger.Print("GetSelectNodes")
	return u.model.GetSelectNodes(single)
}

func (u *Undo) GetSelectElements(single bool, filter func(_ ElType) (acceptable bool)) (ids []uint) {
	logger.Print("GetSelectElements")
	return u.model.GetSelectElements(single, filter)
}

func (u *Undo) InvertSelect(nodes bool, elements []bool) {
	logger.Print("InvertSelect")
	u.model.InvertSelect(nodes, elements)
}

func (u *Undo) SelectLinesOrtho(x, y, z bool) {
	logger.Print("SelectLinesOrtho")
	u.model.SelectLinesOrtho(x, y, z)
}

func (u *Undo) SelectLinesOnPlane(xoy, xoz, yoz bool) {
	logger.Print("SelectLinesOnPlane")
	u.model.SelectLinesOnPlane(xoy, xoz, yoz)
}

func (u *Undo) SelectLinesParallel(lines []uint) {
	logger.Print("SelectLinesParallel")
	u.model.SelectLinesParallel(lines)
}

func (u *Undo) SelectLinesByLenght(more bool, lenght float64) {
	logger.Print("SelectLinesByLenght")
	u.model.SelectLinesByLenght(more, lenght)
}

func (u *Undo) SelectLinesCylindrical(node uint, radiant, conc bool, axe Direction) {
	logger.Print("SelectLinesCylindrical")
	u.model.SelectLinesCylindrical(node, radiant, conc, axe)
}

func (u *Undo) SelectScreen(from, to [2]int32) {
	logger.Print("SelectScreen")
	u.op.SelectScreen(from, to)
}

func (u *Undo) DeselectAll() {
	logger.Print("DeselectAll")
	u.model.DeselectAll()
}

func (u *Undo) SelectAll(nodes bool, elements []bool) {
	logger.Print("SelectAll")
	u.model.SelectAll(nodes, elements)
}

func (u *Undo) SplitLinesByDistance(lines []uint, distance float64, atBegin bool) {
	logger.Print("SplitLinesByDistance")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByDistance(lines, distance, atBegin)
}

func (u *Undo) SplitLinesByRatio(lines []uint, proportional float64, atBegin bool) {
	logger.Print("SplitLinesByRatio")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByRatio(lines, proportional, atBegin)
}

func (u *Undo) SplitLinesByEqualParts(lines []uint, parts uint) {
	logger.Print("SplitLinesByEqualParts")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitLinesByEqualParts(lines, parts)
}

func (u *Undo) SplitTri3To3Tri3(tris []uint) {
	logger.Print("SplitTri3To3Tri3")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.SplitTri3To3Tri3(tris)
}

func (u *Undo) MergeNodes(minDistance float64) {
	logger.Print("MergeNodes")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.MergeNodes(minDistance)
}

func (u *Undo) MergeLines(lines []uint) {
	logger.Print("MergeLines")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.MergeLines(lines)
}

func (u *Undo) Intersection(nodes, elements []uint) {
	logger.Print("Intersection")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Intersection(nodes, elements)
}

func (u *Undo) Move(nodes, elements []uint,
	basePoint [3]float64,
	path diffCoordinate) {
	logger.Print("Move")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Move(nodes, elements, basePoint, path)
}

func (u *Undo) Copy(nodes, elements []uint,
	basePoint [3]float64,
	paths []diffCoordinate,
	addLines, addTri bool) {
	logger.Print("Copy")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Copy(nodes, elements, basePoint, paths, addLines, addTri)
}

func (u *Undo) Mirror(nodes, elements []uint,
	basePoint [3][3]float64,
	copy bool,
	addLines, addTri bool) {
	logger.Print("Mirror")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Mirror(nodes, elements, basePoint, copy, addLines, addTri)
}

func (u *Undo) DemoSpiral(n uint) {
	logger.Print("DemoSpiral")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.DemoSpiral(n)
}

func (u *Undo) Remove(nodes, elements []uint) {
	logger.Print("Remove")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.Remove(nodes, elements)
}

func (u *Undo) RemoveSameCoordinates() {
	logger.Print("RemoveSameCoordinates")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.RemoveSameCoordinates()
}

func (u *Undo) RemoveNodesWithoutElements() {
	logger.Print("RemoveNodesWithoutElements")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.RemoveNodesWithoutElements()
}

func (u *Undo) RemoveZeroLines() {
	logger.Print("RemoveZeroLines")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.RemoveZeroLines()
}

func (u *Undo) RemoveZeroTriangles() {
	logger.Print("RemoveZeroTriangles")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	u.model.RemoveZeroTriangles()
}

func (u *Undo) Check() error {
	logger.Print("Check")
	// sync
	pre, post := u.sync(false)
	pre()
	defer post()
	// action
	return u.model.Check()
}

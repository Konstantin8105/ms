package ms

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/msh"
)

func Gmsh2Model(filename string) (model Model, err error) {
	if filename == "" {
		err = fmt.Errorf("empty gmsh GEO filename")
		return
	}
	var geo []byte
	geo, err = os.ReadFile(filename)
	if err != nil {
		return
	}
	var mesh *msh.Msh
	mesh, err = msh.New(string(geo))
	if err != nil {
		return
	}
	for _, n := range mesh.Nodes {
		var node Coordinate
		node.Point3d[0] = n.Coord[0]
		node.Point3d[1] = n.Coord[1]
		node.Point3d[2] = n.Coord[2]
		model.Coords = append(model.Coords, node)
	}
	for _, e := range mesh.Elements {
		switch e.EType {
		case msh.Line:
			var el Element
			el.ElementType = Line2
			el.Indexes = e.NodeId
			for i := range el.Indexes {
				el.Indexes[i] -= 1
			}
			model.Elements = append(model.Elements, el)
		case msh.Triangle:
			var el Element
			el.ElementType = Triangle3
			el.Indexes = e.NodeId
			for i := range el.Indexes {
				el.Indexes[i] -= 1
			}
			model.Elements = append(model.Elements, el)
		}
	}
	return
}

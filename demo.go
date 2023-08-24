//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
)

// gmsh -2 -smooth 10 -format msh22 1.geo

func main() {
	// root, action, err := ms.UserInterface()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v", err)
	// 	os.Exit(1)
	// }
	// err = vl.Run(root, action, nil, tcell.KeyCtrlC)
	// if err != nil {
	// 	fmt.Println(ms.PrintInfo())
	// }
	// create a new model
	// if err := ms.Run("testdata/1.geo", nil); err != nil {
	if err := ms.Run("testdata/IntersectionSpiral", nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

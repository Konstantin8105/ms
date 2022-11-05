//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
)

// gmsh -2 -smooth 10 -format msh22 1.geo

func main() {
	defer func() {
		fmt.Println(ms.PrintInfo())
	}()
	// create a new model
	if err := ms.Run("testdata/1.geo", nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

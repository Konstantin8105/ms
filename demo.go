//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
)

func main() {
	// TODO: Prototype
	// filename := ""
	// if err := ms.Run(filename); err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v", err)
	// 	return
	// }
	// create user interface
	root, action, err := ms.UserInterface()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	go func() {
		if err := ms.M3(); err != nil {
			fmt.Fprintf(os.Stderr, "M3: %v", err)
		}
	}()
	err = vl.Run(root, action, nil, tcell.KeyCtrlC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	for i := range ms.Debug {
		fmt.Println(ms.Debug[i])
	}
}

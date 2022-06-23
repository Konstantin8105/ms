//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/vl"
	"github.com/Konstantin8105/ms"
	"github.com/gdamore/tcell/v2"
)

func main() {
	root, action, err := ms.UserInterface()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	err = vl.Run(root, action, tcell.KeyCtrlC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	for i := range ms.Debug {
		fmt.Println(ms.Debug[i])
	}
}


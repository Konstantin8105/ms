//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/ms"
	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"
)

func main() {
	log.Printf("Only for debugging tui")
	var (
		mm        ms.Undo
		closedApp = false
		ch        = make(chan ds.Action, 1000)
		chQuit    = make(chan func(), 1)
	)
	tui, initialization, err := ms.NewTui(&mm, &closedApp, &ch)
	if err != nil {
		return
	}
	_ = initialization
	err = vl.Run(tui, chQuit, nil, tcell.KeyCtrlC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

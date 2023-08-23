package main

import (
	"github.com/Konstantin8105/ds"
	"github.com/Konstantin8105/glsymbol"
	"github.com/Konstantin8105/ms/window"
	"github.com/Konstantin8105/vl"
)

func main() {
	var ws [2]ds.Window

	w0, a0 := vl.Demo()
	tw0 := window.NewTui(w0)
	ws[0] = tw0

	w1, a1 := vl.Demo()
	tw1 := window.NewTui(w1)
	ws[1] = tw1

	ch := make(chan func(), 1000)
	go func() {
		for a := range a0 {
			ch <- a
		}
	}()
	go func() {
		for a := range a1 {
			ch <- a
		}
	}()

	screen, err := ds.New("Demo", ws, &ch)
	if err != nil {
		panic(err)
	}

	// add fonts
	f, err := glsymbol.DefaultFont()
	if err != nil {
		panic(err)
	}
	tw0.SetFont(f)
	tw1.SetFont(f)

	screen.Run()
}

//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
)

func main() {
	defer func() {
		fmt.Println(ms.PrintInfo())
	}()
	// create a new model
	if err := ms.Run(nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

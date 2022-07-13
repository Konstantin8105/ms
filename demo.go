//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
)

func main() {
	defer func() {
		for i := range ms.Debug {
			fmt.Println(ms.Debug[i])
		}
	}()
	// create a new model
	if err := ms.Run(nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

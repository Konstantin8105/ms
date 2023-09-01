//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/Konstantin8105/ms"
)

func main() {
	if err := ms.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

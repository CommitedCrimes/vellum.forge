package main

import (
	"fmt"
	"sort"

	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	styleNames := styles.Names()
	sort.Strings(styleNames)
	fmt.Println("Available Chroma styles:")
	for _, name := range styleNames {
		fmt.Printf("  %s\n", name)
	}
}

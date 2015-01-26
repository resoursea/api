package api

import (
	"fmt"
	"strings"
)

func PrintRouter(rt Router) {
	fmt.Println("\n--- PRINT ROUTER ---\n")
	printRouter(rt, 0)
	fmt.Println("\n--- END PRINT ---\n")
}

func printRouter(r Router, lvl int) {
	fmt.Printf("%sRoute: %s\n", strings.Repeat("	", lvl), r)

	for _, m := range r.Methods() {
		printHandler(m, lvl)
	}

	for _, c := range r.Children() {
		if r.IsSlice() {
			printRouter(c, lvl)
		} else {
			printRouter(c, lvl+1)
		}
	}
}

func printHandler(m Method, lvl int) {
	fmt.Printf("%s- Method: %s\n", strings.Repeat("	", lvl), m)
}

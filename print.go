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
	fmt.Printf("%s%s\n", strings.Repeat("	", lvl), r.String())

	for _, h := range r.Handlers() {
		printHandler(h, lvl)
	}

	for _, c := range r.Children() {
		printRouter(c, lvl+1)
	}
}

func printHandler(h Handler, lvl int) {
	fmt.Printf("%s- %s\n", strings.Repeat("	", lvl), h.String())
}

package api

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func PrintResource(r *Resource) {
	fmt.Println("\n--- PRINT RESOURCE ---\n")
	printResource(r, 0)
	fmt.Println("\n--- END PRINT ---\n")
}

func printResource(r *Resource, lvl int) {
	fmt.Printf("%-20s %-18s ", strings.Repeat("|  ", lvl)+"|-["+r.Name+"]", r.Value.Type())

	if len(r.Tag) > 0 {
		fmt.Printf("tag: '%s' ", r.Tag)
	}

	if len(r.Extends) > 0 {
		fmt.Printf("extends: ")
		for _, e := range r.Extends {
			fmt.Printf("%s ", e.Value.Type())
		}
	}

	fmt.Printf("%#v", r.Value.Interface())

	fmt.Println()

	if r.IsSlice {
		printResource(r.Elem, lvl)
	}

	for _, c := range r.Children {
		printResource(c, lvl+1)
	}
}

func PrintRoute(ro *Route) {
	fmt.Println("\n--- PRINT ROUTE ---\n")
	printRoute(ro, 0)
	fmt.Println("\n--- END PRINT ---\n")
}

func printRoute(ro *Route, lvl int) {
	fmt.Printf("%s", strings.Repeat("|	", lvl)+"|-/"+ro.URI)

	if ro.IsSlice {
		fmt.Printf(" *is slice*")
	}

	fmt.Println()

	for _, h := range ro.Handlers {
		printHandler(h, true, lvl+1)
	}

	for _, c := range ro.Children {
		printRoute(c, lvl+1)
	}
}

func printHandler(h *handler, isSlice bool, lvl int) {
	fmt.Printf("%s ", strings.Repeat("|	", lvl)+"| -")

	if isSlice {
		fmt.Printf("[] ")
	}

	fmt.Println(h.Method.Name + ": " + h.Method.HTTPMethod + "()   Dependencies:")

	for t, d := range h.Dependencies {
		printDependency(t, d, lvl+1)
	}

}

func printDependency(t reflect.Type, d *dependency, lvl int) {
	fmt.Printf("%s %-24s as %-24s", strings.Repeat("|	", lvl)+"-", t, d.Value.Type())

	if d.Method != nil {
		fmt.Printf(" Init Input:")
	} else {
		fmt.Printf(" Desn't have Init method")
	}

	for _, input := range d.Method.Inputs {
		fmt.Printf(" %-24s", input)
	}

	fmt.Println()

}

// Print the Resource the stack,
// marking some resource within the stack
// It is used to alert user for circular dependency
// in the resource relationships
func printResourceStack(r, mark *Resource) {

	if r.Parent != nil {
		printResourceStack(r.Parent, mark)
	}

	if r.isSameType(mark) {
		log.Printf("*** -> %s\n", r.String())
	} else {
		log.Printf("    -> %s\n", r.String())
	}

}

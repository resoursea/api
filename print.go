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
	printRoute(ro, 0, false)
	fmt.Println("\n--- END PRINT ---\n")
}

// Pass Elem when printing a Elem inside an Slice
func printRoute(ro *Route, lvl int, elem bool) {

	// This add the last bar for the addres of Elements of Slices
	bar := ""
	if elem {
		bar = "/"
	}

	fmt.Printf("%s /%s%s   %s\n", strings.Repeat("|	", lvl)+"| -", ro.Name, bar, ro.Value.Type())

	for _, h := range ro.Handlers {
		printHandler(h, lvl+1)
	}

	if ro.IsSlice {
		printRoute(ro.Elem, lvl, true)
	}

	for _, c := range ro.Children {
		printRoute(c, lvl+1, false)
	}
}

func printHandler(h *handler, lvl int) {
	fmt.Printf("%s ", strings.Repeat("|	", lvl)+"| -")

	fmt.Printf("[%s] /%s     Dependencies:\n", h.Method.HTTPMethod, h.Method.Name)

	for t, d := range h.Dependencies {
		printDependency(t, d, lvl+1)
	}

}

func printDependency(t reflect.Type, d *dependency, lvl int) {
	fmt.Printf("%s %-24s as %-24s", strings.Repeat("|	", lvl)+"- - -", t, d.Value.Type())

	if d.Method != nil {
		fmt.Printf(" Init Input:")
		for _, input := range d.Method.Inputs {
			fmt.Printf(" %-24s", input)
		}
	} else {
		fmt.Printf(" Desn't have Init method")
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

	if r.Value.Type() == mark.Value.Type() {
		log.Printf("*** -> %s\n", r.String())
	} else {
		log.Printf("    -> %s\n", r.String())
	}

}

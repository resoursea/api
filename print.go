package api

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func printResource(r *resource, lvl int) {
	fmt.Printf("%-16s %-20s %-5v  ",
		strings.Repeat("|  ", lvl)+"|-["+r.Name+"]",
		r.Value.Type(), r.isSlice())

	if len(r.Tag) > 0 {
		fmt.Printf("tag: '%s' ", r.Tag)
	}

	if r.isSlice() {
		fmt.Printf("slice: %s ", r.SliceValue.Type())
	}

	for _, e := range r.Extends {
		fmt.Printf("extends: %s ", e.Value.Type())
	}

	fmt.Println()

	for _, c := range r.Children {
		printResource(c, lvl+1)
	}
}

func printRoute(ro *route, lvl int) {
	fmt.Printf("%s", strings.Repeat("|	", lvl)+"|-/"+ro.URI)

	if ro.IsSlice {
		fmt.Printf(" *is slice*")
	}

	fmt.Println()

	for _, h := range ro.SliceHandlers {
		printHandler(h, true, lvl+1)
	}

	for _, h := range ro.Handlers {
		printHandler(h, false, lvl+1)
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

	fmt.Println(h.Name + "()   Dependencies:")

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
func printResourceStack(r, mark *resource) {

	if r.Parent != nil {
		printResourceStack(r.Parent, mark)
	}

	if r.isEqual(mark) {
		log.Printf("*** -> %s\n", r.String())
	} else {
		log.Printf("    -> %s\n", r.String())
	}

}
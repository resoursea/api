package resource

import (
	"fmt"
	"reflect"
	"strings"
)

func printResource(r *Resource, lvl int) {
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

func printRoute(r *Route, lvl int) {
	fmt.Printf("%-20s  ", strings.Repeat("|	", lvl)+"|-["+r.URI+"]")

	fmt.Println()

	for _, m := range r.Methods {
		printMethod(m, lvl+1)
	}

	for _, c := range r.Children {
		printRoute(c, lvl+1)
	}
}

func printMethod(m *Method, lvl int) {
	fmt.Printf("%-20s  ", strings.Repeat("|	", lvl)+"|-"+m.Name+"()   Dependencies:")

	fmt.Println()

	for t, d := range m.Dependencies {
		printDependency(t, d, lvl+1)
	}

}

func printDependency(t reflect.Type, d *Dependency, lvl int) {
	fmt.Printf("%s %-24s as %-24s", strings.Repeat("|	", lvl)+"-", t, d.Value.Type())

	if len(d.Input) > 0 {
		fmt.Printf(" Init Input:")
	} else {
		fmt.Printf(" Desn't have Init method")
	}

	for _, input := range d.Input {
		fmt.Printf(" %-24s", input)
	}

	fmt.Println()

}

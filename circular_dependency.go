package api

import (
	"errors"
	"fmt"
	"reflect"
)

type circularDependency struct {
	Checked    []*dependency
	Dependents []reflect.Type
}

// Check the existence of Circular Dependency on the route
func checkCircularDependency(ro *Route) error {
	cd := &circularDependency{
		Checked:    []*dependency{},
		Dependents: []reflect.Type{},
	}
	return cd.checkRoute(ro)
}

func (cd *circularDependency) checkRoute(ro *Route) error {
	for _, h := range ro.Handlers {
		//log.Println("Check CD for Method", h.Method)
		for _, d := range h.Dependencies {
			err := cd.checkDependency(d, h)
			if err != nil {
				return err
			}
		}
	}

	for _, child := range ro.Children {
		err := cd.checkRoute(child)
		if err != nil {
			return err
		}
	}

	return nil
}

// This method add de Dependency to the Dependents list testing if it conflicts
// and moves recursively on each Dependency of this Dependency...
// at the end of the method the Dependency is removed from the Dependents list
func (cd *circularDependency) checkDependency(d *dependency, h *handler) error {
	// If this Dependency is already checked,
	// we don't need to check it again
	if cd.isChecked(d) {
		return nil
	}

	//log.Println("CD for Dependency", d.Value.Type())

	// Add this dependency type to the dependency list
	// and check if this type desn't already exist
	err := cd.addAndCheck(d.Value.Type())
	if err != nil {
		return err
	}

	// Check if this Dependency has Init Method
	if d.Method != nil {
		for _, t := range d.Method.Inputs {
			//log.Println("CD for Dependency Init Dependency", i, t, dependency.isType(t))

			// The first element will always be the dependency itself
			if d.isType(t) {
				continue
			}

			// All context types doesn't need to be checked
			// it will always be present in the context
			if isContextType(t) {
				continue
			}

			d, exist := h.Dependencies.vaueOf(t)
			if !exist { // It should never occurs!
				return fmt.Errorf("Danger! No dependency %s found! Something very wrong happened!", t)
			}

			// Go ahead recursively on each Dependency
			err := cd.checkDependency(d, h)
			if err != nil {
				return err
			}
		}
	}

	// Remove itself from the list
	cd.pop()

	// Add this dependency to the checked list
	cd.Checked = append(cd.Checked, d)

	return nil
}

// Check if this dependency Type doesn't exist in the Dependents list
// If it already exist, it indicates a circular dependency!
// Throws an error showing the list of dependencies that caused it
func (cd *circularDependency) addAndCheck(t reflect.Type) error {

	// Check for circular dependency
	ok := true
	errMsg := ""

	for _, t2 := range cd.Dependents {
		if !ok {
			errMsg += fmt.Sprintf("%s that depends on ", t2)
		}
		if t == t2 {
			ok = false
			errMsg += fmt.Sprintf("%s depends on ", t)
		}
	}

	if !ok {
		errMsg += fmt.Sprintf("%s\n", t)
		return errors.New(errMsg)
	}

	//log.Println("Adding:", t)

	// Everything ok, add this new type dependency
	cd.Dependents = append(cd.Dependents, t)
	return nil
}

// Remove the last element from the Dependents list
func (cd *circularDependency) pop() {
	//log.Println("Removing:", cd.Dependents[len(cd.Dependents)-1])
	cd.Dependents = cd.Dependents[:len(cd.Dependents)-1]
}

func (cd *circularDependency) isChecked(dependency *dependency) bool {
	for _, d := range cd.Checked {
		if dependency == d {
			return true
		}
	}
	return false
}

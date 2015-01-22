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
	c := &circularDependency{
		Dependents: []reflect.Type{},
	}
	err := c.checkRoute(ro)
	if err != nil {
		return err
	}
	return nil
}

func (c *circularDependency) checkRoute(ro *Route) error {

	for _, h := range ro.Handlers {

		//log.Println("Check CD for Method", h.Method)

		for _, d := range h.Dependencies {
			// It's necessary cause we will have many depenedencies
			// indexed by differente types
			if c.notChecked(d) {

				err := c.checkDependency(d, h)
				if err != nil {
					return err
				}

				// Add this dependency to the checked list
				c.Checked = append(c.Checked, d)
			}
		}
	}

	for _, child := range ro.Children {
		err := c.checkRoute(child)
		if err != nil {
			return err
		}
	}

	return nil
}

// This method add de Dependency to the Dependents list testing if it conflicts
// and moves recursively on each Dependency of this Dependency...
// at the end of the method the Dependency is removed from the Dependents list
func (c *circularDependency) checkDependency(dependency *dependency, h *handler) error {

	//log.Println("CD for Dependency", dependency.Value.Type())

	// Add this dependency type to the dependency list
	// and check if this type desn't already exist
	err := c.addAndCheck(dependency.Value.Type())
	if err != nil {
		return err
	}

	// Check if this Dependency has Init Method
	if dependency.Method != nil {
		for _, t := range dependency.Method.Inputs {

			//log.Println("CD for Dependency Init Dependency", i, t, dependency.isType(t))

			// The first element will always be the dependency itself
			if dependency.isType(t) {
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
			err := c.checkDependency(d, h)
			if err != nil {
				return err
			}
		}
	}

	// Remove itself from the list
	c.pop()

	return nil
}

// Checks whether the Dependents doesn't fall into a circular dependency
// Add a new type to the Dependents list
func (c *circularDependency) addAndCheck(t reflect.Type) error {

	// Check for circular dependency
	ok := true
	errMsg := ""

	for _, t2 := range c.Dependents {
		if !ok {
			errMsg += fmt.Sprintf("%s that depends on ", t2)
		}
		if t == t2 {
			ok = false
			errMsg += fmt.Sprintf("%s depends on ", t2)
		}
	}

	if !ok {
		errMsg += fmt.Sprintf("%s\n", t)
		return errors.New(errMsg)
	}

	//log.Println("Adding:", t)

	// Everything ok, add this new type dependency
	c.Dependents = append(c.Dependents, t)
	return nil
}

// Remove the last element from the Dependents list
func (c *circularDependency) pop() {
	//log.Println("Removing:", c.Dependents[len(c.Dependents)-1])
	c.Dependents = c.Dependents[:len(c.Dependents)-1]
}

func (c *circularDependency) notChecked(dependency *dependency) bool {
	for _, d := range c.Checked {
		if dependency == d {
			return false
		}
	}
	return true
}

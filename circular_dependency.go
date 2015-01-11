package api

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

type circularDependency struct {
	Checked    []*dependency
	Dependents []reflect.Type
}

func checkCircularDependency(ro *route) error {
	c := &circularDependency{
		Dependents: []reflect.Type{},
	}
	err := c.checkRoute(ro)
	if err != nil {
		return err
	}
	return nil
}

func (c *circularDependency) checkRoute(ro *route) error {

	for _, h := range ro.Handlers {

		//log.Println("Check CD for", h.Method.Method)

		for _, d := range h.Dependencies {
			// It's necessary cause we will have many depenedencies
			// indexed by differente types
			if c.notChecked(d) {

				//log.Println("###Checking : ", d.Value.Type())

				err := c.checkDependency(d, h.Dependencies)
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

func (c *circularDependency) checkDependency(dependency *dependency, dependencies dependencies) error {

	// Add this dependency type to the dependency list
	err := c.add(dependency.Value.Type())
	if err != nil {
		log.Fatalln(err)
		return err
	}

	for i, t := range dependency.Method.Inputs {

		// The first element will always be the dependency itself
		if i == 0 {
			continue
		}

		// IDs types desn't need to be declared,
		// cause it will be present in the context
		if t == idType {
			continue
		}

		d, exist := dependencies.vaueOf(t)
		if !exist { // It should never occurs!
			log.Panicf("Danger! No dependency %s found!\n", t)
		}
		c.checkDependency(d, dependencies)
	}

	// Remove itself from the list
	c.pop()

	return nil
}

// Checks whether the Dependents doesn't fall into a circular dependency
// Add a new type to the Dependents list
func (c *circularDependency) add(t reflect.Type) error {

	// Check for circular dependency
	ok := true
	text := ""

	for _, t2 := range c.Dependents {
		if !ok {
			text += fmt.Sprintf("%s that depends on ", t2)
		}
		if t == t2 {
			ok = false
			text += fmt.Sprintf("%s depends on ", t2)
		}
	}

	if !ok {
		text += fmt.Sprintf("%s\n", t)
		return errors.New(text)
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

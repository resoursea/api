package resource

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

type CircularDependency struct {
	Checked    []*Dependency
	Dependents []reflect.Type
}

func circularDependency(r *Route) error {
	c := &CircularDependency{
		Dependents: []reflect.Type{},
	}
	err := c.checkRoute(r)
	if err != nil {
		return err
	}
	return nil
}

func (c *CircularDependency) checkRoute(r *Route) error {

	for _, h := range r.Handlers {

		log.Println("Check CD for", h.Method.Method)

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

	for _, child := range r.Children {
		err := c.checkRoute(child)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CircularDependency) checkDependency(dependency *Dependency, dependencies Dependencies) error {

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
		if t == IDType {
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
func (c *CircularDependency) add(t reflect.Type) error {

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
func (c *CircularDependency) pop() {
	//log.Println("Removing:", c.Dependents[len(c.Dependents)-1])
	c.Dependents = c.Dependents[:len(c.Dependents)-1]
}

func (c *CircularDependency) notChecked(dependency *Dependency) bool {
	for _, d := range c.Checked {
		if dependency == d {
			return false
		}
	}
	return true
}

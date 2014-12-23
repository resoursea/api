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

func checkCircularDependency(r *Route) {
	c := &CircularDependency{
		Dependents: []reflect.Type{},
	}
	err := c.checkRoute(r)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *CircularDependency) checkRoute(r *Route) error {

	for _, m := range r.Methods {

		log.Println("Check CD for", m.Method)

		for _, d := range m.Dependencies {
			// It's necessary cause we will have many depenedencies
			// indexed by differente types
			if c.notChecked(d) {

				//log.Println("###Checking", index, " : ", d.Value.Type())

				err := c.checkDependency(d, m.Dependencies)
				if err != nil {
					return err
				}

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

func (c *CircularDependency) checkDependency(d *Dependency, dependencies Dependencies) error {

	// Add this dependency type to the dependency list
	err := c.add(d.Value.Type())
	if err != nil {
		log.Fatalln(err)
		return err
	}

	for i, t := range d.Input {

		// The first element will always be the dependency itself
		if i == 0 {
			continue
		}

		d2, exist := dependencies.vaueOf(t)
		if !exist { // It should never occurs!
			log.Panicf("Danger! No dependency %s found!\n", t)
		}
		c.checkDependency(d2, dependencies)
	}

	// Remove itself from the list
	c.pop()

	return nil
}

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

func (c *CircularDependency) notChecked(d *Dependency) bool {
	for _, checkedDependency := range c.Checked {
		if d == checkedDependency {
			return false
		}
	}
	return true
}

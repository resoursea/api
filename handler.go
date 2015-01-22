package api

import (
	"fmt"
	"reflect"
)

type handler struct {
	Method *method
	// Many Types could point to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies dependencies
}

func newHandler(m reflect.Method, r *Resource) (*handler, error) {

	//log.Println("Creating Handler for method", m.Name, m.Type)

	met := newMethod(m)

	h := &handler{
		Method:       met,
		Dependencies: make(map[reflect.Type]*dependency),
	}

	// So we scan all dependencies to create a tree
	for _, input := range met.Inputs {
		// Scan this dependency and its dependencies recursively
		err := h.newDependency(input, r)
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

// Scan the dependencies recursively and add it to the Handler dependencies list
// This method ensures that all dependencies will be present
// when the dependents methods want them
func (h *handler) newDependency(t reflect.Type, r *Resource) error {

	//log.Println("Trying to create a new dependency", t)

	// If the required resource is http.ResponseWriter or *http.Request or ID
	// it will be added to context on each request and don't need to be mapped
	if isContextType(t) {
		return nil // Not need to be mapped as a dependency
	}

	err := isValidDependencyType(t)
	if err != nil {
		return err
	}

	// Check if this type already exists in the dependencies
	// If it was indexed by another type, this method
	// ensures that it will be indexed for this type too
	_, exist := h.Dependencies.vaueOf(t)
	if exist {
		//log.Printf("Found dependency %s to use as %s\n", dp.Value, t)
		return nil // This type already exist in the Dependencies list
	}

	// If this dependency is an Interface,
	// we should search which resource satisfies this Interface in the Resource Tree
	// If this is a Struct, just find for the initial value,
	// if the Struct doesn't exist, create one and return it
	v, err := r.valueOf(t)
	if err != nil {
		return err
	}

	d := &dependency{
		Value:  v,
		Method: nil,
	}

	// We should add this dependency before scan its Init Dependencies
	// cause the Init first argument will requires the Resource itself
	h.Dependencies[t] = d

	//log.Printf("Created dependency %s to use as %s\n", v, t)

	// If this Resource has an Init Method,
	// then we should create it too
	// If Init Method is defined wrong, it trows an error
	return h.newInitMethod(d, r)
}

// Scan the dependencies of the Init method of some type
func (h *handler) newInitMethod(d *dependency, r *Resource) error {

	m, exists := d.Value.Type().MethodByName("Init")
	if !exists {
		//log.Printf("Type %s doesn't have Init method\n", d.Value.Type())
		return nil
	}

	//log.Println("Scanning Init ", m.Type)

	// Init method should have no return,
	// or return just the resource itself for idiomatic reasons
	err := isValidInit(m)
	if err != nil {
		return err
	}

	// Creates the Init Method
	// and attach it into the Dependency
	d.Method = newMethod(m)

	//log.Println("Scan Dependencies for Init method", d.Method.Method.Type)

	for _, input := range d.Method.Inputs {

		//log.Printf("Init %s depends on %s\n", d.Method.Method.Type, input)

		h.newDependency(input, r)
	}

	return nil
}

func (h *handler) String() string {
	return fmt.Sprintf("Handler: [%s%s] %s", h.Method.HTTPMethod, h.Method.Name, h.Method.Method.Type)
}

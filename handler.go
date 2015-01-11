package api

import (
	"log"
	"reflect"
)

type handler struct {
	Name   string
	Method *method
	// Many Types could poiny to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies dependencies
}

func newHandler(r *resource, m *method) *handler {

	//log.Println("Creating Handler", method.Name, "for resource", resource.Name)

	h := &handler{
		Name:         m.Name,
		Method:       m,
		Dependencies: make(map[reflect.Type]*dependency),
	}

	// So we scan all dependencies to create a tree
	for _, input := range m.Inputs {
		// Scan this dependency and its dependencies recursively
		h.scanDependency(input, r)
	}

	return h
}

// Scan the dependencies recursively and add it to the method
// This method ensures that all dependencies will be present
// when the dependent want then
func (h *handler) scanDependency(dependencyType reflect.Type, r *resource) {

	log.Println("Scanning dependency", dependencyType)

	// If the required resource is http.ResponseWriter or *http.Request or ID
	// it will be added to context on each request and don't need to be mapped
	if isContextType(dependencyType) {
		return
	}

	if !isValidDependencyType(dependencyType) {
		log.Fatalf("Type %s is not allowed as dependency\n", dependencyType)
	}

	// Check if this type already exists in the dependencies
	// If it was indexed by another type, this method
	// ensures that it will be indexed for this type too
	dp, exist := h.Dependencies.vaueOf(dependencyType)
	if exist {
		log.Printf("Found dependency %s to use as %s\n", dp.Value, dependencyType)
		return // This type already exist
	}

	// If this dependency is an Interface,
	// we should search which resource satisfies this Interface
	// If this is a Struct, just find for the initial value,
	// if doesn't exist, create one and return it
	value, err := r.valueOf(dependencyType)
	if err != nil {
		log.Fatal(err)
	}

	d := &dependency{
		Value:  value,
		Method: nil,
	}

	// We should add this dependency before scan its Init Dependencies
	// cause its first argument will requires itself
	h.Dependencies[dependencyType] = d

	log.Printf("Created dependency %s to use as %s\n", value, dependencyType)

	d.Method = scanInit(value, h, r)

}

// Scan the dependencies of the Init method of some type
func scanInit(value reflect.Value, h *handler, r *resource) *method {

	method, exists := value.Type().MethodByName("Init")
	if !exists {
		//log.Printf("Type %s doesn't have Init method\n", value.Type())
		return nil
	}

	// Init method should have no return,
	// or return just the resource itself for idiomatic reasons
	if !isValidInit(method) {
		log.Panicf("Resource %s has an invalid Init method, it should have no return,"+
			" or just return the resource itself %s \n",
			value.Type(), method.Type.In(0))
	}

	m := newMethod(method)

	//log.Println("Scan Init method for", m.Method.Type)

	for _, input := range m.Inputs {

		//log.Printf("Init %s depends on %s\n", m.Method.Type, input)

		h.scanDependency(input, r)
	}

	return m
}

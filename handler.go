package resource

import (
	"log"
	"reflect"
)

type Handler struct {
	Name   string
	Method *Method
	// Many Types could poiny to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies Dependencies
}

func NewHandler(resource *Resource, method *Method) *Handler {

	//log.Println("Creating Handler", method.Name, "for resource", resource.Name)

	h := &Handler{
		Name:         method.Name,
		Method:       method,
		Dependencies: make(map[reflect.Type]*Dependency),
	}

	// So we scan all dependencies to create a tree
	for _, input := range method.Inputs {
		// Scan this dependency and its dependencies recursively
		h.scanDependency(input, resource)
	}

	return h
}

// Scan the dependencies recursively and add it to the method
// This method ensures that all dependencies will be present
// when the dependent want then
func (h *Handler) scanDependency(dependencyType reflect.Type, resource *Resource) {

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
	value, err := resource.valueOf(dependencyType)
	if err != nil {
		log.Fatal(err)
	}

	d := &Dependency{
		Value:  value,
		Method: nil,
	}

	// We should add this dependency before scan its Init Dependencies
	// cause its first argument will requires itself
	h.Dependencies[dependencyType] = d

	log.Printf("Created dependency %s to use as %s\n", value, dependencyType)

	d.Method = scanInit(value, h, resource)

}

// Scan the dependencies of the Init method of some type
func scanInit(value reflect.Value, handler *Handler, resource *Resource) *Method {

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

	m := NewMethod(method)

	//log.Println("Scan Init method for", m.Method.Type)

	for _, input := range m.Inputs {

		//log.Printf("Init %s depends on %s\n", m.Method.Type, input)

		handler.scanDependency(input, resource)
	}

	return m
}

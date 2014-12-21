package resource

import (
	"log"
	"reflect"
)

type Method struct {
	Name   string
	Method reflect.Method
	Input  []reflect.Type
	// Many Types could poiny to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies Dependencies
}

func NewMethod(method reflect.Method, resource *Resource) *Method {

	log.Println("Creating method", method)

	m := &Method{
		Name:         method.Name,
		Method:       method,
		Input:        make([]reflect.Type, method.Type.NumIn()),
		Dependencies: make(map[reflect.Type]*Dependency),
	}

	// So we scan all dependencies to create a tree
	for i := 0; i < method.Type.NumIn(); i++ {

		m.Input[i] = method.Type.In(i)

		// Scan this dependency and its dependencies recursively
		m.scanDependency(m.Input[i], resource)

	}

	// Sort dependencies
	return m
}

// Scan the dependencies recursively and add it to the method
// This method ensures that all dependencies will be present
// when the dependent want then
func (m *Method) scanDependency(dependencyType reflect.Type, resource *Resource) {

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
	d, exist := m.Dependencies.vaueOf(dependencyType)
	if exist {
		log.Printf("Found dependency %s to use as %s\n", d.Value, dependencyType)
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

	dependency := &Dependency{
		Value:  value,
		Method: reflect.Method{},
		Input:  []reflect.Type{},
	}

	m.Dependencies[dependencyType] = dependency

	log.Printf("Created dependency %s to use as %s\n", value, dependencyType)
}

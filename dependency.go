package resource

import (
	"log"
	"reflect"
)

type Dependency struct {
	Value reflect.Value // The initial dependency state
	// Init method and its input
	Method reflect.Method
	Input  []reflect.Type
}

type Dependencies map[reflect.Type]*Dependency

// Return true if this dependency has a init method
func (d *Dependency) hasInit() bool {
	return d.Method.Type != nil
}

// This method checks if exist an value for the received type
// If it already exist, but in indexed by another type
// it will index for the received type too
func (dependencies Dependencies) vaueOf(dependencyType reflect.Type) (*Dependency, bool) {

	log.Println("Searching for dependency", dependencyType)

	d, exist := dependencies[dependencyType]
	if exist {
		log.Println("Indexed")
		return d, true
	}

	log.Println("Not indexed")

	// If its a Struct, check the existance for Struct Pointer type
	if dependencyType.Kind() == reflect.Struct {
		// Get the type for the pointer to this type
		t := reflect.New(dependencyType).Type()
		d, exist := dependencies[t]
		if exist {
			// Intex this dependency for this Struct type
			dependencies[dependencyType] = d // Ensure that we are using pointer to the same instance
			return d, true
		}
	}

	// If its a Pointer, check the existance for non Struct Pointer type
	if dependencyType.Kind() == reflect.Ptr {
		// Get the type for the element it points to
		t := dependencyType.Elem()
		d, exist := dependencies[t]
		if exist {
			// Intex this dependency for this Struct Pointer type
			dependencies[dependencyType] = d // Ensure that we are using pointer to the same instance
			return d, true
		}
	}

	if dependencyType.Kind() == reflect.Interface {
		log.Println("### INTERFACE")
		// Check if one of the dependencies implements this Interface
		for _, d := range dependencies {

			log.Println("###", d.Value.Type(), dependencyType, d.Value.Type().Implements(dependencyType))

			// Not working, cause
			// this value type is a non pointer,
			// and some methods are implemented on pointer values

			if d.Value.Type().Implements(dependencyType) {
				// Intex this dependency for this Interface type
				dependencies[dependencyType] = d
				return d, true
			}
		}
	}

	// Not found
	return nil, false
}

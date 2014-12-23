package resource

import (
	"reflect"
)

type Dependency struct {
	// The initial dependency state
	// Type Ptr to Struct, or Ptr to Slice of Struct
	Value reflect.Value

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
// it will index for the new type too
func (dependencies Dependencies) vaueOf(t reflect.Type) (*Dependency, bool) {

	//log.Println("Dependency: Searching for dependency", t)

	d, exist := dependencies[t]
	if exist {
		//log.Println("Dependency: Found:", d.Value.Type())
		return d, true
	}

	// Check if one of the dependencies is of this type
	for _, d := range dependencies {
		if d.isType(t) {
			//log.Println("Dependency: Found out of index", d.Value.Type())

			// Index this dependency with this new type it implements
			dependencies[t] = d
			return d, true
		}
	}

	//log.Println("Dependency: Not Exist")

	// Not found
	return nil, false
}

// Return true if this Resrouce is from by this Type
func (d *Dependency) isType(t reflect.Type) bool {

	if t.Kind() == reflect.Interface {
		return d.Value.Type().Implements(t)
	}

	// The value stored in Dependency
	// is the type Ptr to Struct, or Ptr to Slice of Struct
	t = ptrOfType(t)

	return d.Value.Type() == t
}

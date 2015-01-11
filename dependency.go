package api

import (
	"reflect"
)

type dependency struct {
	// The initial dependency state
	// Type Ptr to Struct, or Ptr to Slice of Struct
	Value reflect.Value

	// Init method and its input
	Method *method
}

type dependencies map[reflect.Type]*dependency

// This method checks if exist an value for the received type
// If it already exist, but in indexed by another type
// it will index for the new type too
func (ds dependencies) vaueOf(t reflect.Type) (*dependency, bool) {

	//log.Println("Dependency: Searching for dependency", t)

	d, exist := ds[t]
	if exist {
		//log.Println("Dependency: Found:", d.Value.Type())
		return d, true
	}

	// Check if one of the dependencies is of this type
	for _, d := range ds {
		if d.isType(t) {
			//log.Println("Dependency: Found out of index", d.Value.Type())

			// Index this dependency with this new type it implements
			ds[t] = d
			return d, true
		}
	}

	//log.Println("Dependency: Not Exist")

	// Not found
	return nil, false
}

// Return true if this Resrouce is from by this Type
func (d *dependency) isType(t reflect.Type) bool {

	if t.Kind() == reflect.Interface {
		return d.Value.Type().Implements(t)
	}

	// The value stored in Dependency
	// is the type Ptr to Struct, or Ptr to Slice of Struct
	t = ptrOfType(t)

	return d.Value.Type() == t
}

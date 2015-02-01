package api

import (
	"reflect"
)

type dependency struct {
	// The initial dependency state
	// Type Ptr to Struct, or Ptr to Slice of Struct
	value reflect.Value

	// Constructor method New()
	constructor *reflect.Method
}

type dependencies map[reflect.Type]*dependency

// Create a new Dependencies map
func newDependencies(m reflect.Method, r *resource) (dependencies, error) {
	//log.Println("Creating dependencies for method", m.Type)
	ds := dependencies{} //make(map[reflect.Type]*dependency)
	err := ds.scanMethodInputs(m, r)
	if err != nil {
		return nil, err
	}
	//log.Printf("Dependencies created: %v\n", ds)
	return ds, nil
}

// Scan the dependencies of a Method
func (ds dependencies) scanMethodInputs(m reflect.Method, r *resource) error {
	//log.Println("Trying to scan method", m.Type)
	// So we scan all dependencies to create a tree

	for i := 0; i < m.Type.NumIn(); i++ {
		input := m.Type.In(i)

		//log.Println("Scanning for dependency", input, "on method", m.Type)

		// Check if this type already exists in the dependencies
		// If it was indexed by another type, this method
		// ensures that it will be indexed for this type too
		if ds.exists(input) {
			//log.Printf("Found dependency to use as %s\n", input)
			continue
		}

		// If the required resource is http.ResponseWriter or *http.Request or ID
		// it will be added to context on each request and don't need to be mapped
		if isContextType(input) {
			continue // Not need to be mapped as a dependency
		}

		// Scan this dependency and its dependencies recursively
		// and add it to Dependencies list
		d, err := newDependency(input, r)
		if err != nil {
			return err
		}

		//log.Printf("Dependency created [%s]%s", input, d.value)

		// We should add this dependency before scan its constructor Dependencies
		// cause the constructor's first argument will requires the Resource itself

		ds.add(input, d)

		// Check if Dependency constructor exists
		constructor, exists := d.value.Type().MethodByName("New")
		if !exists {
			//log.Printf("Type %s doesn't have New method\n", d.Value.Type())
			continue
		}

		//log.Println("Scanning New for ", constructor.Type)

		// 'New' method should have no return,
		// or return just the resource itself and/or error
		err = isValidConstructor(constructor)
		if err != nil {
			return err
		}

		//log.Println("Scan Dependencies for 'New' method", d.Method.Method.Type)

		err = ds.scanMethodInputs(constructor, r)
		if err != nil {
			return err
		}

		// And attach it into the Dependency Method
		d.constructor = &constructor
	}
	return nil
}

// Scan the dependencies recursively and add it to the Method dependencies list
// This method ensures that all dependencies will be present
// when the dependents methods want them
func newDependency(t reflect.Type, r *resource) (*dependency, error) {

	//log.Println("Trying to create a new dependency", t)

	err := isValidDependencyType(t)
	if err != nil {
		return nil, err
	}

	// If this dependency is an Interface,
	// we should search which resource satisfies this Interface in the Resource Tree
	// If this is a Struct, just find for the initial value,
	// if the Struct doesn't exist, create one and return it
	v, err := r.valueOf(t)
	if err != nil {
		return nil, err
	}

	d := &dependency{
		value:       v,
		constructor: nil,
	}

	//log.Printf("Created dependency %s to use as %s\n", v, t)

	return d, nil
}

// Add a new dependency to the Dependencies list
// Receives an type, to use as index, and the dependency itself
// This type could be both Interface or Struct type
func (ds dependencies) add(t reflect.Type, d *dependency) {
	//log.Println("Adding dependency", d.Value.Type())
	ds[t] = d
}

// This method checks if exist an value for the received type
// If it already exist, but its indexed by another type
// it will index for the new type too
func (ds dependencies) vaueOf(t reflect.Type) (*dependency, bool) {

	//log.Println("Dependency: Searching for dependency", t)

	d, exist := ds[t]
	if exist {
		//log.Println("Dependency: Found:", d.value.Type(), d.value.Interface())
		return d, true
	}

	// Check if one of the dependencies is of this type
	for _, d := range ds {
		if d.isType(t) {
			//log.Println("Dependency: Found out of index", d.value.Interface())

			// Index this dependency with this new type it implements
			ds[t] = d
			return d, true
		}
	}

	//log.Println("Dependency: Not Exist")

	// Not found
	return nil, false
}

// Return true if required type already exists in the Dependencies map
// Check if this type already exists in the dependencies
// If it was indexed by another type, this method
// ensures that it will be indexed for this type too
func (ds dependencies) exists(t reflect.Type) bool {
	_, exist := ds.vaueOf(t)
	return exist
}

// Return true if this Resrouce is from by this Type
func (d *dependency) isType(t reflect.Type) bool {

	if t.Kind() == reflect.Interface {
		return d.value.Type().Implements(t)
	}

	// The Value stored in Dependency
	// is from Type Ptr to Struct, or Ptr to Slice of Struct
	return d.value.Type() == ptrOfType(t)
}

// Cosntruct a new dependency in a new memory space with the initial dependency value
func (d *dependency) new() reflect.Value {
	v := reflect.New(d.value.Type().Elem())
	v.Elem().Set(d.value.Elem())
	return v
}

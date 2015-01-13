package api

import (
	"log"
	"net/http"
	"reflect"
)

// This constants will be used on the method above
var (
	responseWriterPtrType = reflect.TypeOf((*http.ResponseWriter)(nil))
	tesponseWriterType    = responseWriterPtrType.Elem()
	requestPtrType        = reflect.TypeOf((*http.Request)(nil))
	requestType           = requestPtrType.Elem()
)

// This method return true if the received type is an context type
// It means that it doesn't need to be mapped and will be present in the context
// It also return an error message if user used *http.ResponseWriter or used http.Request
func isContextType(resourceType reflect.Type) bool {
	// Test if user used *http.ResponseWriter insted of http.ResponseWriter
	if resourceType.AssignableTo(responseWriterPtrType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, tesponseWriterType)
	}
	// Test if user used http.Request insted of *http.Request
	if resourceType.AssignableTo(requestType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, requestPtrType)
	}
	// Test if user used *ID insted of ID
	if resourceType.AssignableTo(idPtrType) {
		log.Fatalf("You asked for %s when you should used %s", idPtrType, idType)
	}

	return resourceType.AssignableTo(tesponseWriterType) ||
		resourceType.AssignableTo(requestPtrType) ||
		resourceType == idType
}

// Return the Pointer to Struct Value if passed one of those types
// Struct, *Struct, []Struct or []*Struct
// Return one Pointer to Struct value and one Pointer to Slice value, if exists
func getPtrValue(value reflect.Value) (reflect.Value, bool) {

	// If its an pointer, extract the element of that pointer
	value = elemOf(value)

	if value.Kind() == reflect.Slice {
		if value.IsNil() || value.Len() == 0 {
			// If it is nil or have no initial value
			// we will use an empty value of the slice type

			// Create a new Ptr to Slice value by the Slice Type passed
			return reflect.New(value.Type()), true
		} else {
			// If the slice has an element inside,
			// so set the value passed

			// Create a new Ptr to Slice value by the Slice Type passed
			ptrValue := reflect.New(value.Type())
			// Add the current Value to this new Ptr to Slice Value
			ptrValue.Elem().Set(value)

			return ptrValue, true
		}
	}

	if value.Kind() == reflect.Struct {
		// Create a new Ptr to Struct value by this Struct type
		ptrValue := reflect.New(value.Type())
		// Add the current value to this new Ptr to Struct value
		ptrValue.Elem().Set(value)

		return ptrValue, true
	}

	// Return false if the final value is not an Struct
	return reflect.Value{}, false
}

// Return the Value of the Elem of this Slice
func sliceElem(value reflect.Value) reflect.Value {

	value = elemOf(value)
	log.Println("*** ", value.Type())

	if value.IsNil() || value.Len() == 0 {
		// Create a new Elem of the Type inside the Slice
		log.Println("***1 ", reflect.New(elemOfType(value.Type().Elem())))
		return reflect.New(elemOfType(value.Type().Elem()))
	}

	// TODO
	// Not so good!
	log.Println("***2 ", value.Index(0).Type())
	return value.Index(0)

}

// If Value is a Ptr, return the Elem it points to
func elemOf(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

// If Type is a Ptr, return the Type of the Elem it points to
func elemOfType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// If Type is a Slice, return the Type of the Elem it stores
func elemOfSliceType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Slice {
		return t.Elem()
	}
	return t
}

// If Type is a Ptr, return the Type of the Elem it points to
// If Type is a Slice, return the Type of the Elem it stores
func mainElemOfType(t reflect.Type) reflect.Type {
	t = elemOfType(t)
	t = elemOfSliceType(t)
	t = elemOfType(t)
	return t

}

// If Type is not a Ptr, return the one Type that points to it
func ptrOfType(t reflect.Type) reflect.Type {
	if t.Kind() != reflect.Ptr {
		return reflect.PtrTo(t)
	}
	return t
}

// Return the true if passed one of those types
// Struct, *Struct, []Struct or []*Struct
func isValidValue(v reflect.Value) bool {
	_, isValid := getPtrValue(v)

	return isValid
}

// Return the true if the dependency is one of those types
// Interface, Struct, *Struct, []Struct or []*Struct
func isValidDependencyType(t reflect.Type) bool {
	t = elemOfType(t)

	return t.Kind() == reflect.Interface ||
		t.Kind() == reflect.Struct ||
		t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Struct
}

// Return true if this Type is a Slice or Ptr to Slice
func isSliceType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Slice
}

// Init methods should have no Output,
// it should alter the first argument as a pointer
// Or, at least, return itself
func isValidInit(method reflect.Method) bool {
	// If it has no output it's accepted
	if method.Type.NumOut() == 0 {
		return true
	}

	// It could return just one resource, itself
	if method.Type.NumOut() == 1 {
		//log.Printf("### TESTING: %s \n", method.Type)
		//log.Printf("### TESTING: %s, %s \n", method.Type.In(0), method.Type.Out(0))
		return method.Type.In(0) == method.Type.Out(0)
	}

	return false
}

package resource

import (
	"log"
	"net/http"
	"reflect"
	"strings"
)

// This constants will be used on the method above
var ResponseWriterPtrType = reflect.TypeOf((*http.ResponseWriter)(nil))
var ResponseWriterType = ResponseWriterPtrType.Elem()
var RequestPtrType = reflect.TypeOf((*http.Request)(nil))
var RequestType = RequestPtrType.Elem()

// This method return true if the received type is an context type
// It means that it doesn't need to be mapped and will be present in the context
// It also return an error message if user used *http.ResponseWriter or used http.Request
func isContextType(resourceType reflect.Type) bool {
	// Test if user used *http.ResponseWriter insted of http.ResponseWriter
	if resourceType.AssignableTo(ResponseWriterPtrType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, ResponseWriterType)
	}
	// Test if user used http.Request insted of *http.Request
	if resourceType.AssignableTo(RequestType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, RequestPtrType)
	}
	// Test if user used *ID insted of ID
	if resourceType.AssignableTo(IDPtrType) {
		log.Fatalf("You asked for %s when you should used %s", IDPtrType, IDType)
	}
	return resourceType.AssignableTo(ResponseWriterType) ||
		resourceType.AssignableTo(RequestPtrType) ||
		resourceType == IDType
}

// Return if this method should be mapped or not
// Methods starting with GET, POST, PUT, DELETE or HEAD should be mapped
func isMappedMethod(method reflect.Method) bool {
	return strings.HasPrefix(method.Name, "GET") ||
		strings.HasPrefix(method.Name, "POST") ||
		strings.HasPrefix(method.Name, "PUT") ||
		strings.HasPrefix(method.Name, "DELETE") ||
		strings.HasPrefix(method.Name, "HEAD")
}

// Return the Pointer to Struct Value if passed one of those types
// Struct, *Struct, []Struct or []*Struct
// Return one Pointer to Struct value and one Pointer to Slice value, if exists
func getPtrValues(value reflect.Value) (structPtr reflect.Value, slicePtr reflect.Value, isValid bool) {

	// If its an pointer, extract the element of that pointer
	value = elemOf(value)

	if value.Kind() == reflect.Slice {
		if value.IsNil() || value.Len() == 0 {
			// If it is nil or have no initial value
			// we will use an empty value of the slice type

			// Create a new Ptr to Slice value by the Slice Type passed
			slicePtr = reflect.New(value.Type())

			// Create a new Elem of the Type inside the Slice
			value = reflect.New(elemOfType(value.Type().Elem())).Elem()

		} else {
			// If the slice has an element inside,
			// so get the first element to represent the Struct Value of this slice

			// Create a new Ptr to Slice value by the Slice Type passed
			slicePtr = reflect.New(value.Type())
			// Add the current Value to this new Ptr to Slice Value
			slicePtr.Elem().Set(value)

			// Get the value of the first element in the Slice
			value = elemOf(value.Index(0))
		}
	}

	if value.Kind() == reflect.Struct {
		// Create a new Ptr to Struct value by this Struct type
		structPtr = reflect.New(value.Type())
		// Add the current value to this new Ptr to Struct value
		structPtr.Elem().Set(value)

		return structPtr, slicePtr, true
	}

	// Return false if the final value is not an Struct
	return structPtr, slicePtr, false
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
	_, _, isValid := getPtrValues(v)

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
func isSlice(t reflect.Type) bool {
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

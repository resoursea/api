package resource

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

// This constants will be used on the method above
var responseWriterPtrType = reflect.TypeOf((*http.ResponseWriter)(nil))
var responseWriterType = responseWriterPtrType.Elem()
var requestPtrType = reflect.TypeOf((*http.Request)(nil))
var requestType = requestPtrType.Elem()

// This method return true if the received type is an context type
// It means that it doesn't need to be mapped and will be present in the context
// It also return an error message if user used *http.ResponseWriter or used http.Request
func isContextType(resourceType reflect.Type) bool {
	// Test if user used *http.ResponseWriter insted of http.ResponseWriter
	if resourceType.AssignableTo(responseWriterPtrType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, responseWriterType)
	}
	// Test if user used http.Request insted of *http.Request
	if resourceType.AssignableTo(requestType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, requestPtrType)
	}
	// Test if user used *ID insted of ID
	if resourceType.AssignableTo(idPtrType) {
		log.Fatalf("You asked for %s when you should used %s", idPtrType, idType)
	}
	return resourceType.AssignableTo(responseWriterType) ||
		resourceType.AssignableTo(requestPtrType) ||
		resourceType == idType
}

func printResource(r *Resource, lvl int) {
	fmt.Printf("%-16s %-20s %-5v  ",
		strings.Repeat("|  ", lvl)+"|-["+r.Name+"]",
		r.Value.Type(), r.isSlice())

	if len(r.Tag) > 0 {
		fmt.Printf("tag: '%s' ", r.Tag)
	}

	if r.isSlice() {
		fmt.Printf("slice: %s ", r.SliceValue.Type())
	}

	if r.Anonymous {
		fmt.Printf("anonymous")
	}

	fmt.Println()

	for _, c := range r.Children {
		printResource(c, lvl+1)
	}
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

// If Value is a Ptr, return the Elem it points to
func elemOfType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// Return the true if passed one of those types
// Struct, *Struct, []Struct or []*Struct
func isValidValue(v reflect.Value) bool {
	_, _, isValid := getPtrValues(v)

	return isValid
}

// TODO
//
// Return the true if the dependency is one of those types
// Interface, Struct, *Struct, []Struct or []*Struct
func isValidDependencyType(t reflect.Type) bool {
	log.Println("Validating", t)
	//log.Println("Validating2", t.Elem())

	return true
}

package api

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

// Constants to test type equality
var (
	responseWriterPtrType = reflect.TypeOf((*http.ResponseWriter)(nil))
	tesponseWriterType    = responseWriterPtrType.Elem()
	requestPtrType        = reflect.TypeOf((*http.Request)(nil))
	requestType           = requestPtrType.Elem()
	errorSliceType        = reflect.TypeOf(([]error)(nil))
	errorType             = errorSliceType.Elem()
	errorNilValue         = reflect.New(errorType).Elem()
)

// This method return true if the received type is an context type
// It means that it doesn't need to be mapped and will be present in the context
// It also return an error message if user used *http.ResponseWriter or used http.Request
// Context types include error and []error types
func isContextType(resourceType reflect.Type) bool {
	// Test if user used *http.ResponseWriter insted of http.ResponseWriter
	if resourceType.AssignableTo(responseWriterPtrType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, tesponseWriterType)
	}
	// Test if user used http.Request insted of *http.Request
	if resourceType.AssignableTo(requestType) {
		log.Fatalf("You asked for %s when you should used %s", resourceType, requestPtrType)
	}
	// Test if user used ID insted of *ID
	if resourceType.AssignableTo(idType) {
		log.Fatalf("You asked for %s when you should used %s", idType, idPtrType)
	}

	return resourceType.AssignableTo(tesponseWriterType) ||
		resourceType.AssignableTo(requestPtrType) ||
		resourceType.AssignableTo(errorType) ||
		resourceType.AssignableTo(errorSliceType) ||
		resourceType == idPtrType
}

// Return one Ptr to the given Value...
// - If receive Struct or *Struct, resturn *Struct
// - If receive []Struct, return *[]Struct
// - If receive []*Struct, return *[]*Struct
func ptrOfValue(value reflect.Value) reflect.Value {
	ptr := reflect.New(elemOfType(value.Type()))
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		ptr.Elem().Set(value.Elem())
	}
	if value.Kind() == reflect.Struct {
		ptr.Elem().Set(value)
	}
	return ptr
}

// Return the Value of the Ptr to Elem of the inside the given Slice
// []struct, *[]struct, *[]*struct -> return struct Value
func elemOfSliceValue(value reflect.Value) reflect.Value {

	// If Value is a Ptr, get the Elem it points to
	sliceValue := elemOfValue(value)

	// Type of the Elem inside the given Slice
	// []struct or []*struct -> return struct type
	elemType := elemOfType(sliceValue.Type().Elem())

	// Creates a new Ptr to Elem of the Type this Slice stores
	ptrToElem := reflect.New(elemType)

	if sliceValue.IsNil() || sliceValue.Len() == 0 {
		// If given Slice is null or it has nothing inside it
		// return the new empty Value of the Elem inside this slice
		return ptrToElem
	}

	// If this slice has an Elem inside it
	// and set the value of the first Elem inside this Slice
	ptrToElem.Elem().Set(elemOfValue(sliceValue.Index(0)))

	return ptrToElem
}

// If Value is a Ptr, return the Elem it points to
func elemOfValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		// It occours if an Struct has an Field that points itself
		// so it should never occours, and will be cauch by the check for Circular Dependency
		if !value.Elem().IsValid() {
			value = reflect.New(elemOfType(value.Type()))
		}
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

// Return the true if is an exported field of those of types
// Struct, *Struct, []Struct or []*Struct
// Check if this field is exported, begin with UpperCase: v.CanInterface()
// and if this field is valid fo create Resources: Structs or Slices of Structs
func isValidValue(v reflect.Value) bool {
	if !v.CanInterface() {
		return false
	}
	if mainElemOfType(v.Type()).Kind() == reflect.Struct {
		return true
	}
	return false
}

// Return the true if the dependency is one of those types
// Interface, Struct, *Struct, []Struct or []*Struct
func isValidDependencyType(t reflect.Type) error {
	t = elemOfType(t)

	if t.Kind() == reflect.Interface || t.Kind() == reflect.Struct ||
		t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Struct {
		return nil
	}

	return fmt.Errorf("Type %s is not allowed as dependency", t)
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
func isValidInit(method reflect.Method) error {

	// Test if Init method return itself and/or error types
	// just one of each type is accepted
	itself := false
	err := false

	//errorType := reflect.TypeOf(errors.New("432"))

	// Method Struct owner Type
	owner := mainElemOfType(method.Type.In(0))
	for i := 0; i < method.Type.NumOut(); i++ {
		t := mainElemOfType(method.Type.Out(i))
		if t == owner && !itself {
			itself = true
			continue
		}
		if t == errorType && !err {
			err = true
			continue
		}

		return fmt.Errorf("Resource %s has an invalid Init method %s. "+
			"It can't outputs %s \n", method.Type.In(0), method.Type, t)
	}

	return nil
}

// Return true if given StructField is an exported Field
// return false if is an unexported Field
func isExportedField(field reflect.StructField) bool {
	firstChar := string([]rune(field.Name)[0])
	return firstChar == strings.ToUpper(firstChar)
}

// Return a new empty Value for one of these Types
// Struct, *Struct, Slice, *Slice
func newEmptyValue(t reflect.Type) (reflect.Value, error) {
	t = elemOfType(t)
	if t.Kind() == reflect.Struct || t.Kind() == reflect.Slice {
		return reflect.New(t), nil // A new Ptr to Struct of this type
	}
	return reflect.Value{}, fmt.Errorf("Can't create an empty Value for type  %s", t)
}

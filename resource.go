package api

import (
	"fmt"
	"reflect"
	"strings"
)

// We are storing the Pointer to Struct value and Pointer to Slice as Value
type resource struct {
	name      string
	value     reflect.Value
	parent    *resource
	children  []*resource
	extends   []*resource // Spot for Anonymous fields
	anonymous bool        // Is Anonymous field?
	tag       reflect.StructTag
	isSlice   bool
}

// Create a new Resource tree based on given Struct, its Struct Field and its Resource parent
func newResource(value reflect.Value, field reflect.StructField, parent *resource) (*resource, error) {
	// Check if the value is valid, valid values are:
	// struct, *struct, []struct, *[]struct, *[]*struct
	if !isValidValue(value) {
		return nil, fmt.Errorf("Can't create a Resource with type %s", value.Type())
	}

	// Garants we are working with a Ptr to Struct or Slice
	value = ptrOfValue(value)

	//log.Println("Scanning Struct:", value.Type(), "name:", strings.ToLower(field.Name), value.Interface())

	r := &resource{
		name:      strings.ToLower(field.Name),
		value:     value,
		parent:    parent,
		children:  []*resource{},
		extends:   []*resource{},
		anonymous: field.Anonymous,
		tag:       field.Tag,
		isSlice:   isSliceType(value.Type()),
	}

	// Check for circular dependency !!!
	exist, p := r.existParentOfType(r)
	if exist {
		return nil, fmt.Errorf("The resource %s as '%s' have an circular dependency in %s as '%s'",
			r.value.Type(), r.name, p.value.Type(), p.name)
	}

	// If it is slice, scan the Elem of this slice
	if r.isSlice {

		elemValue := elemOfSliceValue(value)

		elem, err := newResource(elemValue, field, r)
		if err != nil {
			return nil, err
		}

		//r.Elem = elem
		r.addChild(elem)

		return r, nil
	}

	for i := 0; i < value.Elem().Type().NumField(); i++ {

		field := value.Elem().Type().Field(i)
		fieldValue := value.Elem().Field(i)

		//log.Println("Field:", field.Name, field.Type, "of", value.Elem().Type(), "is valid", isValidValue(fieldValue))

		// Check if this field is exported: fieldValue.CanInterface()
		// and if this field is valid fo create Resources: Structs or Slices of Structs
		if isValidValue(fieldValue) {
			child, err := newResource(fieldValue, field, r)
			if err != nil {
				return nil, err
			}
			err = r.addChild(child)
			if err != nil {
				return nil, err
			}
		}
	}

	return r, nil
}

// The child should be added to the first non anonymous parent
// An anonymous field indicates that the containing non anonymous parent Struct
// should have all the fields and methos this anonymous field has
func (parent *resource) addChild(child *resource) error {
	//log.Printf("%s Anonymous: %v adding Child %s",
	//	parent.Value.Type(), parent.Anonymous, child.Value.Type())

	// Just add the child to the first non anonymous parent
	if parent.anonymous {
		parent.parent.addChild(child)
		return nil
	}

	// If this child is Anonymous, its father will extends its behavior
	if child.anonymous {
		parent.extends = append(parent.extends, child)
		return nil
	}

	// Two children can't have the same name, check it before insert them
	for _, sibling := range parent.children {
		if child.name == sibling.name {
			return fmt.Errorf("Two resources have the same name '%s' \nR1: %s, R2: %s, Parent: %s",
				child.name, sibling.value.Type(), child.value.Type(), parent.value.Type())
		}
	}

	parent.children = append(parent.children, child)
	return nil
}

// Return Value of the implementation of some Interface,
// this Resource that satisfies this interface
// should be present in this Resource children or in its parents children recursively
// If requested type is an Struct return the initial Value of this Type, if exists,
// if Struct type not contained on the resource tree, create a new empty Value for this Type
func (r *resource) valueOf(t reflect.Type) (reflect.Value, error) {

	for _, child := range r.children {
		if child.isType(t) {
			return child.value, nil
		}
	}

	// Go recursively until reaching the root
	if r.parent != nil {
		return r.parent.valueOf(t)
	}

	// Testing the root of the Resource Tree
	ok := r.isType(t)
	if ok {
		return r.value, nil
	}

	// At this point we tested all Resources in the tree
	// If we are searching for an Interface, and noone implements it
	// so we shall throws an error informing user to satisfy this Interface in the Resource Tree
	if t.Kind() == reflect.Interface {
		return reflect.Value{}, fmt.Errorf(
			"Not found any Resource that implements the Interface "+
				"type  %s in the Resource tree %s", t, r)
	}

	// If it isn't present in the Resource tree
	// and this type we are searching isn't an interface
	// So we will use an empty new value for it!
	return newEmptyValue(t)
}

// Return true if this Resrouce is from by this Type
func (r *resource) isType(t reflect.Type) bool {

	if t.Kind() == reflect.Interface {
		if r.value.Type().Implements(t) {
			return true
		}
	}

	// If its not an Ptr to Struct or to Slice
	// so thest the type of this Ptr
	if r.value.Type() == ptrOfType(t) {
		return true
	}

	return false
}

// Return true any of its father have the same type of this resrouce
// This method prevents for Circular Dependency
func (r *resource) existParentOfType(re *resource) (bool, *resource) {
	if r.parent != nil {
		if r.parent.value.Type() == re.value.Type() {
			return true, r.parent
		}
		return r.parent.existParentOfType(re)
	}
	return false, nil
}

func (r *resource) String() string {

	name := "[" + r.name + "] "

	response := fmt.Sprintf("%-20s ", name+r.value.Type().String())

	return response
}

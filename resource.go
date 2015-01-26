package api

import (
	"fmt"
	"reflect"
	"strings"
)

// We are storing the Pointer to Struct value and Pointer to Slice as Value
type resource struct {
	Name      string
	Value     reflect.Value
	Parent    *resource
	Children  []*resource
	Extends   []*resource // Spot for Anonymous fields
	Anonymous bool        // Is Anonymous field?
	Tag       reflect.StructTag
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
		Name:      strings.ToLower(field.Name),
		Value:     value,
		Parent:    parent,
		Children:  []*resource{},
		Extends:   []*resource{},
		Anonymous: field.Anonymous,
		Tag:       field.Tag,
		isSlice:   isSliceType(value.Type()),
	}

	// Check for circular dependency !!!
	exist, p := r.existParentOfType(r)
	if exist {
		return nil, fmt.Errorf("The resource %s as '%s' have an circular dependency in %s as '%s'",
			r.Value.Type(), r.Name, p.Value.Type(), p.Name)
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
	if parent.Anonymous {
		parent.Parent.addChild(child)
		return nil
	}

	// If this child is Anonymous, its father will extends its behavior
	if child.Anonymous {
		parent.Extends = append(parent.Extends, child)
		return nil
	}

	// Two children can't have the same name, check it before insert them
	for _, sibling := range parent.Children {
		if child.Name == sibling.Name {
			return fmt.Errorf("Two resources have the same name '%s' \nR1: %s, R2: %s, Parent: %s",
				child.Name, sibling.Value.Type(), child.Value.Type(), parent.Value.Type())
		}
	}

	parent.Children = append(parent.Children, child)
	return nil
}

// Return Value of the implementation of some Interface,
// this Resource that satisfies this interface
// should be present in this Resource children or in its parents children recursively
// If requested type is an Struct return the initial Value of this Type, if exists,
// if Struct type not contained on the resource tree, create a new empty Value for this Type
func (r *resource) valueOf(t reflect.Type) (reflect.Value, error) {

	for _, child := range r.Children {
		if child.isType(t) {
			return child.Value, nil
		}
	}

	// Go recursively until reaching the root
	if r.Parent != nil {
		return r.Parent.valueOf(t)
	}

	// Testing the root of the Resource Tree
	ok := r.isType(t)
	if ok {
		return r.Value, nil
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
		if r.Value.Type().Implements(t) {
			return true
		}
	}

	// If its not an Ptr to Struct or to Slice
	// so thest the type of this Ptr
	if r.Value.Type() == ptrOfType(t) {
		return true
	}

	return false
}

// Return true any of its father have the same type of this resrouce
// This method prevents for Circular Dependency
func (r *resource) existParentOfType(re *resource) (bool, *resource) {
	if r.Parent != nil {
		if r.Parent.Value.Type() == re.Value.Type() {
			return true, r.Parent
		}
		return r.Parent.existParentOfType(re)
	}
	return false, nil
}

func (r *resource) String() string {

	name := "[" + r.Name + "] "

	response := fmt.Sprintf("%-20s ", name+r.Value.Type().String())

	return response
}

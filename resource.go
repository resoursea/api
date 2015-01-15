package api

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// We are storing the Pointer to Struct value and Pointer to Slice as Value
type Resource struct {
	Name      string
	Value     reflect.Value
	Parent    *Resource
	Elem      *Resource // If it is an Slice Resource, it points to the Elem Resource
	Children  []*Resource
	Extends   []*Resource // Spot for Anonymous fields
	Anonymous bool        // Is Anonymous field?
	Tag       reflect.StructTag
	IsSlice   bool
}

// Creates a new Resource tree based on given Struct
// Receives the Struct to be mapped in a new Resource Tree,
// it also receive the Field name and Field tag as optional arguments
func NewResource(object interface{}, args ...string) (*Resource, error) {

	value := reflect.ValueOf(object)

	name := value.Type().Name()
	tag := ""

	// Defining a name as an opitional secound argument
	if len(args) >= 1 {
		name = args[0]
	}

	// Defining a tag as an opitional thrid argument
	if len(args) >= 2 {
		tag = args[1]
	}

	field := reflect.StructField{
		Name:      name,
		Tag:       reflect.StructTag(tag),
		Anonymous: false,
	}

	return newResource(value, field, nil)
}

// Create a new Resource tree based on given Struct, its Struct Field and its Resource parent
func newResource(value reflect.Value, field reflect.StructField, parent *Resource) (*Resource, error) {

	// If its a Ptr or a Slice or both, get the Ptr to this type
	value, err := validPtrOfValue(value)

	if err != nil {
		return nil, err
	}

	log.Println("Scanning Struct:", value.Type(), "name:", strings.ToLower(field.Name))

	resource := &Resource{
		Name:      strings.ToLower(field.Name),
		Value:     value,
		Parent:    parent,
		Children:  []*Resource{},
		Extends:   []*Resource{},
		Anonymous: field.Anonymous,
		Tag:       field.Tag,
		IsSlice:   isSliceType(value.Type()),
	}

	if parent != nil {
		//log.Printf("CHECKING CD, parent: %s, child: %s \n", parent.Value.Type(), resource.Value.Type())

		// Check for circular dependency !!!
		exist, p := parent.existParentOfType(resource)
		if exist {
			printResourceStack(resource, resource)
			return nil, errors.New(fmt.Sprintf("The resource %s as '%s' have an circular dependency in %s as '%s'",
				resource.Value.Type(), resource.Name, p.Value.Type(), p.Name))
		}

	}

	// If it is slice, scan the Elem of this slice
	if resource.IsSlice {

		elemValue := slicePtrToElemValue(value)

		elem, err := newResource(elemValue, field, resource)
		if err != nil {
			return nil, err
		}

		resource.Elem = elem

		return resource, nil
	}

	for i := 0; i < value.Elem().Type().NumField(); i++ {

		field := value.Elem().Type().Field(i)
		fieldValue := value.Elem().Field(i)

		log.Println("Field:", field.Name, field.Type, "of", value.Elem().Type())

		if isExportedField(field) && isValidValue(fieldValue) {
			child, err := newResource(fieldValue, field, resource)
			if err != nil {
				return nil, err
			}
			resource.addChild(child)
		}
	}

	return resource, nil
}

// The child should be added to the first non anonymous parent
// An anonymous field indicates that the containing non anonymous parent Struct
// should have all the fields and methos this anonymous field has
func (parent *Resource) addChild(child *Resource) {
	//log.Printf("%s Anonymous: %v adding Child %s",
	//	parent.Value.Type(), parent.Anonymous, child.Value.Type())

	// Just add the child to the first non anonymous parent
	if parent.Anonymous {
		parent.Parent.addChild(child)
		return
	}

	// If this child is Anonymous, its father will extends its behavior
	if child.Anonymous {
		parent.Extends = append(parent.Extends, child)
		return
	}

	// Two children can't have the same name, check it before insert them
	for _, sibling := range parent.Children {
		if child.Name == sibling.Name {
			log.Fatalf("Thwo resources have the same name '%s' \nR1: %s, R2: %s, Parent: %s",
				child.Name, sibling.Value.Type(), child.Value.Type(), parent.Value.Type())
		}
	}

	parent.Children = append(parent.Children, child)
}

// Return Value of the implementation of some Interface
// This Resource that satisfies this interface
// should be present in this Resource or in its parents recursively
func (r *Resource) valueOf(t reflect.Type) (reflect.Value, error) {

	for _, child := range r.Children {
		ok := child.isType(t)
		if ok {
			return child.Value, nil
		}
	}

	// Go recursively until reaching the root
	if r.Parent != nil {
		return r.Parent.valueOf(t)
	}

	// The special case that to get the root value
	if r.Parent == nil {
		ok := r.isType(t)
		if ok {
			return r.Value, nil
		}
	}

	// If it isn't present in the Resource tree
	// and this type we are searching isn't an interface
	// So we will use an empty new value for it!

	// For Struct
	if t.Kind() == reflect.Struct {
		return reflect.New(t), nil // A new Ptr to Struct of this type
	}
	// For Ptr to Struct
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return reflect.New(t.Elem()), nil // A new Ptr to Struct of this type
	}
	// For Slice
	if t.Kind() == reflect.Slice {
		return reflect.New(t), nil
	}
	// For Ptr to Slice
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Slice {
		return reflect.New(t.Elem()), nil
	}

	return reflect.Value{}, fmt.Errorf(
		"Not found any Resource that implements the type  %s in the tree %s",
		t, r.Value.Type())
}

// Return true if this Resrouce is from by this Type
func (r *Resource) isType(t reflect.Type) bool {

	if t.Kind() == reflect.Interface {
		if r.Value.Type().Implements(t) {
			return true
		}
	}

	// If its not an Ptr to Struct or to Slice
	// so get the type of this Ptr
	t = ptrOfType(t)

	if r.Value.Type() == t {
		return true
	}

	return false
}

// Return true any of its father have the same type of this resrouce
// This method prevents for Circular Dependency
func (r *Resource) existParentOfType(resource *Resource) (bool, *Resource) {

	if r.isSameType(resource) {
		return true, r
	}

	if r.Parent != nil {
		return r.Parent.existParentOfType(resource)
	}

	return false, nil
}

// Return true if this Resrouce is from by this Type
func (r *Resource) isSameType(resource *Resource) bool {
	return r.Value.Type() == resource.Value.Type()
}

func (r *Resource) String() string {

	name := "[" + r.Name + "]"

	response := fmt.Sprintf("%-14s %24s", name, r.Value.Type().String())

	return response
}

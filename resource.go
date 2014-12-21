package resource

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// We are storing the Pointer to Struct value
// and Pointer to slice

type Resource struct {
	Name       string
	Value      reflect.Value
	SliceValue reflect.Value
	Parent     *Resource
	Children   []*Resource
	Anonymous  bool // Is Anonymous Field?
	Tag        reflect.StructTag
}

// Creates a new resource
// Receives a
func NewResource(object interface{}, args ...string) *Resource {

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

	log.Printf("field: %#v\n", field)

	return scanStruct(value, field, nil)
}

func scanStruct(value reflect.Value, field reflect.StructField, parent *Resource) *Resource {

	// If its a Ptr or a Slice or both, get the Ptr to Struct value
	value, sliceValue, isValid := getPtrValues(value)

	if !isValid {
		log.Fatal("You should pass an struct or an slice of structs")
	}

	log.Println("Scanning Struct:", value.Type(), "slice:")

	resource := &Resource{
		Name:       strings.ToLower(field.Name),
		Value:      value,
		SliceValue: sliceValue,
		Parent:     parent,
		Children:   []*Resource{},
		Anonymous:  field.Anonymous,
		Tag:        field.Tag,
	}

	// Add this resource as child of its parent if it is not the root
	// If this resource is an anonymous field, add as anonymous
	if parent != nil {
		parent.addChild(resource)
	}

	log.Println("Scanning Fields:", value.Elem().Type())

	for i := 0; i < value.Elem().Type().NumField(); i++ {

		field := value.Elem().Type().Field(i)
		fieldValue := value.Elem().Field(i)

		log.Println("Field:", field.Name, field.Type, "of", value.Elem().Type())

		if isValidValue(fieldValue) {
			scanStruct(fieldValue, field, resource)
		}
	}

	return resource
}

// The child should be added to the first non anonymous father
// An anonymous field indicates that the containing non anonymous parent Struct
// should have all the fields and methos this anonymous field has
func (parent *Resource) addChild(resource *Resource) {
	log.Printf("%s Anonymous: %v adding Child %s",
		parent.Value.Type(), parent.Anonymous, resource.Value.Type())

	if parent.Anonymous {
		parent.Parent.addChild(resource)
	} else {
		for _, child := range parent.Children {
			if child.Name == resource.Name {
				log.Fatalf("Thwo resources have the same name '%s' \nR1: %s, R2: %s, Parent: %s",
					resource.Name, child.Value.Type(), resource.Value.Type(), parent.Value.Type())
			}
		}
		parent.Children = append(parent.Children, resource)
	}
}

// Return Value of the implementation of some Interface
// This Resource that satisfies this interface
// should be present in this Resource or in its parents recursively
func (r *Resource) valueOf(dependencyType reflect.Type) (reflect.Value, error) {

	t := elemOfType(dependencyType)

	for _, child := range r.Children {
		if t.Kind() == reflect.Interface {
			if child.Value.Type().Implements(t) {
				return child.Value, nil
			}
		} else {
			if child.Value.Type() == t {
				return child.Value, nil
			}
		}
	}

	// Go recursively until reaching the root
	if r.Parent != nil {
		return r.Parent.valueOf(t)
	}

	// The special case that to get the root value
	if r.Parent == nil {
		if t.Kind() == reflect.Interface {
			if r.Value.Type().Implements(t) {
				return r.Value, nil
			}
		} else {
			if r.Value.Type() == t {
				return r.Value, nil
			}
		}
	}

	return reflect.Value{}, fmt.Errorf(
		"Not found any Resource that implements the type  %s in the tree %s",
		t, r.Value.Type())
}

// Return true if this Resrouce have an Slice type
func (r *Resource) isSlice() bool {
	return r.SliceValue.IsValid()
}

package resource

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// We are storing the Pointer to Struct value and Pointer to Slice as Value
type Resource struct {
	Name       string
	Value      reflect.Value
	SliceValue reflect.Value
	Parent     *Resource
	Children   []*Resource
	Extends    []*Resource // Spot for Anonymous fields
	Anonymous  bool        // Is Anonymous field?
	Tag        reflect.StructTag
}

// Creates a new resource
// Receives the object to be mappen in a new Resource
// and receive the field name and field tag as optional arguments
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

	//log.Printf("field: %#v\n", field)

	return scanStruct(value, field, nil)
}

func scanStruct(value reflect.Value, field reflect.StructField, parent *Resource) *Resource {

	// If its a Ptr or a Slice or both, get the Ptr to Struct value
	value, sliceValue, isValid := getPtrValues(value)

	if !isValid {
		log.Fatal("You should pass an struct or an slice of structs")
	}

	//log.Println("Scanning Struct:", value.Type(), "slice:")

	resource := &Resource{
		Name:       strings.ToLower(field.Name),
		Value:      value,
		SliceValue: sliceValue,
		Parent:     parent,
		Children:   []*Resource{},
		Extends:    []*Resource{},
		Anonymous:  field.Anonymous,
		Tag:        field.Tag,
	}

	// Add this resource as child of its parent if it is not the root
	// If this resource is an anonymous field, add as anonymous
	if parent != nil {

		// Check for circular dependency !!!
		exist, p := parent.existParent(resource)
		if exist {
			printResourceStack(resource, resource)
			log.Fatalf("The resource %s as '%s' have an circular dependency in %s as '%s'",
				resource.Type(), resource.Name, p.Type(), p.Name)
		}

		parent.addChild(resource)
	}

	//log.Println("Scanning Fields:", value.Elem().Type())

	for i := 0; i < value.Elem().Type().NumField(); i++ {

		field := value.Elem().Type().Field(i)
		fieldValue := value.Elem().Field(i)

		//log.Println("Field:", field.Name, field.Type, "of", value.Elem().Type())

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
	//log.Printf("%s Anonymous: %v adding Child %s",
	//	parent.Value.Type(), parent.Anonymous, resource.Value.Type())

	if parent.Anonymous {
		parent.Parent.addChild(resource)
		return
	}

	// If this Resource is Anonymous, its father will extends its behavior
	if resource.Anonymous {
		parent.Extends = append(parent.Extends, resource)
		return
	}

	// Two children can't have the same name, check it before insert them
	for _, child := range parent.Children {
		if child.Name == resource.Name {
			log.Fatalf("Thwo resources have the same name '%s' \nR1: %s, R2: %s, Parent: %s",
				resource.Name, child.Value.Type(), resource.Value.Type(), parent.Value.Type())
		}
	}
	parent.Children = append(parent.Children, resource)
}

// Return Value of the implementation of some Interface
// This Resource that satisfies this interface
// should be present in this Resource or in its parents recursively
func (r *Resource) valueOf(t reflect.Type) (reflect.Value, error) {

	for _, child := range r.Children {
		v, ok := child.isType(t)
		if ok {
			return v, nil
		}
	}

	// Go recursively until reaching the root
	if r.Parent != nil {
		return r.Parent.valueOf(t)
	}

	// The special case that to get the root value
	if r.Parent == nil {
		v, ok := r.isType(t)
		if ok {
			return v, nil
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
func (r *Resource) isType(t reflect.Type) (reflect.Value, bool) {

	if t.Kind() == reflect.Interface {
		if r.Value.Type().Implements(t) {
			return r.Value, true
		}
		if r.SliceValue.Type().Implements(t) {
			return r.SliceValue, true
		}

		// Not found
		return r.Value, false
	}

	// If its not an Ptr to Struct or to Slice
	// so get the type of this Ptr
	t = ptrOfType(t)

	if r.Value.Type() == t {
		return r.Value, true
	}

	if r.isSlice() && r.SliceValue.Type() == t {
		return r.SliceValue, true
	}

	// Not found
	return reflect.Value{}, false
}

// Return true if this Resrouce have an Slice type
func (r *Resource) isSlice() bool {
	return r.SliceValue.IsValid()
}

// Return true any of its father have the same type of this resrouce
// this method prevent for Circular Dependency
func (r *Resource) existParent(resource *Resource) (bool, *Resource) {

	if r.isEqual(resource) {
		return true, r
	}

	if r.Parent != nil {
		return r.Parent.existParent(resource)
	}

	return false, nil
}

// Return true if this Resrouce is from by this Type
func (r *Resource) isEqual(resource *Resource) bool {
	return r.Type() == resource.Type()
}

// Return true if this Resrouce is from by this Type
func (r *Resource) Type() reflect.Type {
	return r.Value.Type()
}

// Return true if this Resrouce is from by this Type
func (r *Resource) SliceType() reflect.Type {
	return r.SliceValue.Type()
}

func (r *Resource) String() string {

	name := "[" + r.Name + "]"

	response := fmt.Sprintf("%-14s %24s", name, r.Type().String())

	if r.isSlice() {
		response += fmt.Sprintf(" %24s", "[]"+r.SliceType().String())
	}

	return response
}

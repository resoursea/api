package resource

import (
	"log"
	"reflect"
	"strings"
)

//"net/http"

type Route struct {
	// The resource name
	URI string
	// Methods by its name, the HTTP action name
	Methods map[string]*Method

	// map[URI]*Route
	Children map[string]*Route
}

// Receives the Root Resource and interate recursively
// creating the Route tree
func NewRoute(resource *Resource) *Route {

	log.Println("Building routes for", resource.Value.Type())

	r := &Route{
		URI:      strings.ToLower(resource.Value.Type().Name()),
		Methods:  make(map[string]*Method),
		Children: make(map[string]*Route),
	}

	r.ScanMethods(resource.Value, resource)

	return r
}

// Scan the methods of some type
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {}
func (r *Route) ScanMethods(value reflect.Value, resource *Resource) {

	// Get the *Resource type
	ptrType := reflect.New(value.Type()).Type()

	log.Println("Searching methods from", ptrType)

	// Searching for mapped methods
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)

		if isMappedMethod(method) {
			m := NewMethod(method, resource)
			r.Methods[m.Name] = m
		}
	}
}

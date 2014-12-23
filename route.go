package resource

import (
	"log"
)

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

	log.Println("###Building routes for", resource.Value.Type())

	route := &Route{
		URI:      resource.Name,
		Methods:  make(map[string]*Method),
		Children: make(map[string]*Route),
	}

	// This route pic the methods of the resource father...
	route.ScanMethods(resource)
	// ... and all its Exstends
	for _, extend := range resource.Extends {
		route.ScanMethods(extend)
	}

	// Check for Circular Dependency
	// on the Dependencies of each method
	checkCircularDependency(route)

	// Creating recursivelly
	for _, child := range resource.Children {
		childRoute := NewRoute(child)
		if len(childRoute.Methods) > 0 {
			route.Children[childRoute.URI] = childRoute
		}
	}

	return route
}

// Scan the methods of some type
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {}

// Getting the resource to be scanned, cause we will scan for
// the anonymous resource fields in the same route
func (r *Route) ScanMethods(resource *Resource) {

	// Get the *Resource type
	ptrType := resource.Value.Type()

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

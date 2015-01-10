package resource

import (
	"errors"
	"log"
	"reflect"
	"strings"
)

type Route struct {
	// The resource name
	URI string

	// The main type of this route
	Value reflect.Value

	// Methods by its name, the HTTP action name
	// They are separated in methods for set and for unit
	SliceHandlers map[string]*Handler
	Handlers      map[string]*Handler

	// map[URI]*Route
	Children map[string]*Route

	// True if the resource is an Slice of Resources
	IsSlice bool
}

// Receives the Root Resource and interate recursively
// creating the Route tree
func NewRoute(resource *Resource) *Route {

	//log.Println("###Building routes for", resource.Value.Type())

	route := &Route{
		URI:           resource.Name,
		Value:         resource.Value,
		SliceHandlers: make(map[string]*Handler),
		Handlers:      make(map[string]*Handler),
		Children:      make(map[string]*Route),
		IsSlice:       resource.isSlice(),
	}

	// This route take the methods of the main resource...
	route.ScanResource(resource)

	// Check for Circular Dependency
	// on the Dependencies of each method
	err := circularDependency(route)
	if err != nil {
		log.Fatal(err)
	}

	// Creating routes recursivelly for each resource child
	for _, child := range resource.Children {
		err := route.addChild(NewRoute(child))
		if err != nil {
			log.Panicln(err)
		}

	}

	return route
}

// Scan the methods of some type
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {}
func (r *Route) ScanResource(resource *Resource) {
	// If this resource has an Slice Type, we should scan it
	if resource.isSlice() {
		r.ScanType(resource.SliceValue.Type(), resource)
	}
	// Scan the Ptr to the main Struct
	r.ScanType(resource.Value.Type(), resource)

	// ... and all the resources it Exstends will be mapped too
	for _, extend := range resource.Extends {
		r.ScanResource(extend)
	}

}

// Scan the methods from one Type and add it to the route
// This type could be []*Resource or just *Resource
func (r *Route) ScanType(t reflect.Type, resource *Resource) {

	//log.Println("Scanning methods from type", t, isSlice(t))

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		if isMappedMethod(m) {

			method := NewMethod(m)
			h := NewHandler(resource, method)

			if isSlice(t) {
				r.SliceHandlers[h.Name] = h
			} else {
				r.Handlers[h.Name] = h
			}

		}
	}

}

// Return false if this route have no methods declared
func (r *Route) hasHandler() bool {

	if len(r.Handlers) > 0 {
		return true
	}

	if len(r.SliceHandlers) > 0 {
		return true
	}

	for _, child := range r.Children {
		if child.hasHandler() {
			return true
		}
	}

	return false
}

// Add a new route child
func (r *Route) addChild(child *Route) error {
	// Add this route to the tree only if it has handlers
	if child.hasHandler() {

		// Test if this URI wasn't in use yet
		_, exist := r.Children[child.URI]
		if exist {
			return errors.New("Route " + r.URI + " already has child " + child.URI)
		}

		r.Children[child.URI] = child
	}
	return nil
}

// Return the Route from the especified URI
func (r *Route) getHandler(uri []string, method string) (*Handler, IDMap, error) {

	useSliceMethods := false

	next := strings.ToLower(uri[0])
	uri = uri[1:len(uri)]

	// Store the IDs of the resources took in the url
	ids := IDMap{}

	//log.Println("Getting Handler", next)

	route, exist := r.Children[next]
	if !exist {
		return nil, ids, errors.New(
			"Resource '" + next + "' doesn't exist inside: " + r.URI)
	}

	// If this route is an Slice,
	// so the next URI will by an ID, if it exist
	if route.IsSlice {
		if len(uri) > 0 && len(uri[0]) > 0 {
			ids[route.Value.Type()] = reflect.ValueOf(ID(uri[0]))
			uri = uri[1:len(uri)]
		} else {
			// It will only use the slice Methods
			// if user accesed an Sliced route
			// and doesn't give any ID
			useSliceMethods = true
		}
	}

	var h *Handler

	// If we need to search deeply in the tree
	if len(uri) > 0 && len(uri[0]) > 0 {

		childHandler, childIds, err := route.getHandler(uri, method)
		if err != nil {
			return nil, ids, err
		}
		h = childHandler
		ids.extend(childIds)

	} else {

		// If we are on the final Route user requested
		if useSliceMethods {
			h, exist = route.SliceHandlers[method]
		} else {
			h, exist = route.Handlers[method]
		}

		if !exist {
			msg := "Resource '" + route.URI + "' doesn't have method " + method
			if useSliceMethods {
				msg += " in the slice"
			} else {
				msg += " in the element"
			}
			return nil, ids, errors.New(msg)
		}

	}

	return h, ids, nil
}

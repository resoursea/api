package api

import (
	"errors"
	"log"
	"reflect"
	"strings"
)

type route struct {
	// The resource name
	URI string

	// The main type of this route
	Value reflect.Value

	// Methods by its name, the HTTP action name
	// They are separated in methods for set and for unit
	SliceHandlers map[string]*handler
	Handlers      map[string]*handler

	// map[URI]*route
	Children map[string]*route

	// True if the resource is an Slice of Resources
	IsSlice bool
}

// Receives the Root Resource and interate recursively
// creating the Route tree
func NewRoute(r *resource) *route {

	//log.Println("###Building routes for", resource.Value.Type())

	ro := &route{
		URI:           r.Name,
		Value:         r.Value,
		SliceHandlers: make(map[string]*handler),
		Handlers:      make(map[string]*handler),
		Children:      make(map[string]*route),
		IsSlice:       r.isSlice(),
	}

	// This route take the methods of the main resource...
	ro.ScanResource(r)

	// Check for Circular Dependency
	// on the Dependencies of each method
	err := checkCircularDependency(ro)
	if err != nil {
		log.Fatal(err)
	}

	// Creating routes recursivelly for each resource child
	for _, child := range r.Children {
		err := ro.addChild(NewRoute(child))
		if err != nil {
			log.Panicln(err)
		}

	}

	return ro
}

// Scan the methods of some type
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {}
func (ro *route) ScanResource(r *resource) {
	// If this resource has an Slice Type, we should scan it
	if r.isSlice() {
		ro.ScanType(r.SliceValue.Type(), r)
	}
	// Scan the Ptr to the main Struct
	ro.ScanType(r.Value.Type(), r)

	// ... and all the resource it Exstends will be mapped too
	for _, extend := range r.Extends {
		ro.ScanResource(extend)
	}

}

// Scan the methods from one Type and add it to the route
// This type could be []*Resource or just *Resource
func (ro *route) ScanType(t reflect.Type, r *resource) {

	//log.Println("Scanning methods from type", t, isSlice(t))

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		if isMappedMethod(m) {

			method := newMethod(m)
			h := newHandler(r, method)

			if isSlice(t) {
				ro.SliceHandlers[h.Name] = h
			} else {
				ro.Handlers[h.Name] = h
			}

		}
	}

}

// Return false if this route have no methods declared
func (ro *route) hasHandler() bool {

	if len(ro.Handlers) > 0 {
		return true
	}

	if len(ro.SliceHandlers) > 0 {
		return true
	}

	for _, child := range ro.Children {
		if child.hasHandler() {
			return true
		}
	}

	return false
}

// Add a new route child
func (ro *route) addChild(child *route) error {
	// Add this route to the tree only if it has handlers
	if child.hasHandler() {

		// Test if this URI wasn't in use yet
		_, exist := ro.Children[child.URI]
		if exist {
			return errors.New("Route " + ro.URI + " already has child " + child.URI)
		}

		ro.Children[child.URI] = child
	}
	return nil
}

// Return the Route from the especified URI
func (ro *route) getHandler(uri []string, method string) (*handler, idMap, error) {

	useSliceMethods := false

	next := strings.ToLower(uri[0])
	uri = uri[1:len(uri)]

	// Store the IDs of the resources took in the url
	ids := idMap{}

	//log.Println("Getting Handler", next)

	route, exist := ro.Children[next]
	if !exist {
		return nil, ids, errors.New(
			"Resource '" + next + "' doesn't exist inside: " + ro.URI)
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

	var h *handler

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

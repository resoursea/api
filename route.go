package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// This struct stores a tree of routed methods
// It implements the Router interface
// and implements the net/http.Handler interface
type route struct {
	// The Route URI
	// The name of the Resource in lowercase
	name string

	// The Resource value
	// that created this route
	value reflect.Value

	// Mapped Methods attached in this Route
	// Indexed by the route identifier in lowercase
	// ex: get, postlike
	methods map[string]*method

	// Children Route that builds a tree
	// Indexed by the child name
	children map[string]*route

	// True if this is a Route for a set of Resources
	isSlice bool
}

// It maps the Resource's mapped methods and creates a new Route tree
func newRoute(r *resource) (*route, error) {

	//log.Printf("Building Routes for %s\n", r)

	ro := &route{
		name:     r.name,
		value:    r.value,
		methods:  make(map[string]*method),
		children: make(map[string]*route),
		isSlice:  r.isSlice,
	}

	// Maps the Resource's mapped Methods
	// and also the Resources it extends recursively
	err := ro.scanRoutesFrom(r)
	if err != nil {
		return nil, err
	}

	// Check for Circular Dependency
	// on the Dependencies of each mapped Method
	err = checkCircularDependency(ro)
	if err != nil {
		return nil, err
	}

	// Go down to the Resource tree
	// and create Routes recursivelly for each Resource child
	for _, child := range r.children {
		c, err := newRoute(child)
		if err != nil {
			return nil, err
		}

		// Add this new routed child
		// and ensures that there is no URI conflict
		err = ro.addChild(c)
		if err != nil {
			return nil, err
		}

	}

	return ro, nil
}

// Scan the methods of some Resourse
// Resources are actually saved as a pointer to the Resource
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {} will not be visible to non pointer
func (ro *route) scanRoutesFrom(r *resource) error {

	err := ro.mapsMethods(r)
	if err != nil {
		return err
	}

	//
	// TODO
	//
	// This Resource Type already extends all the
	// extended methods, and don't need to map
	// this extended resource methods

	// All the resources it Exstends
	// should be mapped to this Route too
	/*
		for _, extend := range r.extends {
			err := ro.scanRoutesFrom(extend)
			if err != nil {
				return err
			}
		}
	*/

	return nil
}

// Maps the methods from one Resource type and attach it to the Route
func (ro *route) mapsMethods(r *resource) error {

	t := r.value.Type()

	//log.Println("### Scanning methods from type", t, "is slice:", isSliceType(t))

	for i := 0; i < t.NumMethod(); i++ {

		m := t.Method(i)

		// We will accept all methods that
		// has GET, POST, PUT, DELETE, HEAD
		// in the prefix of the method name
		if isMappedMethod(m) {

			m, err := newMethod(m, r)
			if err != nil {
				return err
			}

			//log.Printf("Adding Method %s for route %s\n", m, ro)

			// Check if this new Method will conflict with some address of Method that already exist
			// Action Handlers Names could conflict with Children Names...
			err = ro.addMethod(m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Return true if this route, or children/grandchildren...
// have mapped methods attached
func (ro *route) hasMethod() bool {
	if len(ro.methods) > 0 {
		return true
	}
	for _, child := range ro.children {
		if child.hasMethod() {
			return true
		}
	}
	return false
}

// Add a new Route child
func (ro *route) addChild(child *route) error {
	//log.Printf("addChild %s %v\n", child, child.hasMethod())

	// Add this Route to the tree only if it has methods
	if child.hasMethod() {

		// Test if this Name wasn't in use yet by one child
		_, exist := ro.children[child.name]
		if exist {
			return errors.New("Route " + ro.name + " already has child " + child.name)
		}

		// Test if this Name isn't used by one Method
		// Remember for Action Handlers
		for _, m := range ro.methods {
			_, addr := splitsMethodName(m)
			if addr == child.name {
				return fmt.Errorf("The address %s used by the resource %s"+
					" is already in use by an action in the route %s", addr, child, ro)
			}
		}

		ro.children[child.name] = child

		//log.Printf("Child name %s added %s\n", child.Name, child)
	}
	return nil
}

// Check if this new Method will conflict with some Method already created
// Action Handlers Names could conflict with Children Names...
func (ro *route) addMethod(m *method) error {

	_, address := splitsMethodName(m)

	// If this Method is an Action with address,
	// we should ensure that there is no other child with this address
	if len(address) > 0 {
		for addr, child := range ro.children {
			if addr == address {
				return fmt.Errorf("The address %s already used by the child %s in the route %s", addr, child, ro)
			}
		}
	}

	_, exist := ro.methods[strings.ToLower(m.method.Name)]
	if exist {
		return fmt.Errorf("%s already has method %s", ro, m)
	}

	// Index: GETLogin, POST, or POSTMessage...
	ro.methods[strings.ToLower(m.method.Name)] = m

	return nil
}

package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Route struct {
	// The resource name
	URI string

	// The main type of this Route
	Value reflect.Value

	// If this Route is for an Slice
	// Pointer to the Route of its Elem
	Elem *Route

	// Methods by its name,
	// the Action name plus the HTTP MEthod
	// Like: GETLogin our POST (for main methodos)
	Handlers map[string]*handler

	// map[URI]*Route
	Children map[string]*Route

	// True if the resource is an Slice of Resources
	IsSlice bool
}

// Receives the Root Resource and interate recursively
// creating the Route tree
func NewRoute(r *Resource) (*Route, error) {

	log.Printf("*Building Routes for %v\n", r)
	log.Println("Building Routes for", r.Value.Type())

	ro := &Route{
		URI:      r.Name,
		Value:    r.Value,
		Elem:     nil,
		Handlers: make(map[string]*handler),
		Children: make(map[string]*Route),
		IsSlice:  r.IsSlice,
	}

	// This Route take the methods of the main resource...
	ro.scanResource(r)

	// Check for Circular Dependency
	// on the Dependencies of each method
	//err := checkCircularDependency(ro)
	//if err != nil {
	//	return nil, err
	//}

	// If this Route is for an Slice
	// Map the Route for this Elem
	if r.IsSlice {
		e, err := NewRoute(r.Elem)
		if err != nil {
			return nil, err
		}

		ro.Elem = e
	}

	// Creating routes recursivelly for each resource child
	for _, child := range r.Children {
		c, err := NewRoute(child)
		if err != nil {
			return nil, err
		}

		err = ro.AddChild(c)
		if err != nil {
			return nil, err
		}

	}

	return ro, nil
}

// Scan the methods of some type
// We need to scan the methods of the Ptr to the Struct,
// cause some methods could be attached to the pointer,
// like func (r *Resource) GET() {}
func (ro *Route) scanResource(r *Resource) {

	ro.scanType(r)

	// ... and all the resource it Exstends will be mapped too
	for _, extend := range r.Extends {
		ro.scanResource(extend)
	}

}

// Scan the methods from one Type and add it to the Route
// This type could be []*Resource or just *Resource
func (ro *Route) scanType(r *Resource) {

	t := r.Value.Type()

	log.Println("Scanning methods from type", t, "is slice:", isSliceType(t))

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		log.Println("Testing:", m.Name)

		// We will accept all methods with
		// has GET, POST, PUT, DELETE, HEAD
		// at the end of the method
		if isMappedMethod(m) {

			log.Println("Passed")

			method := newMethod(m)

			// Check if this resource have any child to conflict
			// with this method
			//if methodConflict(r, method) {
			//	log.Fatal("Your method conflicts")
			//}

			h := newHandler(r, method)

			// Index: GETLogin or POST...
			ro.Handlers[h.Method.HTTPMethod+h.Method.Name] = h
		}
	}
}

// Return false if this Route have no methods declared
func (ro *Route) hasHandler() bool {

	if len(ro.Handlers) > 0 {
		return true
	}

	if ro.IsSlice {
		return ro.Elem.hasHandler()
	}

	for _, child := range ro.Children {
		if child.hasHandler() {
			return true
		}
	}

	return false
}

// Add a new Route child
func (ro *Route) AddChild(child *Route) error {
	// Add this Route to the tree only if it has Elemhandlers
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
func (ro *Route) handler(uri []string, httpMethod string, ids idMap) (*handler, error) {

	log.Println("Route Handling", uri, " in the ", ro.Value.Type())

	// Check if we should return
	// some Slice Handler of this Route
	if len(uri) == 0 {
		h, exist := ro.Handlers[httpMethod]
		if !exist {
			return nil, errors.New("No Method " + httpMethod + " in the " + ro.URI)
		}
		return h, nil
	}

	if len(uri) == 1 {
		// Check if is using some Action of this Resource
		h, exist := ro.Handlers[uri[0]+httpMethod]
		if exist { // Return the action
			return h, nil
		}

		log.Println("* action " + uri[0] + httpMethod + " NOT FOUND")
		//log.Println(ro.Handlers)
	}

	if ro.IsSlice {
		// Add this ID to the list
		ids[ro.Value.Type()] = reflect.ValueOf(ID(uri[0]))

		return ro.Elem.handler(uri[1:], httpMethod, ids)
	}

	child, exist := ro.Children[uri[0]]
	if exist {
		return child.handler(uri[1:], httpMethod, ids)
	}

	return nil, errors.New("Route " + ro.URI + " doesn't any Child or Action " + uri[0])
}

// To implement http.Handler
func (ro *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("### Serving the resource", req.URL.RequestURI())

	// Get the resource identfiers from the URL
	// Remember to descart the first empty element of the list
	uri := strings.Split(req.URL.RequestURI(), "/")[1:]

	//log.Printf("URI: %v\n", uri)

	// Check if this main Route matches with the requested URI
	if ro.URI != uri[0] {
		http.Error(w, "Route "+ro.URI+" not match with "+uri[0], http.StatusNotFound)
		return
	}

	// Store the IDs of the resources in the URI
	ids := idMap{}

	handler, err := ro.handler(uri[1:], req.Method, ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Route found: [%s] %s ids: %q\n",
		handler.Method.Name+handler.Method.HTTPMethod, req.URL.RequestURI(), ids)

	context := newContext(handler, w, req, ids)

	output := context.run()

	// If there is jsut one resource to send back
	if handler.Method.NumOut == 1 {

		// Compile our response in JSON
		jsonResponse, err := json.MarshalIndent(output[0].Interface(), "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	// If there is no resource, answare with empty list
	// if there is more than one, put them on list

	// Trans form the method output into an slice of the values
	// * Needed to generate a JSON response
	response := make(map[string]interface{}, handler.Method.NumOut)
	for i, v := range output {
		response[handler.Method.OutName[i]] = v.Interface()
	}

	// Compile our response in JSON
	jsonResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

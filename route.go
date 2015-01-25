package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Router interface {
	String() string
	Children() []Router
	Handlers() []Handler
	http.Handler
}

type Handler interface {
	String() string
}

type Route struct {
	// The resource name
	Name string

	// The main type of this Route
	Value reflect.Value

	// If this Route is for an Slice
	// Pointer to the Route of its Elem
	Elem *Route

	// Methods by its name,
	// the Action name plus the HTTP MEthod
	// Like: GETLogin our POST (for main methodos)
	handlers map[string]*handler

	// map[Name]*Route
	children map[string]*Route

	// True if the resource is an Slice of Resources
	IsSlice bool
}

// Creates a new Resource tree based on given Struct
// Receives the Struct to be mapped in a new Resource Tree,
// it also receive the Field name and Field tag as optional arguments
func NewRoute(object interface{}, args ...string) (*Route, error) {

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

	r, err := newResource(value, field, nil)
	if err != nil {
		return nil, err
	}

	return newRoute(r)
}

// Receives the Root Resource and interate recursively
// creating the Route tree
func newRoute(r *resource) (*Route, error) {

	//log.Printf("Building Routes for %s\n", r)

	ro := &Route{
		Name:     r.Name,
		Value:    r.Value,
		Elem:     nil,
		handlers: make(map[string]*handler),
		children: make(map[string]*Route),
		IsSlice:  r.IsSlice,
	}

	// This Route take the methods of the main resource
	// and all the resource it Exstends will be mapped too
	err := ro.scanRoutesFrom(r)
	if err != nil {
		return nil, err
	}

	// Check for Circular Dependency
	// on the Dependencies of each method
	err = checkCircularDependency(ro)
	if err != nil {
		return nil, err
	}

	// If this Route is for an Slice
	// Map the Route for this Elem
	if r.IsSlice {
		ro.Elem, err = newRoute(r.Elem)
		if err != nil {
			return nil, err
		}
	}

	// Creating routes recursivelly for each resource child
	for _, child := range r.Children {
		c, err := newRoute(child)
		if err != nil {
			return nil, err
		}

		//log.Printf("Adding child %s to parent %s\n", c, r)

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
func (ro *Route) scanRoutesFrom(r *resource) error {

	err := ro.scanMethods(r)
	if err != nil {
		return err
	}

	// All the resource it Exstends will be mapped too
	for _, extend := range r.Extends {
		err := ro.scanRoutesFrom(extend)
		if err != nil {
			return err
		}
	}

	return nil
}

// Scan the methods from one Type and add it to the Route
// This type could be []*Resource or just *Resource
func (ro *Route) scanMethods(r *resource) error {

	t := r.Value.Type()

	//log.Println("Scanning methods from type", t, "is slice:", isSliceType(t))

	for i := 0; i < t.NumMethod(); i++ {

		m := t.Method(i)

		//log.Println("Testing:", m.Name, isMappedMethod(m))

		// We will accept all methods that
		// has GET, POST, PUT, DELETE, HEAD
		// in the prefix of the method name
		if isMappedMethod(m) {

			h, err := newHandler(m, r)
			if err != nil {
				return err
			}

			//log.Printf("Adding Handler %s for route %s\n", h, ro)

			// Check if this new Handler will conflict with some address of Handler that already exist
			// Action Handlers Names could conflict with Children Names...
			err = ro.checkAddrConflict(h)
			if err != nil {
				return err
			}

			// Index: GETLogin, POST, or POSTMessage...
			ro.handlers[h.Method.HTTPMethod+h.Method.Name] = h
		}
	}

	return nil
}

// Return false if this Route have no methods declared
func (ro *Route) hasHandler() bool {

	if len(ro.handlers) > 0 {
		return true
	}

	if ro.IsSlice {
		return ro.Elem.hasHandler()
	}

	for _, child := range ro.children {
		if child.hasHandler() {
			return true
		}
	}

	return false
}

// Add a new Route child
func (ro *Route) AddChild(child *Route) error {
	//log.Printf("AddChild %s %v\n", child, child.hasHandler())

	// Add this Route to the tree only if it has handlers
	if child.hasHandler() {

		// Test if this Name wasn't in use yet by one child
		_, exist := ro.children[child.Name]
		if exist {
			return errors.New("Route " + ro.Name + " already has child " + child.Name)
		}

		// Test if this Name isn't used by one Handler
		// Remember for Action Handlers
		for _, h := range ro.handlers {
			//log.Printf("TESTING %s WITH %s\n", child, h)
			if child.Name == h.Method.Name {
				return fmt.Errorf("%s children of %s conflicts with %s", child, ro, h)
			}
		}

		ro.children[child.Name] = child

		//log.Printf("Child name %s added %s\n", child.Name, child)
	}
	return nil
}

// Check if this new Handler will conflict with some Handler already created
// Action Handlers Names could conflict with Children Names...
func (ro *Route) checkAddrConflict(h *handler) error {

	for name, child := range ro.children {
		if name == h.Method.Name {
			return fmt.Errorf("%s children of %s conflicts with %s", child, ro, h)
		}
	}

	_, exist := ro.handlers[h.Method.HTTPMethod+h.Method.Name]
	if exist {
		return fmt.Errorf("%s already has handler %s", ro, h)
	}

	return nil
}

// Return the Route from the especified Name
// Fulfill the ID Map with IDs present in the requested URI
func (ro *Route) handler(uri []string, httpMethod string, ids idMap) (*handler, error) {

	//log.Println("Route Handling", uri, "in the", ro)

	// Check if is trying to request some Handler of this Route
	if len(uri) == 0 {
		h, exist := ro.handlers[httpMethod]
		if !exist {
			return nil, fmt.Errorf("Method %s not found in the %s", httpMethod, ro)
		}
		return h, nil
	}

	// Check if is trying to request some Action Handler of this Route
	if len(uri) == 1 {

		h, exist := ro.handlers[httpMethod+uri[0]]
		if exist {
			return h, nil
		}

		// It is not an error, cause could have an resources with this name, not an action
		//log.Println("Action " + httpMethod + uri[0] + " NOT FOUND")
	}

	// If we are in a Slice Route, get its ID and search in the Elem Route
	if ro.IsSlice {
		// Add its ID to the Map
		id := &ID{id: uri[0]}
		ids[ro.Elem.Value.Type()] = reflect.ValueOf(id)

		return ro.Elem.handler(uri[1:], httpMethod, ids)
	}

	// If we are in an Elem Route, the only possibility is to have a Child with this Name
	child, exist := ro.children[uri[0]]
	if exist {
		return child.handler(uri[1:], httpMethod, ids)
	}

	return nil, fmt.Errorf("Not exist any Child '%s' or Action '%s' in the %s", uri[0], httpMethod+strings.Title(uri[0]), ro)
}

// Implementing the http.Handler Interface
// TODO: Error messages should be sent in JSON
func (ro *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//log.Println("### Serving the resource", req.URL.RequestURI())

	// Get the resource identfiers from the URL
	// Remember to descart the query string: ?q=sfxt&x=132...
	// Remember to descart the first empty element of the list
	uri := strings.Split(strings.Split(req.URL.RequestURI(), "?")[0], "/")[1:]

	// Check if the requested URI maches with this main Route
	if ro.Name != uri[0] {
		writeError(w, errors.New("Route "+ro.Name+" not match with "+uri[0]), http.StatusNotFound)
		return
	}

	// Store the IDs of the resources in the URI
	ids := idMap{}

	handler, err := ro.handler(uri[1:], req.Method, ids)
	if err != nil {
		writeError(w, err, http.StatusNotFound)
		return
	}

	//log.Printf("Route found: %s = %s ids: %q\n", req.URL.RequestURI(), handler, ids)

	// Process the request with the found Handler
	output := newContext(handler, w, req, ids).run()

	// If there is no output to sent back
	if handler.Method.NumOut == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If there is just one resource to send back
	if handler.Method.NumOut == 1 {

		// Encode the output in JSON
		jsonResponse, err := json.MarshalIndent(output[0].Interface(), "", "\t")
		if err != nil {
			writeError(w, errors.New("Error encoding to Json: "+err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	// If there is more than one output

	// Trans form the method output into an slice of the values
	// * Needed to generate a JSON response
	response := make(map[string]interface{}, handler.Method.NumOut)
	for i, v := range output {
		if !v.IsNil() {
			// Error is printing empty structs, treat that...
			if handler.Method.Outputs[i] == errorType {
				response[handler.Method.OutName[i]] = v.Interface().(error).Error()
				continue
			}
			if handler.Method.Outputs[i] == errorSliceType {

				errs := ""
				for _, err := range v.Interface().([]error) {
					errs += err.Error() + ". "
				}
				response[handler.Method.OutName[i]] = errs
				continue
			}

			response[handler.Method.OutName[i]] = v.Interface()
		}
	}

	// Encode the output in JSON
	jsonResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		writeError(w, errors.New("Error encoding to Json: "+err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func writeError(w http.ResponseWriter, err error, status int) {
	// Encode the output in JSON
	jsonResponse, err := json.MarshalIndent(map[string]string{"error": err.Error()}, "", "\t")
	if err != nil {
		http.Error(w, "{error: \"Error encoding the error message to Json: "+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (ro *Route) String() string {
	str := fmt.Sprintf("Route: [%s] %s:", ro.Name, ro.Value.Type())
	for _, h := range ro.handlers {
		str += fmt.Sprintf(" ([%s] /%s)", h.Method.HTTPMethod, h.Method.Name)
	}
	return str
}

func (ro *Route) Children() []Router {
	children := make([]Router, len(ro.children))
	i := 0
	for _, c := range ro.children {
		children[i] = c
		i += 1
	}
	return children
}

func (ro *Route) Handlers() []Handler {
	handlers := make([]Handler, len(ro.handlers))
	i := 0
	for _, h := range ro.handlers {
		handlers[i] = h
		i += 1
	}
	return handlers
}

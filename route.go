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
	Methods() []Method
	IsSlice() bool
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type Method interface {
	String() string
}

type route struct {
	// The resource name
	name string

	// The main type of this Route
	value reflect.Value

	// Methods by its name,
	// the Action name plus the HTTP MEthod
	// Like: GETLogin our POST (for main methodos)
	methods map[string]*method

	// map[Name]*Route
	children map[string]*route

	// True if the resource is an Slice of Resources
	isSlice bool
}

// Creates a new Resource tree based on given Struct
// Receives the Struct to be mapped in a new Resource Tree,
// it also receive the Field name and Field tag as optional arguments
func NewRoute(object interface{}, args ...string) (*route, error) {

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
func newRoute(r *resource) (*route, error) {

	//log.Printf("Building Routes for %s\n", r)

	ro := &route{
		name:     r.name,
		value:    r.value,
		methods:  make(map[string]*method),
		children: make(map[string]*route),
		isSlice:  r.isSlice,
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

	// Creating routes recursivelly for each resource child
	for _, child := range r.children {
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
func (ro *route) scanRoutesFrom(r *resource) error {

	err := ro.scanMethods(r)
	if err != nil {
		return err
	}

	// All the resource it Exstends will be mapped too
	for _, extend := range r.extends {
		err := ro.scanRoutesFrom(extend)
		if err != nil {
			return err
		}
	}

	return nil
}

// Scan the methods from one Type and add it to the Route
// This type could be []*Resource or just *Resource
func (ro *route) scanMethods(r *resource) error {

	t := r.value.Type()

	//log.Println("Scanning methods from type", t, "is slice:", isSliceType(t))

	for i := 0; i < t.NumMethod(); i++ {

		m := t.Method(i)

		//log.Println("Testing:", m.Name, isMappedMethod(m))

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
			err = ro.checkAddrConflict(m)
			if err != nil {
				return err
			}

			// Index: GETLogin, POST, or POSTMessage...
			ro.methods[strings.ToLower(m.method.Name)] = m
		}
	}

	return nil
}

// Return false if this Route have no methods declared
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
func (ro *route) AddChild(child *route) error {
	//log.Printf("AddChild %s %v\n", child, child.hasMethod())

	// Add this Route to the tree only if it has methods
	if child.hasMethod() {

		// Test if this Name wasn't in use yet by one child
		_, exist := ro.children[child.name]
		if exist {
			return errors.New("Route " + ro.name + " already has child " + child.name)
		}

		// Test if this Name isn't used by one Method
		// Remember for Action Handlers
		for _, h := range ro.methods {
			//log.Printf("TESTING %s WITH %s\n", child, h)
			if child.name == h.method.Name {
				return fmt.Errorf("%s children of %s conflicts with %s", child, ro, h)
			}
		}

		ro.children[child.name] = child

		//log.Printf("Child name %s added %s\n", child.Name, child)
	}
	return nil
}

// Check if this new Method will conflict with some Method already created
// Action Handlers Names could conflict with Children Names...
func (ro *route) checkAddrConflict(m *method) error {

	for name, child := range ro.children {
		if name == m.method.Name {
			return fmt.Errorf("%s children of %s conflicts with %s", child, ro, m)
		}
	}

	_, exist := ro.methods[m.method.Name]
	if exist {
		return fmt.Errorf("%s already has method %s", ro, m)
	}

	return nil
}

// Return the Route from the especified Name
// Fulfill the ID Map with IDs present in the requested URI
func (ro *route) method(uri []string, httpMethod string, ids idMap) (*method, error) {

	//log.Println("Route Handling", uri, "in the", ro)

	// Check if is trying to request some Method of this Route
	if len(uri) == 0 {
		m, exist := ro.methods[httpMethod]
		if !exist {
			return nil, fmt.Errorf("Method %s not found in the %s", httpMethod, ro)
		}
		return m, nil
	}

	// Check if is trying to request some Action Method of this Route
	if len(uri) == 1 {

		m, exist := ro.methods[httpMethod+uri[0]]
		if exist {
			return m, nil
		}

		// It is not an error, cause could have an resources with this name, not an action
		//log.Println("Action " + httpMethod + uri[0] + " NOT FOUND")
	}

	// If we are in a Slice Route, get its ID and search in the Elem Route
	if ro.isSlice {
		// Get the only child this route has
		for _, child := range ro.children {
			// Add its ID to the Map
			ids[child.value.Type()] = reflect.ValueOf(&id{id: uri[0]})
			// Continue searching in the Route child
			return child.method(uri[1:], httpMethod, ids)
		}
		// It should never occours
		return nil, fmt.Errorf("Route %s is an slice and has no child!")
	}

	// If we are in an Elem Route, the only possibility is to have a Child with this Name
	child, exist := ro.children[uri[0]]
	if exist {
		return child.method(uri[1:], httpMethod, ids)
	}

	return nil, fmt.Errorf("Not exist any Child '%s' or Action '%s' in the %s", uri[0], httpMethod+strings.Title(uri[0]), ro)
}

// Implementing the http.Handler Interface
// TODO: Error messages should be sent in JSON
func (ro *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//log.Println("### Serving the resource", req.URL.RequestURI())

	// Get the resource identfiers from the URL
	// Remember to descart the query string: ?q=sfxt&x=132...
	// Remember to descart the first empty element of the list
	uri := strings.Split(strings.Split(req.URL.RequestURI(), "?")[0], "/")[1:]

	// Check if the requested URI maches with this main Route
	if ro.name != uri[0] {
		writeError(w, errors.New("Route "+ro.name+" not match with "+uri[0]), http.StatusNotFound)
		return
	}

	// Store the IDs of the resources in the URI
	ids := idMap{}
	httpMethod := strings.ToLower(req.Method)

	method, err := ro.method(uri[1:], httpMethod, ids)
	if err != nil {
		writeError(w, err, http.StatusNotFound)
		return
	}

	//log.Printf("Route found: %s = %s ids: %q\n", req.URL.RequestURI(), method, ids)

	// Process the request with the found Method
	output := newContext(method, w, req, ids).run()

	// If there is no output to sent back
	if method.method.Type.NumOut() == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If there is just one resource to send back
	if method.method.Type.NumOut() == 1 {

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
	response := make(map[string]interface{}, method.method.Type.NumOut())
	for i, v := range output {
		if !v.IsNil() {
			// Error is printing empty structs, treat that...
			if v.Type() == errorType {
				response["error"] = v.Interface().(error).Error()
				continue
			}
			if v.Type() == errorSliceType {

				errs := make([]string, v.Len())
				for i := 0; i < v.Len(); i++ {
					errs[i] = v.Index(i).Interface().(error).Error()
				}
				response["errors"] = errs
				continue
			}

			response[method.outName[i]] = v.Interface()
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

func (ro *route) String() string {
	return fmt.Sprintf("[%s] %s", ro.name, ro.value.Type())
}

func (ro *route) Children() []Router {
	children := make([]Router, len(ro.children))
	i := 0
	for _, c := range ro.children {
		children[i] = c
		i += 1
	}
	return children
}

func (ro *route) Methods() []Method {
	methods := make([]Method, len(ro.methods))
	i := 0
	for _, m := range ro.methods {
		methods[i] = m
		i += 1
	}
	return methods
}

func (ro *route) IsSlice() bool {
	return ro.isSlice
}

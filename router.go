package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// This is the main interface returned to user
// It routes the Resources and Methods
// and can be used with the Go's net/http standard library
type Router interface {
	ServeHTTP(http.ResponseWriter, *http.Request)

	Methods() []Method
	Children() []Router
	IsSlice() bool
	String() string
}

type router struct {
	route
}

// This interface is returned when
// user asks for an Route Method
type Method interface {
	String() string
}

// Creates a new Resource tree based on given Struct
// Receives the Struct to be mapped in a new Resource Tree,
// it also receive the Field name and Field tag as optional arguments
func NewRouter(object interface{}, args ...string) (*route, error) {

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

///////////////////////////////////////////////////
//  Router methods attached to the Route struct  //
///////////////////////////////////////////////////

//
// This methods are attached to the Route struct
// to garants it implments the Router interface
// These are saved here to reduce de size of the route.go file
//

// Return the the method pointed by the URI and httpMethod
// Fulfill the IDMap with IDs present in the requested URI
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
	}

	// If we are in a Slice Route, get its ID and search in the Child Route
	if ro.isSlice {
		// Get the only child this route has, the slice Element
		for _, child := range ro.children {
			// Add its ID to the Map
			ids[child.value.Type()] = reflect.ValueOf(&id{id: uri[0]})
			// Continue searching in the Route child
			return child.method(uri[1:], httpMethod, ids)
		}
		// It should never occurs, because a slice Resource always has a Elem Resource
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
	// Remember to descart the first empty element of the list, before the first /
	uri := strings.Split(strings.Split(req.URL.RequestURI(), "?")[0], "/")[1:]

	// Check if the requested URI maches with this main Route
	if ro.name != uri[0] {
		writeError(w, errors.New("Route "+ro.name+" not match with "+uri[0]), http.StatusNotFound)
		return
	}

	// Store the IDs of the resources in the URI
	ids := idMap{}
	httpMethod := strings.ToLower(req.Method)

	// Get the method this URI and HTTP method is pointing to
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

// Return all accessible Methods in a specific Route
func (ro *route) Methods() []Method {
	methods := make([]Method, len(ro.methods))
	i := 0
	for _, m := range ro.methods {
		methods[i] = m
		i += 1
	}
	return methods
}

// Return all children Routes of a specific Route
func (ro *route) Children() []Router {
	children := make([]Router, len(ro.children))
	i := 0
	for _, c := range ro.children {
		children[i] = c
		i += 1
	}
	return children
}

// Return true if this Route wraps a list of Resources
func (ro *route) IsSlice() bool {
	return ro.isSlice
}

// Return a text with the name and the type of a specific Route
func (ro *route) String() string {
	return fmt.Sprintf("[%s] %s", ro.name, ro.value.Type())
}

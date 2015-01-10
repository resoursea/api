package resource

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	// The Root Route
	Route *Route
}

// Creates a new server with a new root Route
func NewServer() *Server {
	return &Server{
		Route: &Route{Children: make(map[string]*Route)},
	}
}

// Add a new resource to the root Route
func (s *Server) Add(resource *Resource) error {
	err := s.Route.addChild(NewRoute(resource))
	if err != nil {
		return err
	}
	return nil
}

// To implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Serving the resource", req.URL.RequestURI())

	// Get the resource identfiers from the URL
	// Remember to descart the first empty element of the list
	uri := strings.Split(req.URL.RequestURI(), "/")[1:]

	log.Printf("URI: %v\n", uri)

	handler, ids, err := s.Route.getHandler(uri, req.Method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Route [%s] %s got, ids: %q\n", handler.Name, req.URL.RequestURI(), ids)

	context := newContext(handler, w, req, ids)

	output := context.run()

	// Trans form the method output into an slice of the values
	// * Needed to generate a JSON response
	response := make(map[string]interface{}, handler.Method.NumOut)
	for i, v := range output {
		response[handler.Method.OutName[i]] = v.Interface()
	}

	// Compile our response in JSON
	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

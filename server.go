package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	// The Root Route
	Route *route
}

// Creates a new server with a new root Route
func NewServer() *Server {
	return &Server{
		Route: &route{Children: make(map[string]*route)},
	}
}

// Add a new resource to the root Route
func (s *Server) Add(r *resource) error {
	err := s.Route.addChild(NewRoute(r))
	if err != nil {
		return err
	}
	return nil
}

// To implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("### Serving the resource", req.URL.RequestURI())

	// Get the resource identfiers from the URL
	// Remember to descart the first empty element of the list
	uri := strings.Split(req.URL.RequestURI(), "/")[1:]

	//log.Printf("URI: %v\n", uri)

	handler, ids, err := s.Route.getHandler(uri, req.Method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Route found: [%s] %s ids: %q\n", handler.Name, req.URL.RequestURI(), ids)

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

func (s *Server) Run(addr string) {
	log.Println("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, s))
}

package resource

import (
	"net/http"
)

type Server struct {
	// The Root Route
	Route *Route
}

// Receives the Root Resource and creates a new server
func NewServer(resource *Resource) *Server {
	return &Server{
		Route: NewRoute(resource),
	}
}

// To implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

}

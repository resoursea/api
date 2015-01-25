package api

import (
	"fmt"
	"reflect"
)

type handler struct {
	Method *method
	// Many Types could point to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies dependencies
}

func newHandler(m reflect.Method, r *resource) (*handler, error) {

	//log.Println("Creating Handler for method", m.Name, m.Type)

	ds, err := newDependencies(m, r)
	if err != nil {
		return nil, err
	}

	met := newMethod(m)

	h := &handler{
		Method:       met,
		Dependencies: ds,
	}

	return h, nil
}

func (h *handler) String() string {
	return fmt.Sprintf("Handler: [%s%s] %s", h.Method.HTTPMethod, h.Method.Name, h.Method.Method.Type)
}

package api

import (
	"fmt"
	"reflect"
)

type method struct {
	Method reflect.Method
	// Many Types could point to the same dependencie
	// It could occour couse could have any number of Interfaces
	// that could be satisfied by a single dependency
	Dependencies dependencies
	OutName      []string
}

func newMethod(m reflect.Method, r *resource) (*method, error) {

	//log.Println("Creating Method", m.Name, m.Type)

	ds, err := newDependencies(m, r)
	if err != nil {
		return nil, err
	}

	h := &method{
		Method:       m,
		Dependencies: ds,
		OutName:      make([]string, m.Type.NumOut()),
	}

	// Caching the Output Resources name
	for i := 0; i < m.Type.NumOut(); i++ {
		h.OutName[i] = elemOfType(m.Type.Out(i)).Name()
	}

	return h, nil
}

func (h *method) String() string {
	return fmt.Sprintf("[%s] %s", h.Method.Name, h.Method.Type)
}

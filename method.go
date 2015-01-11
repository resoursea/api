package api

import (
	"reflect"
	"strings"
)

type method struct {
	Name    string
	Method  reflect.Method
	Owner   reflect.Type // The Struct Type that contains this Method
	NumIn   int
	Inputs  []reflect.Type
	NumOut  int
	OutName []string
}

func newMethod(m reflect.Method) *method {

	//log.Println("Creating method", method)

	met := &method{
		Name:    m.Name,
		Method:  m,
		Owner:   m.Type.In(0), // The first parameter will always be the Struct itself
		NumIn:   m.Type.NumIn(),
		Inputs:  make([]reflect.Type, m.Type.NumIn()),
		NumOut:  m.Type.NumOut(),
		OutName: make([]string, m.Type.NumOut()),
	}

	// Store the input Types in a slice
	for i := 0; i < met.NumIn; i++ {
		met.Inputs[i] = m.Type.In(i)
	}

	// Save the output Types name to use in the response
	for i := 0; i < m.Type.NumOut(); i++ {

		// Gets the type of the output
		t := m.Type.Out(i)
		// If it is an Slice, get the type of the element it carries
		t = mainElemOfType(t)

		met.OutName[i] = strings.ToLower(t.Name())
	}

	return met
}

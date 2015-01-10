package resource

import (
	"reflect"
	"strings"
)

type Method struct {
	Name    string
	Method  reflect.Method
	Owner   reflect.Type // The Struct Type that contains this Method
	NumIn   int
	Inputs  []reflect.Type
	NumOut  int
	OutName []string
}

func NewMethod(method reflect.Method) *Method {

	//log.Println("Creating method", method)

	m := &Method{
		Name:    method.Name,
		Method:  method,
		Owner:   method.Type.In(0), // The first parameter will always be the Struct itself
		NumIn:   method.Type.NumIn(),
		Inputs:  make([]reflect.Type, method.Type.NumIn()),
		NumOut:  method.Type.NumOut(),
		OutName: make([]string, method.Type.NumOut()),
	}

	// Store the input Types in a slice
	for i := 0; i < m.NumIn; i++ {
		m.Inputs[i] = method.Type.In(i)
	}

	// Save the output Types name to use in the response
	for i := 0; i < method.Type.NumOut(); i++ {

		// Gets the type of the output
		t := method.Type.Out(i)
		// If it is an Slice, get the type of the element it carries
		t = mainElemOfType(t)

		m.OutName[i] = strings.ToLower(t.Name())
	}

	return m
}

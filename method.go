package api

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type method struct {
	// For Resource Method could be:
	// Main: GET, PUT...
	// Actions: GETLogin, GETPic, DELETELogin
	Name       string
	HTTPMethod string
	Method     reflect.Method
	Owner      reflect.Type // The Struct Type that contains this Method
	NumIn      int
	Inputs     []reflect.Type
	NumOut     int
	OutName    []string
}

var httpMethods = [...]string{
	"GET",
	"PUT",
	"POST",
	"DELETE",
	"HEAD",
}

func newMethod(m reflect.Method) *method {

	// Extract the method Action and HTTP Method
	// Ex: GETLogin
	// Name: login, HTTP Method: GET
	// Ex: POST
	// Name: 'empty', HTTP Method: POST
	httpMethod, name := decodeMethodName(m)

	met := &method{
		Name:       name,
		HTTPMethod: httpMethod,
		Method:     m,
		Owner:      m.Type.In(0), // It is the Struct Type that has the method
		NumIn:      m.Type.NumIn(),
		Inputs:     make([]reflect.Type, m.Type.NumIn()),
		NumOut:     m.Type.NumOut(),
		OutName:    make([]string, m.Type.NumOut()),
	}

	log.Println("New Method:", met)

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

// Methods could be Main methods and Action methods
// Main methods respond to directly to the:
// GET, PUT, POST, DELETE, HEAD of the resources
// Action methods respond for some action of the resource,
// ex: LoginGET, respond to: [GET] resource/login
func decodeMethodName(m reflect.Method) (httpMethod string, name string) {

	for _, httpMethod := range httpMethods {
		if strings.HasPrefix(m.Name, httpMethod) {
			name = strings.TrimPrefix(m.Name, httpMethod)
			return httpMethod, strings.ToLower(name)
		}
	}

	return "", ""
}

// Return if this method should be mapped or not
// Methods starting with GET, POST, PUT, DELETE or HEAD should be mapped
func isMappedMethod(m reflect.Method) bool {

	for _, httpMethod := range httpMethods {
		if strings.HasPrefix(m.Name, httpMethod) {
			return true
		}
	}

	return false
}

func (m *method) String() string {
	return fmt.Sprintf("[%s%s] %s", m.HTTPMethod, m.Name, m.Method.Type)
}

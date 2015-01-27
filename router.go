package api

import (
	"net/http"
	"reflect"
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

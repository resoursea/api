package resource

import (
	"log"
	"testing"
)

//
// Test
//
func TestResource(t *testing.T) {
	a := A{
		Name: "Testing",
		X:    X{Test: "Tested"},
	}

	resource := NewResource(a, "recurso")

	//NewServer(resource)

	//server.BuildRoutes()

	log.Println("----------")

	printResource(resource, 0)

}

//
// Resources used in test
//

type A struct {
	Id   string
	Name string
	Bs   BList "Tag de Bs"
	X
	//C C // conflit
}

type BList []B

type B struct {
	Id   int
	Name string
	C    C "Tag de C"
}

type C struct {
	Nothing string
}

type X struct {
	Test string
	Gogo int
	C    C
}

func (a *A) Init() *A {
	return &A{}
}

func (a *A) GET(x A, y A, z InterfaceA) *A {
	return a
}

func (a *A) doSomething() {}

type InterfaceA interface {
	doSomething()
}

// Set the value passed by user on creation
//ptrValue.Elem().Set(value)

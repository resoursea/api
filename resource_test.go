package resource

import (
	"fmt"
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

	server := NewServer(resource)

	//server.BuildRoutes()

	fmt.Println("----------")

	printResource(resource, 0)

	fmt.Println("----------")

	printRoute(server.Route, 0)

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

func (b *BList) Init(c *C) {}

func (b *BList) POST() {}

type B struct {
	Id   int
	Name string
	C    C "Tag de C"
}

func (b *B) GET() {}

type C struct {
	Nothing string
}

func (c *C) Init(d D) {}

func (c *C) PUT(d *D) {}

type X struct {
	Test string
	Gogo int
	C    C
}

func (b *X) PUT() {}

func (a *A) Init(b BList) *A {
	return &A{}
}

func (a *A) GET(x A, y A, z InterfaceA, i *BList) *A {
	return a
}

func (a *A) doSomething() {}

type InterfaceA interface {
	doSomething()
}

type D struct {
	Test string
}

func (d *D) Init() {}

// Set the value passed by user on creation
//ptrValue.Elem().Set(value)

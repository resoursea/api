package resource

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

//
// Test
//
func TestResource(t *testing.T) {
	a := A{
		Name: "Testing",
		X:    X{Test: "Tested"},
		Bs:   BList{B{Name: "Started"}},
	}

	resource := NewResource(a, "recurso")

	server := NewServer()

	err := server.Add(resource)
	if err != nil {
		log.Println(err)
	}

	//server.BuildRoutes()

	fmt.Println("\n1 ----------\n")

	printResource(resource, 0)

	fmt.Println("\n2 ----------\n")

	printRoute(server.Route, 0)

	fmt.Println("\n3 ----------\n")

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/recurso/bs/", nil)

	server.ServeHTTP(res, req)

	fmt.Printf("RETURN: %v\n", res.Body)

	fmt.Println("\nEND --------\n")

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

func (b *BList) Init(c *C, d D) {
	log.Println("*** BList Received", c)
}

func (b *BList) GET() *BList {
	return b
}

type B struct {
	Id   int
	Name string
	C    C "Tag de C"
}

func (b *B) PUT() {}

type C struct {
	Nothing string
}

func (c *C) Init(d D) {
	c.Nothing = "Initialized ok"
	log.Println("*** C Received", d)
}

//func (c *C) PUT(d *D) {}

type X struct {
	Test string
	Gogo int
	C    C
}

//func (b *X) PUT() {}

func (a *A) Init(b BList) *A {
	return &A{}
}

func (a *A) GET(x A, i *BList) *A {
	return a
}

func (a *A) doSomething() {}

type InterfaceA interface {
	doSomething()
}

type D struct {
	Test string
}

func (d *D) Init() {
	d.Test = "TESTED"
}

// Set the value passed by user on creation
//ptrValue.Elem().Set(value)

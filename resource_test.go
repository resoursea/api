package api

import (
	"fmt"
	"log"
	//"net/http"
	//"net/http/httptest"
	"testing"
)

type API struct {
	A A
}

//
// Test
//
func TestResource(t *testing.T) {
	api := API{
		A{
			Name: "Testing",
			//X:    X{Test: "Tested"},
			//Bs: BList{B{Name: "Started"}},
			//B: &B{Name: "Setted"},
		},
	}

	resource, err := NewResource(api)
	if err != nil {
		t.Fatal(err)
	}

	PrintResource(resource)

	route, err := NewRoute(resource)
	if err != nil {
		t.Fatal(err)
	}

	PrintRoute(route)

	/*
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/a/bs/login", nil)

		route.ServeHTTP(res, req)

		fmt.Printf("RETURN: %v\n", res.Body)

	*/

	fmt.Println("\n-------- TEST END --------\n")

}

//
// Resources used in test
//

type A struct {
	Id   string
	Name string
	Bs   *BList "Tag de Bs"
	//X
	//C C // conflit, name 'c' already used
}

func (a A) Init(s A, b BList) *A {
	log.Println("Init A received", b)
	b[0].Name = "Changed"
	return &a
}

func (a *A) GET(x A, b *BList) *A {
	log.Println("GET A received", b)
	return a
}

func (a *A) PUTBsx() *A {
	log.Println("Conflict Handler")
	return a
}

type BList []B

func (bs *BList) Init() *BList {
	//log.Println("*** BList Received", d)
	return &BList{B{Name: "FUCKED"}}
}

func (b *BList) GET() *BList {
	return b
}

func (b BList) GETLogin() BList {
	b[0].Name = "ACTION"
	return b
}

type B struct {
	Id   int
	Name string
	//d    D
	Cs CList "Tag de C"
}

func (b *B) Init(id ID) *B {
	log.Println("*** B ID Received", id)
	b.Name = id.String()
	b.Id, _ = id.Int()
	return b
}
func (b *B) GET(id ID) *B {
	return b
}

func (b *B) PUT() {}

type C struct {
	Id      int
	BId     int
	Nothing string
}

func (c *C) Init(id ID, b B, a A) {
	log.Println("*** C Received ID:", id)
	c.Id, _ = id.Int()
	c.BId = b.Id
}

func (c *C) GET() *C {
	return c
}

type CList []C

func (c *CList) Init() {
	//log.Println("*** C Received", d)
}

//func (c *C) PUT(d *D) {}

type X struct {
	Test string
	Gogo int
	C    C
}

//func (b *X) PUT() {}

type DoSomething interface {
	doSomething()
}

type D struct {
	Test string
}

func (d *D) Init() {
	d.Test = "TESTED"
}

func (d *D) doSomething() {}

// Set the value passed by user on creation
//ptrValue.Elem().Set(value)

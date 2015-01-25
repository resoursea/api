package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type API struct {
	a  A
	A  A
	A2 *A
}

//
// Test
//
func TestResource(t *testing.T) {
	api := API{
		A: A{
			Name: "Testing",
			//X:    X{Test: "Tested"},
			Bs: &BList{B{Name: "Started"}},
			//B: &B{Name: "Setted"},
		},
	}

	rt, err := NewRoute(api)
	if err != nil {
		t.Fatal(err)
	}

	PrintRouter(rt)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/a/bs/123", nil)

	rt.ServeHTTP(res, req)

	fmt.Printf("RETURN: %v\n", res.Body)

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

func (a A) Init(ss A) *A {
	return &a
}

func (a *A) GET(b B, err error) *A {
	log.Println("GET A received", err)
	return a
}

func (a *A) PUTBsx() *A {
	log.Println("Conflict Handler")
	return a
}

type BList []B

func (bs *BList) Init() *BList {
	//log.Println("*** BList Received", d)
	bs2 := append(*bs, B{Name: "FUCKED"})
	return &bs2
}

func (b *BList) GET() *BList {
	return b
}

func (b BList) GETLogin() BList {
	b[0].Name += " ACTION"
	return b
}

type B struct {
	Id   int
	Name string
	//d    D
	Cs CList "Tag de C"
}

func (b *B) Init(id *ID) (B, error) {
	if id != nil {
		log.Println("*** B ID Received", id.String())
	} else {
		log.Println("*** B ID NOT RECEIVED")
	}
	//b.Name = id.String()
	b.Id = 312312
	b.Name = "Altered!"
	return *b, nil //fmt.Errorf("ERROR! LOL!")
}
func (b *B) GET(id *ID) *B {
	//b.Name = "B get" + id.String()
	b.Id, _ = id.Int()
	return b
}

func (b *B) PUT() {}

type C struct {
	Id      int
	BId     int
	Nothing string
}

func (c *C) Init(id *ID, b B, a A) {
	log.Println("*** C Received ID:", id)
	c.Id, _ = id.Int()
	c.BId = b.Id
}

func (c *C) GET() *C {
	return c
}

type CList []*C

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

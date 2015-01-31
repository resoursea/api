package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// The initial states of my Resources
var api = API{
	Gophers: Gophers{
		Gopher{
			Id:      1,
			Message: "I love you",
		},
		Gopher{
			Id:      2,
			Message: "I still love programming",
		},
		Gopher{
			Id:      3,
			Message: "You so cute",
		},
	},
}

func TestGetResource(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/gophers/1", nil)
	if err != nil {
		t.Error(err)
	}

	rt.ServeHTTP(w, req)

	// Try to get the gopher from the response
	var resp GopherData
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(resp.Gopher, api.Gophers[0]) {
		t.Error("The service returned the gopher wrong!")
	}

}

func TestGetAction(t *testing.T) {

	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/gophers/2/message", nil)
	if err != nil {
		t.Error(err)
	}

	rt.ServeHTTP(w, req)

	// Try to get the string returned
	var resp StringData
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.String != "I still love programming" {
		t.Error("The service returned something wrong!")
	}

}

type API struct {
	Gophers Gophers
}

type Gophers []Gopher

type Gopher struct {
	Id      int
	Message string
}

// A constructor for Gopher dependency
// Receives a Gophers dependency, and an ID passed on the URI
// Gophers has no constructor, then is injected the raw initial state for Gophers
func (_ *Gopher) New(gs Gophers, id ID) (*Gopher, error) {
	// Getting the ID in the URI
	i, err := id.Int()
	if err != nil {
		return nil, err
	}

	gopher := &Gopher{}
	for _, g := range gs {
		if g.Id == i {
			*gopher = g
		}
	}
	if gopher == nil {
		return nil, fmt.Errorf("Id %d not found in Gophers list", i)
	}
	return gopher, nil
}

func (g *Gopher) GET(err error) (*Gopher, error) {
	return g, err
}
func (g *Gopher) GETMessage(err error) (string, error) {
	if err != nil {
		return "", nil
	}
	return g.Message, err
}

type GopherData struct {
	Gopher Gopher
}

type StringData struct {
	String string
}

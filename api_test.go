// This package implements an API structure
// used to test many concepts of this framework
package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

// The initial states of my Resources
var api = API{
	Version: Version{
		Version: 1,
		Message: "API version: ",
	},
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
	Maltese: Maltese{}, // My Dogger implementation

}

type API struct {
	Gophers Gophers
	Version Version
	Maltese Maltese
	Date    Date
}

func (a *API) GETDogBark(dog Doger) string {
	return dog.Bark()
}

type Date struct {
	Time time.Time
}

func (d *Date) New() *Date {
	d.Time = time.Now().Local()
	return d
}

func (d *Date) GET() *Date {
	return d
}

type Version struct {
	Version int
	Message string
}

func (v *Version) Init() (*Version, error) {
	if v.Version != 1 {
		return nil, fmt.Errorf("Initial value of version not received")
	}
	v.Message = fmt.Sprintf("%s %d", v.Message, v.Version)
	return v, nil
}

func (v *Version) GET() *Version {
	return v
}

type Gophers []Gopher

// Testing the Init method, returning the new value to be used
func (gs Gophers) Init() (Gophers, error) {
	if len(gs) != len(api.Gophers) {
		return nil, fmt.Errorf("Gophers Init received a different initial value")
	}
	gs = append(gs, Gopher{
		Id:      4,
		Message: "Intruder",
	})
	return gs, nil
}

func (gs *Gophers) GET(err error) (*Gophers, error) {
	return gs, err
}

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
		return "", err
	}
	return g.Message, err
}

type Doger interface {
	Bark() string
}

type Maltese struct{}

func (m *Maltese) Bark() string {
	return "yap-yap"
}

//
// Reponse data
//

type ErrorResp struct {
	Error *string
}

type StringResp struct {
	String string
}

type VersionResp struct {
	Version Version
	Error   string
}

type DateResp struct {
	Date Date
}

type GopherResp struct {
	Gopher Gopher
}

type GophersResp struct {
	Gophers Gophers
}

// Try to get any Error from the response
// If it returned an error, it fails the Test
func errorTest(w *httptest.ResponseRecorder, t *testing.T) {
	var errResp ErrorResp
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	if err != nil {
		t.Fatal(errResp)
	}
	if errResp.Error != nil {
		t.Fatal("Error returned!")
	}
}

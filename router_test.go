package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGopherGet(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/gophers/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	// Test if returned some error
	errorTest(w, t)

	// Try to get the gopher from the response
	var resp GopherResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp.Gopher, api.Gophers[0]) {
		t.Fatal("The service returned the gopher wrong!")
	}

}

func TestGophersInit(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/gophers", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	// Test if returned some error
	errorTest(w, t)

	// Try to get the gopher from the response
	var resp GophersResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Gophers) < 4 {
		t.Fatal("The service not all gophers wrong!")
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
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	// Test if returned some error
	errorTest(w, t)

	// Try to get the string returned
	var resp StringResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.String != "I still love programming" {
		t.Fatal("The service returned something wrong!")
	}

}

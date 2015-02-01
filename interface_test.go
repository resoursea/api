// This package tests the Interface injection
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testing the Interface Injection
// The API resource has an Action called GETDogBark
// It depends on an Interface Dogger
// The interface is implemented by the Maltese struct attached to the API
func TestInterfaceInjection(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/dogbark", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	// Try to get the gopher from the response
	var resp StringResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	maltese := Maltese{}

	// Testing if the interface implementation injection
	if resp.String != maltese.Bark() {
		t.Fatal("Interface not injected correctly")
	}

}

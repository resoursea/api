package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionInit(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	// Try to get the gopher from the response
	var resp VersionResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	// Testing if the initial value of version was set
	if resp.Version.Version != api.Version.Version {
		t.Fatal("The initial value of Version wasn't set")
	}

	// Testing if Version was Initialized
	if resp.Version.Message != fmt.Sprintf("%s %d", api.Version.Message, api.Version.Version) {
		t.Fatal("Version wasn't initialized correctly")
	}
}

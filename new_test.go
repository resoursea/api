// This package tests the Resource Creation method New()
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testing the Resource Creation method New()
func TestNew(t *testing.T) {
	rt, err := NewRouter(api)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/date", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.ServeHTTP(w, req)

	var resp DateResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Date.Time.IsZero() {
		t.Fatal("Date not Created correctly")
	}
}

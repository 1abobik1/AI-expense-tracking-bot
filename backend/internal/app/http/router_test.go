package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	rt := NewRouter()
	rt.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("get " + r.PathValue("id"))) })
	rt.HandleFunc("DELETE /api/items/{id}", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("del " + r.PathValue("id"))) })
	rt.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	cases := []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", "/api/items/7", 200, "get 7"},
		{"DELETE", "/api/items/7", 200, "del 7"},
		{"GET", "/healthz", 200, "ok"},
		{"POST", "/api/items/7", 405, ""},
		{"GET", "/api/items", 404, ""},
		{"GET", "/api/items/7/x", 404, ""},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		rt.ServeHTTP(w, httptest.NewRequest(c.method, c.path, http.NoBody))

		if w.Code != c.status || (c.body != "" && w.Body.String() != c.body) {
			t.Errorf("%s %s = %d %q, want %d %q", c.method, c.path, w.Code, w.Body.String(), c.status, c.body)
		}
	}
}

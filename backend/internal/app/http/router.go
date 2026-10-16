package http

import (
	"net/http"
	"strings"
)

// Router — «МЕТОД /путь/{параметр}»: ServeMux в TinyGo (сборка Worker) таких шаблонов не понимает.
type Router struct {
	routes []route
}

type route struct {
	method string
	parts  []string
	h      http.Handler
}

func NewRouter() *Router { return &Router{} }

func (rt *Router) Handle(pattern string, h http.Handler) {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		panic("router: pattern must be \"METHOD /path\": " + pattern)
	}

	rt.routes = append(rt.routes, route{method: method, parts: split(path), h: h})
}

func (rt *Router) HandleFunc(pattern string, fn http.HandlerFunc) { rt.Handle(pattern, fn) }

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := split(r.URL.Path)

	pathMatched := false
	for _, route := range rt.routes {
		params, ok := match(route.parts, parts)
		if !ok {
			continue
		}

		pathMatched = true

		if route.method != r.Method {
			continue
		}

		for name, value := range params {
			r.SetPathValue(name, value)
		}

		route.h.ServeHTTP(w, r)

		return
	}

	if pathMatched {
		WriteError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	WriteError(w, http.StatusNotFound, "не найдено")
}

func split(path string) []string { return strings.Split(strings.Trim(path, "/"), "/") }

func match(pattern, path []string) (map[string]string, bool) {
	if len(pattern) != len(path) {
		return nil, false
	}

	var params map[string]string
	for i, p := range pattern {
		if name, ok := strings.CutPrefix(p, "{"); ok && strings.HasSuffix(name, "}") {
			if params == nil {
				params = map[string]string{}
			}

			params[strings.TrimSuffix(name, "}")] = path[i]

			continue
		}

		if p != path[i] {
			return nil, false
		}
	}

	return params, true
}

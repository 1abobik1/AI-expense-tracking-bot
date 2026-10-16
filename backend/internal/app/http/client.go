package http

import "net/http"

// Doer — *http.Client на сервере, fetch в Worker; с *http.Client бандл Worker растёт на 1.3 МБ.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

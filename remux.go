// Provides an HTTP router that routes requests based on regex patterns.
//
// Different than http.ServeMux, routing is based on Request.RequestURI (unescaped)
// rather than Request.URL.Path (escaped).
// You can't use http.ServeMux if you want to embed a URL in URL. That's why Remux is written.
//
// Named match groups are saved in the Request.
// You can access them by prefixing the group name with ":".
package remux

import (
	"net/http"
	"net/url"
	"regexp"
)

type Remux struct {
	NotFoundHandler http.Handler
	handlers        []*Handler
}

// Handler is type of returned value from Remux's Handle and HandleFunc methods.
// You can restrict the handler to run only on specified methods by calling Get, Head, Post, Put, Delete and Options methods.
// These methods are chainable.
type Handler struct {
	pattern        *regexp.Regexp
	allowedMethods map[string]struct{}
	http.Handler
}

// Handle adds a new handler for requests matching regex.
// Panics if regex does not compile.
// When a request received, registered handlers are checked in the same order they are registered.
func (r *Remux) Handle(regex string, handler http.Handler) *Handler {
	pattern := regexp.MustCompile(regex)
	h := &Handler{pattern, nil, handler}
	r.handlers = append(r.handlers, h)
	return h
}

func (r *Remux) HandleFunc(regex string, f func(http.ResponseWriter, *http.Request)) *Handler {
	return r.Handle(regex, http.HandlerFunc(f))
}

func (r Remux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, h := range r.handlers {
		submatches := h.pattern.FindStringSubmatch(req.RequestURI)
		if len(submatches) == 0 {
			continue
		}
		if h.allowedMethods != nil {
			if _, ok := h.allowedMethods[req.Method]; !ok {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}
		names := h.pattern.SubexpNames()
		params := make(url.Values)
		for i := range names {
			params.Add(":"+names[i], submatches[i])
		}
		req.URL.RawQuery = params.Encode() + "&" + req.URL.RawQuery
		h.ServeHTTP(w, req)
		return
	}
	if r.NotFoundHandler != nil {
		r.NotFoundHandler.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

// method restricts the handler to be executed only on specific method.
func (h *Handler) method(m string) *Handler {
	if h.allowedMethods == nil {
		h.allowedMethods = make(map[string]struct{})
	}
	h.allowedMethods[m] = struct{}{}
	return h
}

func (h *Handler) Head() *Handler    { return h.method("HEAD") }
func (h *Handler) Get() *Handler     { return h.method("GET") }
func (h *Handler) Post() *Handler    { return h.method("POST") }
func (h *Handler) Put() *Handler     { return h.method("PUT") }
func (h *Handler) Delete() *Handler  { return h.method("DELETE") }
func (h *Handler) Options() *Handler { return h.method("OPTIONS") }

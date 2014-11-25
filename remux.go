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
	handlers        []*regexHandler
}

type regexHandler struct {
	pattern        *regexp.Regexp
	allowedMethods map[string]struct{}
	http.Handler
}

// Handle adds a new handler for requests matching regex.
// Panics if regex does not compile.
// When a request received, registered handlers are checked in FIFO order.
func (r *Remux) Handle(regex string, handler http.Handler) *regexHandler {
	pattern := regexp.MustCompile(regex)
	h := &regexHandler{pattern, nil, handler}
	r.handlers = append(r.handlers, h)
	return h
}

func (r *Remux) HandleFunc(regex string, f func(http.ResponseWriter, *http.Request)) *regexHandler {
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
		req.URL.RawQuery = url.Values(params).Encode() + "&" + req.URL.RawQuery
		h.ServeHTTP(w, req)
		return
	}
	if r.NotFoundHandler != nil {
		r.NotFoundHandler.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

func (h *regexHandler) Method(m string) *regexHandler {
	if h.allowedMethods == nil {
		h.allowedMethods = make(map[string]struct{})
	}
	h.allowedMethods[m] = struct{}{}
	return h
}

func (h *regexHandler) Head() *regexHandler    { return h.Method("HEAD") }
func (h *regexHandler) Get() *regexHandler     { return h.Method("GET") }
func (h *regexHandler) Post() *regexHandler    { return h.Method("POST") }
func (h *regexHandler) Put() *regexHandler     { return h.Method("PUT") }
func (h *regexHandler) Delete() *regexHandler  { return h.Method("DELETE") }
func (h *regexHandler) Options() *regexHandler { return h.Method("OPTIONS") }

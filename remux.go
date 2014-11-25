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
	pattern *regexp.Regexp
	http.Handler
}

// Handle adds a new handler for requests matching regex.
// Panics if regex does not compile.
// When a request received, registered handlers are checked in FIFO order.
func (s *Remux) Handle(regex string, handler http.Handler) {
	pattern := regexp.MustCompile(regex)
	h := &regexHandler{pattern, handler}
	s.handlers = append(s.handlers, h)
}

func (s *Remux) HandleFunc(regex string, f func(w http.ResponseWriter, r *http.Request)) {
	s.Handle(regex, http.HandlerFunc(f))
}

func (s Remux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, h := range s.handlers {
		submatches := h.pattern.FindStringSubmatch(r.RequestURI)
		if len(submatches) > 0 {
			names := h.pattern.SubexpNames()
			params := make(url.Values)
			for i := range names {
				params.Add(":"+names[i], submatches[i])
			}
			r.URL.RawQuery = url.Values(params).Encode() + "&" + r.URL.RawQuery
			h.ServeHTTP(w, r)
			return
		}
	}
	if s.NotFoundHandler != nil {
		s.NotFoundHandler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

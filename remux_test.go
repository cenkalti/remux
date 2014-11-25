package remux_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cenkalti/remux"
)

func Example() {
	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello", r.FormValue(":name"))
	}

	var r remux.Remux
	r.HandleFunc("/hello/(?P<name>.+)", helloHandler)

	go http.ListenAndServe("127.0.0.1:5000", r)

	http.Get("http://localhost:5000/hello/Cenk")
	// Output: Hello Cenk
}

func TestRemux(t *testing.T) {
	called := make(chan struct{})
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		close(called)
	}

	var r remux.Remux
	r.HandleFunc("/asdf", testHandler)

	req, _ := http.NewRequest("GET", "http://localhost/asdf", nil)
	req.RequestURI = "/asdf"

	r.ServeHTTP(nil, req)

	select {
	case <-called:
	default:
		t.Fatal("handler not called")
	}
}

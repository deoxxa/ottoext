package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"

	"fknsrs.biz/p/ottoext/loop"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func TestFetch(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	s := httptest.NewServer(m)
	defer s.Close()

	vm := otto.New()
	l := loop.New(vm)

	if err := Define(vm, l); err != nil {
		panic(err)
	}

	must(l.EvalAndRun(`fetch('` + s.URL + `').then(function(r) {
    return r.text();
  }).then(function(d) {
    if (d.indexOf('hellox') === -1) {
      throw new Error('what');
    }
  });`))
}

func TestFetchCallback(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	s := httptest.NewServer(m)
	defer s.Close()

	vm := otto.New()
	l := loop.New(vm)

	if err := Define(vm, l); err != nil {
		panic(err)
	}

	ch := make(chan bool, 1)

	if err := vm.Set("__capture", func(s string) {
		defer func() { ch <- true }()

		if !strings.Contains(s, "hello") {
			panic(fmt.Errorf("expected to find `hello' in response"))
		}
	}); err != nil {
		panic(err)
	}

	must(l.EvalAndRun(`fetch('` + s.URL + `').then(function(r) {
    return r.text();
  }).then(__capture)`))

	<-ch
}

func TestFetchHeaders(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("header-one", "1")
		w.Header().Add("header-two", "2a")
		w.Header().Add("header-two", "2b")

		w.Write([]byte("hello"))
	})
	s := httptest.NewServer(m)
	defer s.Close()

	vm := otto.New()
	l := loop.New(vm)

	if err := Define(vm, l); err != nil {
		panic(err)
	}

	ch := make(chan bool, 1)

	if err := vm.Set("__capture", func(s string) {
		defer func() { ch <- true }()

		if s != `{"header-one":["1"],"header-two":["2a","2b"]}` {
			panic(fmt.Errorf("expected headers to contain 1, 2a, and 2b"))
		}
	}); err != nil {
		panic(err)
	}

	must(l.EvalAndRun(`fetch('` + s.URL + `').then(function(r) {
    return __capture(JSON.stringify({
      'header-one': r.headers.getAll('header-one'),
      'header-two': r.headers.getAll('header-two'),
    }));
  })`))

	<-ch
}

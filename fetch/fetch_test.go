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

	must(l.EvalAndRun(`fetch('http://` + s.Config.Addr + `/').then(function(r) {
    return r.text();
  }).then(function(d) {
  	if (d.indexOf('hello') === -1) {
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

	if err := vm.Set("__capture", func(s string) {
		if !strings.Contains(s, "hello") {
			panic(fmt.Errorf("expected to find `hello' in response"))
		}
	}); err != nil {
		panic(err)
	}

	must(l.EvalAndRun(`fetch('` + s.Config.Addr + `').then(function(r) {
		return r.text();
	}).then(__capture)`))
}

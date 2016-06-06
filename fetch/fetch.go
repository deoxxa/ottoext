package fetch // import "fknsrs.biz/p/ottoext/fetch"

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/robertkrimen/otto"

	"fknsrs.biz/p/ottoext/loop"
	"fknsrs.biz/p/ottoext/promise"
)

func mustValue(v otto.Value, err error) otto.Value {
	if err != nil {
		panic(err)
	}

	return v
}

type fetchTask struct {
	id           int64
	jsReq, jsRes *otto.Object
	cb           otto.Value
	err          error
	status       int
	statusText   string
	headers      map[string][]string
	body         []byte
}

func (t *fetchTask) SetID(id int64) { t.id = id }
func (t *fetchTask) GetID() int64   { return t.id }

func (t *fetchTask) Execute(vm *otto.Otto, l *loop.Loop) error {
	var arguments []interface{}

	if t.err != nil {
		e, err := vm.Call(`new Error`, nil, t.err.Error())
		if err != nil {
			return err
		}

		arguments = append(arguments, e)
	}

	t.jsRes.Set("status", t.status)
	t.jsRes.Set("statusText", t.statusText)
	h := mustValue(t.jsRes.Get("headers")).Object()
	for k, vs := range t.headers {
		for _, v := range vs {
			if _, err := h.Call("append", k, v); err != nil {
				return err
			}
		}
	}
	t.jsRes.Set("_body", string(t.body))

	if _, err := t.cb.Call(otto.NullValue(), arguments...); err != nil {
		return err
	}

	return nil
}

func (t *fetchTask) Cancel() {
}

func Define(vm *otto.Otto, l *loop.Loop) error {
	return DefineWithHandler(vm, l, nil)
}

func DefineWithHandler(vm *otto.Otto, l *loop.Loop, h http.Handler) error {
	if err := promise.Define(vm, l); err != nil {
		return err
	}

	jsData := rice.MustFindBox("dist-fetch").MustString("bundle.js")
	smData := rice.MustFindBox("dist-fetch").MustString("bundle.js.map")

	s, err := vm.CompileWithSourceMap("fetch-bundle.js", jsData, smData)
	if err != nil {
		return err
	}

	if _, err := vm.Run(s); err != nil {
		return err
	}

	vm.Set("__private__fetch_execute", func(c otto.FunctionCall) otto.Value {
		jsReq := c.Argument(0).Object()
		jsRes := c.Argument(1).Object()
		cb := c.Argument(2)

		method := mustValue(jsReq.Get("method")).String()
		urlStr := mustValue(jsReq.Get("url")).String()
		jsBody := mustValue(jsReq.Get("body"))
		var body io.Reader
		if jsBody.IsString() {
			body = strings.NewReader(jsBody.String())
		}

		// Make map for headers to pass in
		reqHeaders := make(map[string]string)

		reqHdrObj := mustValue(jsReq.Get("headers")).Object()
		// TODO: Get list of headers from jsReq object (idea: add 'list' method to headers.js?)
		// Alternatively: iterate such as in: for (var k in r.headers._headers) { console.log(k + ':', r.headers.get(k)); }
		for _, key := range []string{"Authorization" , "X-Amz-Date", "Content-MD5", "Content-Type", "X-Amz-User-Agent", "X-Amz-Content-Sha256"} {

			var v otto.Value
			var err error
			if v, err = reqHdrObj.Call("get", key); err == nil {
				if !v.IsNull() && v.String() != "undefined" {
					reqHeaders[key] = v.String()
				}
			}
		}

		t := &fetchTask{
			jsReq: jsReq,
			jsRes: jsRes,
			cb:    cb,
		}

		l.Add(t)

		go func() {
			defer l.Ready(t)

			req, err := http.NewRequest(method, urlStr, body)
			for k, v := range reqHeaders {
				req.Header.Set(k, v) // Set headers
			}

			if err != nil {
				t.err = err
				return
			}

			if h != nil && urlStr[0] == '/' {
				res := httptest.NewRecorder()

				h.ServeHTTP(res, req)

				t.status = res.Code
				t.statusText = http.StatusText(res.Code)
				t.headers = res.Header()
				t.body = res.Body.Bytes()
			} else {
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					t.err = err
					return
				}

				d, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.err = err
					return
				}

				t.status = res.StatusCode
				t.statusText = res.Status
				t.headers = res.Header
				t.body = d
			}
		}()

		return otto.UndefinedValue()
	})

	return nil
}

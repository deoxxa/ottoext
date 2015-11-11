package promise // import "fknsrs.biz/p/ottoext/promise"

import (
	"bytes"

	"github.com/GeertJohan/go.rice"
	"github.com/MathieuTurcotte/sourcemap"
	"github.com/robertkrimen/otto"

	"fknsrs.biz/p/ottoext/loop"
	"fknsrs.biz/p/ottoext/timers"
)

type compileWithSourcemap interface {
	CompileWithSourceMap(filename string, src interface{}, sm *sourcemap.SourceMap) (*otto.Script, error)
}

func Define(vm *otto.Otto, l *loop.Loop) error {
	if v, err := vm.Get("Promise"); err != nil {
		return err
	} else if !v.IsUndefined() {
		return nil
	}

	if err := timers.Define(vm, l); err != nil {
		return err
	}

	var v interface{} = vm
	var s *otto.Script

	src := rice.MustFindBox("dist-promise").MustString("bundle.js")

	if withSourcemap, ok := v.(compileWithSourcemap); ok {
		sm, err := sourcemap.Read(bytes.NewReader(rice.MustFindBox("dist-promise").MustBytes("bundle.js.map")))
		if err != nil {
			return err
		}

		s, err = withSourcemap.CompileWithSourceMap("promise-bundle.js", src, &sm)
		if err != nil {
			return err
		}
	} else {
		var err error

		s, err = vm.Compile("promise-bundle.js", src)
		if err != nil {
			return err
		}
	}

	if _, err := vm.Run(s); err != nil {
		return err
	}

	return nil
}

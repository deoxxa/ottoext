package promise // import "fknsrs.biz/p/ottoext/promise"

import (
	"bytes"

	"github.com/GeertJohan/go.rice"
	"github.com/MathieuTurcotte/sourcemap"
	"github.com/robertkrimen/otto"

	"fknsrs.biz/p/ottoext/loop"
	"fknsrs.biz/p/ottoext/timers"
	"fknsrs.biz/p/ottoext/types"
)

func Define(vm types.BasicVM, l *loop.Loop) error {
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

	if svm, ok := v.(types.SourceMapVM); ok {
		sm, err := sourcemap.Read(bytes.NewReader(rice.MustFindBox("dist-promise").MustBytes("bundle.js.map")))
		if err != nil {
			return err
		}

		s, err = svm.CompileWithSourceMap("promise-bundle.js", src, &sm)
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

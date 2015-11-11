package types

import (
	"github.com/MathieuTurcotte/sourcemap"
	"github.com/robertkrimen/otto"
)

type BasicVM interface {
	Get(name string) (otto.Value, error)
	Set(name string, value interface{}) error
	Compile(filename string, src interface{}) (*otto.Script, error)
	Call(source string, this interface{}, argumentList ...interface{}) (otto.Value, error)
	Run(src interface{}) (otto.Value, error)
}

type SourceMapVM interface {
	BasicVM

	CompileWithSourceMap(filename string, src interface{}, sm *sourcemap.SourceMap) (*otto.Script, error)
}

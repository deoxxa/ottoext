package loop

import (
	"errors"

	"fknsrs.biz/p/ottoext/types"
	"github.com/robertkrimen/otto"
)

type IdleTask struct {
	ID int64
}

func NewIdleTask() *IdleTask {
	return &IdleTask{}
}

func (i *IdleTask) SetID(ID int64) { i.ID = ID }
func (i IdleTask) GetID() int64    { return i.ID }
func (i IdleTask) Cancel()         {}
func (i IdleTask) Execute(vm types.BasicVM, l *Loop) error {
	return errors.New("IDle task should never execute")
}

type EvalTask struct {
	ID     int64
	Script *otto.Script
	Result chan error
}

func NewEvalTask(s *otto.Script) *EvalTask {
	return &EvalTask{
		Script: s,
		Result: make(chan error, 1),
	}
}

func (e *EvalTask) SetID(ID int64) { e.ID = ID }
func (e EvalTask) GetID() int64    { return e.ID }
func (e EvalTask) Cancel()         {}
func (e EvalTask) Execute(vm types.BasicVM, l *Loop) error {
	_, err := vm.Run(e.Script)
	e.Result <- err
	return err
}

type CallTask struct {
	ID       int64
	Function otto.Value
	Args     []interface{}
	Result   chan error
}

func NewCallTask(fn otto.Value, args ...interface{}) *CallTask {
	return &CallTask{
		Function: fn,
		Args:     args,
		Result:   make(chan error, 1),
	}
}

func (c *CallTask) SetID(ID int64) { c.ID = ID }
func (c CallTask) GetID() int64    { return c.ID }
func (c CallTask) Cancel()         {}
func (c CallTask) Execute(vm types.BasicVM, l *Loop) error {
	_, err := c.Function.Call(otto.NullValue(), c.Args...)
	c.Result <- err
	return err
}

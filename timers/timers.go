package timers

import (
	"time"

	"github.com/robertkrimen/otto"

	"fknsrs.biz/p/ottoext/loop"
)

func Define(vm *otto.Otto, l *loop.Loop) error {
	if v, err := vm.Get("setTimeout"); err != nil {
		return err
	} else if !v.IsUndefined() {
		return nil
	}

	newTimer := func(interval bool) func(call otto.FunctionCall) otto.Value {
		return func(call otto.FunctionCall) otto.Value {
			delay, _ := call.Argument(1).ToInteger()
			if delay <= 0 {
				delay = 1
			}

			t := &timerTask{
				duration: time.Duration(delay) * time.Millisecond,
				call:     call,
				interval: interval,
			}
			l.Add(t)

			t.timer = time.AfterFunc(t.duration, func() {
				l.Ready(t)
			})

			value, err := call.Otto.ToValue(t)
			if err != nil {
				panic(err)
			}

			return value
		}
	}
	vm.Set("setTimeout", newTimer(false))
	vm.Set("setInterval", newTimer(true))

	clearTimeout := func(call otto.FunctionCall) otto.Value {
		v, _ := call.Argument(0).Export()
		if t, ok := v.(*timerTask); ok {
			t.stopped = true
			t.timer.Stop()
			l.Remove(t)
		}

		return otto.UndefinedValue()
	}
	vm.Set("clearTimeout", clearTimeout)
	vm.Set("clearInterval", clearTimeout)

	return nil
}

type timerTask struct {
	id       int64
	timer    *time.Timer
	duration time.Duration
	interval bool
	call     otto.FunctionCall
	stopped  bool
}

func (t *timerTask) SetID(id int64) { t.id = id }
func (t *timerTask) GetID() int64   { return t.id }

func (t *timerTask) Execute(vm *otto.Otto, l *loop.Loop) error {
	var arguments []interface{}

	if len(t.call.ArgumentList) > 2 {
		tmp := t.call.ArgumentList[2:]
		arguments = make([]interface{}, 2+len(tmp))

		for i, value := range tmp {
			arguments[i+2] = value
		}
	} else {
		arguments = make([]interface{}, 1)
	}

	arguments[0] = t.call.ArgumentList[0]

	if _, err := vm.Call(`Function.call.call`, nil, arguments...); err != nil {
		return err
	}

	if t.interval && !t.stopped {
		t.timer.Reset(t.duration)
		l.Add(t)
	}

	return nil
}

func (t *timerTask) Cancel() {
	t.timer.Stop()
}

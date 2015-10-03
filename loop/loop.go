package loop // import "fknsrs.biz/p/ottoext/loop"

import (
	"sync"
	"sync/atomic"

	"github.com/robertkrimen/otto"
)

type Task interface {
	SetID(id int64)
	GetID() int64
	Execute(vm *otto.Otto, l *Loop) error
	Cancel()
}

type Loop struct {
	vm     *otto.Otto
	id     int64
	lock   sync.RWMutex
	tasks  map[int64]Task
	ready  chan Task
	closed bool
}

func New(vm *otto.Otto) *Loop {
	return &Loop{
		vm:    vm,
		tasks: make(map[int64]Task),
		ready: make(chan Task),
	}
}

func (l *Loop) Add(t Task) {
	l.lock.Lock()
	t.SetID(atomic.AddInt64(&l.id, 1))
	l.tasks[t.GetID()] = t
	l.lock.Unlock()
}

func (l *Loop) Remove(t Task) {
	l.remove(t)
	go l.Ready(nil)
}

func (l *Loop) remove(t Task) {
	l.removeByID(t.GetID())
}

func (l *Loop) removeByID(id int64) {
	l.lock.Lock()
	delete(l.tasks, id)
	l.lock.Unlock()
}

func (l *Loop) Ready(t Task) {
	if l.closed {
		return
	}

	l.ready <- t
}

func (l *Loop) EvalAndRun(s interface{}) error {
	if err := l.Eval(s); err != nil {
		return err
	}

	return l.Run()
}

func (l *Loop) Eval(s interface{}) error {
	if _, err := l.vm.Run(s); err != nil {
		return err
	}

	return nil
}

func (l *Loop) Run() error {
	i := 0

	for t := range l.ready {
		i++

		if t != nil {
			id := t.GetID()

			if err := t.Execute(l.vm, l); err != nil {
				l.lock.RLock()
				for _, t := range l.tasks {
					t.Cancel()
				}
				l.lock.RUnlock()

				return err
			}

			l.removeByID(id)
		}

		l.lock.RLock()
		if len(l.tasks) == 0 {
			close(l.ready)
			l.closed = true
		}
		l.lock.RUnlock()
	}

	return nil
}

func (l *Loop) Step() (error, bool) {
	var tasks []Task

outer:
	for {
		select {
		case t := <-l.ready:
			tasks = append(tasks, t)

			l.remove(t)
		default:
			break outer
		}
	}

	for _, t := range tasks {
		if err := t.Execute(l.vm, l); err != nil {
			l.lock.RLock()
			for _, t := range l.tasks {
				t.Cancel()
			}
			l.lock.RUnlock()

			return err, false
		}
	}

	return nil, len(l.tasks) == 0
}

package loop // import "fknsrs.biz/p/ottoext/loop"

import (
	"fmt"
	"sync"
	"sync/atomic"

	"fknsrs.biz/p/ottoext/types"
)

func formatTask(t Task) string {
	if t == nil {
		return "<nil>"
	}

	return fmt.Sprintf("<%T> %d", t, t.GetID())
}

type Task interface {
	SetID(id int64)
	GetID() int64
	Execute(vm types.BasicVM, l *Loop) error
	Cancel()
}

type Loop struct {
	vm     types.BasicVM
	id     int64
	lock   sync.RWMutex
	tasks  map[int64]Task
	ready  chan Task
	closed bool
}

func New(vm types.BasicVM) *Loop {
	return NewWithBacklog(vm, 0)
}

func NewWithBacklog(vm types.BasicVM, backlog int) *Loop {
	return &Loop{
		vm:    vm,
		tasks: make(map[int64]Task),
		ready: make(chan Task, backlog),
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

func (l *Loop) processTask(t Task) error {
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

	return nil
}

func (l *Loop) Run() error {
	for {
		l.lock.Lock()
		if len(l.tasks) == 0 {
			// prevent any more tasks entering the ready channel
			l.closed = true

			l.lock.Unlock()

			break
		}
		l.lock.Unlock()

		t := <-l.ready

		if t != nil {
			if err := l.processTask(t); err != nil {
				return err
			}
		}
	}

	// drain ready channel of any existing tasks
outer:
	for {
		select {
		case t := <-l.ready:
			if t != nil {
				if err := l.processTask(t); err != nil {
					return err
				}
			}
		default:
			break outer
		}
	}

	close(l.ready)

	return nil
}

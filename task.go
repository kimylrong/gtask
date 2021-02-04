package taskpool

import (
	"sync"
)

type SyncTask func() (interface{}, error)

type AsyncTask func()

type taskResult struct {
	result interface{}
	err    error
}

type taskWrap struct {
	isSignal  bool
	isSync    bool
	syncTask  SyncTask
	asyncTask AsyncTask
	notifier  chan taskResult
}

var (
	taskPool sync.Pool
)

func init() {
	taskPool.New = func() interface{} {
		return &taskWrap{}
	}
}

func getTask() *taskWrap {
	return taskPool.Get().(*taskWrap)
}

func putTask(t *taskWrap) {
	t.zero()
	taskPool.Put(t)
}

func (t *taskWrap) zero() {
	t.isSignal = false
	t.isSync = false
	t.notifier = nil
	t.syncTask = nil
	t.asyncTask = nil
}

func (t *taskWrap) run() (interface{}, error) {
	result, err := func() (r interface{}, e error) {
		defer func() {
			if panic := recover(); panic != nil {
				e = errPoolPanic
			}
		}()
		if t.isSync {
			r, e = t.syncTask()
		} else {
			t.asyncTask()
		}
		return
	}()
	return result, err
}

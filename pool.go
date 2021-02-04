package taskpool

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	poolStatusOpened = int32(1)
	poolStatusClosed = int32(0)
)

const (
	defaultMinPoolSize = int32(4)
	defaultMaxPoolSize = int32(256)
	defaultQueueSize   = int32(1024)
)

var minLifeTime = int64(60)

type TaskPool struct {
	minSize   int32
	maxSize   int32
	queueSize int32

	status int32
	size   int32

	workerSeq int32
	works     []*worker

	tasks     chan *taskWrap
	taskCount int32

	notifier chan struct{}
}

func (p *TaskPool) init() {
	for i := int32(0); i < p.size; i++ {
		newWorker(p)
	}

	go p.alterSize()
}

func (p *TaskPool) alterSize() {
	for {
		// wait
		<-p.notifier
		// batch fetch notifier
		//count := len(p.notifier)
		//for i := 0; i < count; i++ {
		//	<-p.notifier
		//}

		if p.IsClosed() {
			fmt.Printf("IsClosed\n")
			return
		}
		size := atomic.LoadInt32(&p.size)
		if len(p.tasks) < 1 && size > p.minSize {
			// free, then reduce pool size
			now := time.Now().Unix()
			worker := p.works[size-1] // slowly reduce worker count. reduce one worker once.
			createTime := worker.createTime
			if (now - createTime) < minLifeTime {
				continue // worker is too young, give more chance
			}

			atomic.StoreInt32(&worker.status, workerStatusClosed)
			newSize := size - 1
			for i := int32(0); i < size; i++ {
				task := getTask()
				task.isSignal = true
				p.tasks <- task // put nil task to queue, invoke waiting goroutine
			}
			p.works = p.works[:newSize]
			atomic.StoreInt32(&p.size, newSize)
		} else if len(p.tasks) > int(p.minSize) && size < p.maxSize {
			// busy, then add pool size
			newSize := size << 1 // double workers
			if newSize < int32(len(p.tasks)) {
				newSize = int32(len(p.tasks))
			}
			if newSize > p.maxSize {
				newSize = p.maxSize
			}
			atomic.StoreInt32(&p.size, newSize)
			addCount := newSize - size
			for i := int32(0); i < addCount; i++ {
				newWorker(p)
			}
		}
	}
}

func (p *TaskPool) AsyncRunTask(task AsyncTask) error {
	if p.IsClosed() {
		return errPoolClosed
	}
	if size := int(atomic.LoadInt32(&p.size)); len(p.tasks) > size && size < int(p.maxSize) {
		p.notifier <- struct{}{}
	}
	taskWrap := getTask()
	taskWrap.asyncTask = task
	p.tasks <- taskWrap
	atomic.AddInt32(&p.taskCount, 1)
	return nil
}

func (p *TaskPool) SyncRunTask(task SyncTask) (interface{}, error) {
	if p.IsClosed() {
		return nil, errPoolClosed
	}
	if size := int(atomic.LoadInt32(&p.size)); len(p.tasks) > size && size < int(p.maxSize) {
		// busy, notify alter pool size
		p.notifier <- struct{}{}
	}
	taskWrap := getTask()
	taskWrap.isSync = true
	taskWrap.syncTask = task
	taskWrap.notifier = make(chan taskResult, 1) // don't block worker, notifier has buffer size 1
	p.tasks <- taskWrap
	atomic.AddInt32(&p.taskCount, 1)
	r := <-taskWrap.notifier
	close(taskWrap.notifier)
	putTask(taskWrap)
	return r.result, r.err
}

func (p *TaskPool) SyncRunTaskWithTimeout(task SyncTask, timeout time.Duration) (result interface{}, err error) {
	if p.IsClosed() {
		return nil, errPoolClosed
	}
	if size := int(atomic.LoadInt32(&p.size)); len(p.tasks) > size && size < int(p.maxSize) {
		// busy, notify alter pool size
		p.notifier <- struct{}{}
	}
	taskWrap := getTask()
	taskWrap.isSync = true
	taskWrap.syncTask = task
	taskWrap.notifier = make(chan taskResult, 1) // don't block worker, notifier has buffer size 1
	p.tasks <- taskWrap
	atomic.AddInt32(&p.taskCount, 1)

	timer := time.NewTimer(timeout)
	select {
	case r := <-taskWrap.notifier:
		result, err = r.result, r.err
		close(taskWrap.notifier)
		putTask(taskWrap)
	case <-timer.C:
		err = errPoolTimeout
		close(taskWrap.notifier)
		// if timeout, still can't release taskWrap here, not safe
	}
	timer.Stop()
	return
}

func (p *TaskPool) IsClosed() bool {
	if status := atomic.LoadInt32(&p.status); status == poolStatusClosed {
		return true
	}
	return false
}

func (p *TaskPool) Close() {
	atomic.StoreInt32(&p.status, poolStatusClosed)
	for _, r := range p.works {
		atomic.StoreInt32(&r.status, workerStatusClosed)
	}
	close(p.tasks)
	close(p.notifier)
}

func (p *TaskPool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.size)
}

func (p *TaskPool) TaskCount() int32 {
	return atomic.LoadInt32(&p.taskCount)
}

func (p *TaskPool) TaskMinSize() int32 {
	return atomic.LoadInt32(&p.minSize)
}

func (p *TaskPool) TaskMaxSize() int32 {
	return atomic.LoadInt32(&p.maxSize)
}

func (p *TaskPool) TaskQueueSize() int32 {
	return atomic.LoadInt32(&p.queueSize)
}

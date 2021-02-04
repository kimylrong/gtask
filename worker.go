package taskpool

import (
	"sync/atomic"
	"time"
)

const (
	workerStatusOpened = int32(1)
	workerStatusClosed = int32(0)
)

type worker struct {
	id         int32
	status     int32 // 1: Open, 0:Close
	createTime int64
}

func newWorker(p *TaskPool) {
	w := &worker{
		id:         atomic.AddInt32(&p.workerSeq, 1),
		status:     workerStatusOpened,
		createTime: time.Now().Unix(),
	}
	p.works = append(p.works, w)
	//fmt.Printf("start worker %d\n", w.id)
	go func(w *worker) {
		for {
			if atomic.LoadInt32(&w.status) == workerStatusClosed {
				// worker closed, exit
				// fmt.Printf("end goroutine %d\n", w.id)
				return
			}
			taskWrap := <-p.tasks
			if taskWrap.isSignal {
				putTask(taskWrap)
				continue
			}

			result, err := taskWrap.run()
			if !taskWrap.isSync {
				putTask(taskWrap)
			}

			if taskWrap.notifier != nil {
				select {
				case <-taskWrap.notifier:
					// notifier has been closed, release taskWrap here
					putTask(taskWrap)
				default:
					taskWrap.notifier <- taskResult{
						result: result,
						err:    err,
					}
				}
			}

			taskCount := atomic.AddInt32(&p.taskCount, -1)
			poolSize := atomic.LoadInt32(&p.size)
			if taskCount < poolSize && poolSize > p.minSize {
				// free, notify alter pool size
				p.notifier <- struct{}{}
			}
		}
	}(w)
}

package taskpool

import (
	"errors"
	"time"
)

var (
	errPoolSizeNotValid  = errors.New("pool size not valid")
	errQueueSizeNotValid = errors.New("pool size not valid")
	errPoolFull          = errors.New("pool is full")
	errPoolTimeout       = errors.New("timeout")
	errPoolPanic         = errors.New("panic")
	errPoolClosed        = errors.New("pool closed")
)

var defaultPool *TaskPool

func init() {
	defaultPool, _ = NewTaskPool()
}

type Option func(p *TaskPool)

func WithMinSize(minSize int32) Option {
	return func(p *TaskPool) {
		p.minSize = minSize
	}
}

func WithMaxSize(maxSize int32) Option {
	return func(p *TaskPool) {
		p.maxSize = maxSize
	}
}

func WithQueueSize(queueSize int32) Option {
	return func(p *TaskPool) {
		p.queueSize = queueSize
	}
}

func NewTaskPool(options ...Option) (*TaskPool, error) {
	pool := &TaskPool{
		status: poolStatusOpened,
	}
	for _, option := range options {
		option(pool)
	}
	// fill default value
	if pool.minSize == 0 {
		pool.minSize = defaultMinPoolSize
	}
	if pool.maxSize == 0 {
		pool.maxSize = defaultMaxPoolSize
	}
	if pool.queueSize == 0 {
		pool.queueSize = defaultQueueSize
	}

	err := checkPool(pool)
	if err != nil {
		return nil, err
	}

	pool.size = pool.minSize // size init to minSize
	pool.works = make([]*worker, 0, pool.minSize)
	pool.tasks = make(chan *taskWrap, pool.queueSize)
	pool.notifier = make(chan struct{}, pool.queueSize)
	pool.init()
	return pool, nil
}

func checkPool(p *TaskPool) error {
	if p.minSize < 1 || p.minSize > p.maxSize {
		return errPoolSizeNotValid
	}
	if p.queueSize < 1 {
		return errQueueSizeNotValid
	}
	return nil
}

func AsyncRunTask(task AsyncTask) error {
	return defaultPool.AsyncRunTask(task)
}

func SyncRunTask(task SyncTask) (interface{}, error) {
	return defaultPool.SyncRunTask(task)
}

func SyncRunTaskWithTimeout(task SyncTask, timeout time.Duration) (result interface{}, err error) {
	return defaultPool.SyncRunTaskWithTimeout(task, timeout)
}

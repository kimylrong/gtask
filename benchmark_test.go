package taskpool

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkTaskPoolAsync(b *testing.B) {
	var wait sync.WaitGroup
	var taskPool, _ = NewTaskPool()
	for i := 0; i < b.N; i++ {
		err := taskPool.AsyncRunTask(func() {
			doSomething()
			wait.Done()
		})
		if err != nil {
			fmt.Println("ERRRRRRRRRRRRRRRRRRR")
			continue
		}
		wait.Add(1)
	}
	wait.Wait()
}

func BenchmarkTaskPoolSync(b *testing.B) {
	var wait sync.WaitGroup
	var taskPool, _ = NewTaskPool()
	for i := 0; i < b.N; i++ {
		taskPool.SyncRunTask(func() (interface{}, error) {
			doSomething()
			wait.Done()
			return nil, nil
		})
		wait.Add(1)
	}
	wait.Wait()
}

func doSomething() {
	for i := 1; i < 1; i++ {
		_ = i << 2
	}
}

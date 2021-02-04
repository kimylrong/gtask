# Motivation

## Performance

Yes, goroutine is cheap. But goroutine is not free. If we create too much goroutine, it will add heavy load
to golang scheduler, also it wil add heavy load to golang GC. So make a goroutine pool is reasonable.

## Safety

If we run a task with `go somefunc()()`, and `somefunc` panic, application will exit. we need recover panic.

# Design

## Task

Task has two type, SyncTask and AsyncTask. SyncTask is for sync run task. SyncTask has a result, caller can get it.

## Worker

Fetch task from pool, run task, notify task result. Work is goroutine in fact.

## Queue

Hold tasks waiting for process. Queue is implemented with channel.

## Pool

Pool hold queue, hold worker. Pool can extend count of worker if busy. Pool can reduce count of worker if free.

# Example

## Sync run task

```
task := func()(interface{}, err){
    return nil, nil
}
result, err := SyncRunTask(task)
```

## Sync run task with timeout

```
task := func()(interface{}, err){
    return nil, nil
}
result, err := SyncRunTaskWithTimeout(task, time.Second)
```

## Async run task

```
task := func(){}
AsyncRunTask(task)
```

# Benchmark
```
BenchmarkTaskPoolAsync-8         1791747               630 ns/op              17 B/op          1 allocs/op
BenchmarkTaskPoolSync-8          1571095               760 ns/op             144 B/op          3 allocs/op
```

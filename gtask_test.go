package taskpool

import (
	"errors"
	"testing"
	"time"
)

func TestNewTaskPoolErr(t *testing.T) {
	cases := []struct {
		minSize   int32
		maxSize   int32
		queueSize int32
	}{
		{
			minSize:   -1,
			maxSize:   1,
			queueSize: 1,
		},
		{
			minSize:   1,
			maxSize:   -1,
			queueSize: 1,
		},
		{
			minSize:   2,
			maxSize:   1,
			queueSize: 1,
		},
		{
			minSize:   1,
			maxSize:   1,
			queueSize: -1,
		},
	}

	for _, c := range cases {
		_, err := NewTaskPool(WithMinSize(c.minSize), WithMaxSize(c.maxSize), WithQueueSize(c.queueSize))
		if err == nil {
			t.Fail()
		}
	}
}

func TestNewTaskPool(t *testing.T) {
	_, err := NewTaskPool(WithMinSize(1), WithMaxSize(1), WithQueueSize(1))
	if err != nil {
		t.Fail()
	}
}

func TestSyncRunTask(t *testing.T) {
	cases := []struct {
		f            func() (interface{}, error)
		expectResult int32
		expectErr    bool
	}{
		{
			f: func() (interface{}, error) {
				return int32(1), nil
			},
			expectResult: 1,
			expectErr:    false,
		},
		{
			f: func() (interface{}, error) {
				return int32(-1), nil
			},
			expectResult: -1,
			expectErr:    false,
		},
		{
			f: func() (interface{}, error) {
				return int32(0), errors.New("error happens")
			},
			expectResult: 0,
			expectErr:    true,
		},
	}

	for _, c := range cases {
		result, err := SyncRunTask(c.f)
		if c.expectErr {
			if err == nil {
				t.Fail()
			}
		} else {
			if err != nil {
				t.Fail()
			}
		}

		real := result.(int32)
		if real != c.expectResult {
			t.Fail()
		}
	}
}

func TestSyncRunTaskWithTimeout(t *testing.T) {
	cases := []struct {
		f         func() (interface{}, error)
		timeout   time.Duration
		expectErr bool
	}{
		{
			f: func() (interface{}, error) {
				time.Sleep(time.Second * 2)
				return nil, nil
			},
			timeout:   time.Second * 1,
			expectErr: true,
		},
		{
			f: func() (interface{}, error) {
				time.Sleep(time.Millisecond * 500)
				return nil, nil
			},
			timeout:   time.Second * 1,
			expectErr: false,
		},
	}

	for _, c := range cases {
		_, err := SyncRunTaskWithTimeout(c.f, c.timeout)
		if c.expectErr {
			if err == nil {
				t.Fail()
			}
		} else {
			if err != nil {
				t.Fail()
			}
		}
	}
}

func TestAsyncRunTask(t *testing.T) {
	err := AsyncRunTask(func() {
		_ = 1 << 10
	})
	if err != nil {
		t.Fail()
	}
}
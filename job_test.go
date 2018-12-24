package goest_worker

import (
	"testing"
	"context"
	"github.com/pkg/errors"
)

type mockJobCall struct {
	priority	int
}

func (job *mockJobCall) call() {

}

func (job *mockJobCall) Priority() int {
	return job.priority
}

func TestJobFuncResultInt(t *testing.T) {
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()
	sumJob := pool.NewJob(func(a, b int) int {
		return a + b
	})
	var result int
	sumJob.Run(1, 2).Wait().Results(&result)
	if result != 3 {
		t.Errorf("failed result 1 + 2")
	}
}

func TestJobFuncResultTwoInt(t *testing.T) {
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()
	abjob := pool.NewJob(func(a, b int) (int, int) {
		return a, b
	})
	var a, b int
	abjob.Run(1, 2).Wait().Results(&a, &b)
	if a != 1 || b != 2 {
		t.Errorf("invalid answer")
	}
}

func TestJobFuncResultStruct(t *testing.T) {
	type fooBar struct {
		foo string
		bar string
	}
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()
	fooBarPointerJob := pool.NewJob(func(foo, bar string) (*fooBar) {
		return &fooBar{
			foo: foo,
			bar: bar,
		}
	})
	var fb *fooBar
	err := fooBarPointerJob.Run("foo", "bar").Wait().Results(&fb)
	if err != nil {
		t.Error(err)
		return
	}
	if fb.foo != "foo" || fb.bar != "bar" {
		t.Errorf("invalid answer")
		return
	}
}


func TestJobFuncResultStructAndError(t *testing.T) {
	type fooBar struct {
		foo string
		bar string
	}
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()
	fooBarPointerJob := pool.NewJob(func(foo, bar string, fail bool) (*fooBar, error) {
		if fail {
			return nil, errors.New("failed")
		}
		return &fooBar{
			foo: foo,
			bar: bar,
		}, nil
	})
	var fb *fooBar
	var fbErr error
	err := fooBarPointerJob.Run("foo", "bar", false).Wait().Results(&fb, &fbErr)
	if err != nil {
		t.Error(err)
		return
	}
	if fb.foo != "foo" || fb.bar != "bar" {
		t.Errorf("invalid answer")
		return
	}

	err = fooBarPointerJob.Run("foo", "bar", true).Wait().Results(&fb, &fbErr)
	if err != nil {
		t.Error(err)
		return
	}
	if fb != nil {
		t.Error(err)
		return
	}
	if fbErr.Error() != "failed" {
		t.Errorf("invalid answer")
		return
	}
}

func TestJobFuncPointerToSlice(t *testing.T){
	pool := New().Start(context.TODO(), 10)
	defer pool.Stop()

	reverseJob := pool.NewJob(func(slice []int) []int {
		last := len(slice) - 1
		for i := 0; i < len(slice)/2; i++ {
			slice[i], slice[last-i] = slice[last-i], slice[i]
		}
		return slice
	})
	var s = []int{1,2,3}
	reverseJob.Run(s).Wait().Results(&s)
	if s[2] != 1 {
		t.Errorf("invalid answer")
		return
	}

	var ss = []int{1,2,3}
	var dd = []int{}
	reverseJob.Run(ss).Wait().Results(&dd)
	if ss[2] != 1 {
		t.Errorf("invalid answer")
		return
	}
	if dd[2] != 1 {
		t.Errorf("invalid answer")
		return
	}
}
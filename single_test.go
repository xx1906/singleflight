package singleflight

import (
	"sync/atomic"
	"testing"
	"time"
)

var g = NewGroup()

type TestCase struct {
	fn  func() (interface{}, error)
	err error
	val interface{}
	key string
}

var execTime int32

func exec() (interface{}, error) {
	time.Sleep(time.Millisecond)
	atomic.AddInt32(&execTime, 1)
	return "world", nil
}
func exec1() (interface{}, error) {
	return "hello", nil
}
func TestGroup_DoChan(t *testing.T) {
	cases := []TestCase{
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec, err: nil, val: "world", key: "hello"},
		{fn: exec1, err: nil, val: "hello", key: "world"},
		{fn: exec1, err: nil, val: "hello", key: "world"},
		{fn: exec1, err: nil, val: "hello", key: "world"},
	}
	for _, v := range cases {
		go func(v TestCase) {
			val := <-g.DoChan(v.key, v.fn)
			outRes, outErr := val.Value, val.Err
			t.Log(outRes, outErr, v.val, v.err)

		}(v)
		go func(v TestCase) {
			val, err := g.DoCall(v.key, v.fn)
			t.Log(val, err, v.err, v.val)
		}(v)
	}
	time.Sleep(time.Millisecond * 10)
	t.Log(len(g.single), execTime)
}

/***********************************************************************************************************************
* single-flight engine, if you'll need the multi clients to fetch the sample data, may be you need this engine.
* Just like the multi clients to fetch the sample data from the redis. In this case single-flight model just use once
* client to fetch data from redis or other datasource, and then this client auto dispatch the data to other clients.
* If your applications has cases like this, use single-flight will have perfect performance.
*
*
* @author  小超人
* @date    2021-09-01
* @version 0.0.1
*
************************************************************************************************************************/

package singleflight

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	group = NewGroup()
)

type Group struct {
	mu     sync.Mutex
	single map[string]*call
}

type call struct {
	result interface{}
	err    error

	done   chan struct{}
	refJob int32
}

type result struct {
	Value interface{}
	Err   error
}

func NewGroup() *Group {
	return &Group{
		mu:     sync.Mutex{},
		single: make(map[string]*call, 4),
	}
}

// return execute function interface{} and error
func (c *Group) DoCall(key string, execute func() (interface{}, error)) (value interface{}, err error) {
	v := <-c.doChan(key, execute)
	return v.Value, v.Err
}

func DoCall(key string, execute func() (interface{}, error)) (res interface{}, err error) {
	return group.DoCall(key, execute)
}

// return chan of Result
// NOTICE: if already have the key, it will not replace with the new execute
func (c *Group) doChan(key string, execute func() (interface{}, error)) <-chan result {
	c.mu.Lock()
	defer c.mu.Unlock()

	r := make(chan result)
	var ca *call
	var ok bool

	if ca, ok = c.single[key]; !ok {
		ca = &call{done: make(chan struct{})}
		// if single is nil
		if c.single == nil {
			c.single = make(map[string]*call, 4)
		}
		c.single[key] = ca
		go func() {
			defer func() {
				if err := recover(); err != nil {
					_ = fmt.Errorf("execute panic:%s", err)
				}
			}()
			ca.result, ca.err = execute()
			ca.done <- struct{}{}
			close(ca.done)
		}()
	}
	// add job ref
	atomic.AddInt32(&ca.refJob, 1)

	go func() {
		// waiting for execute return
		<-ca.done
		// dispatch the return to the r channel
		r <- result{Err: ca.err, Value: ca.result}
		close(r)
		c.releaseJob(key, ca) // release this job
	}()

	return r
}

// releaseJob, if refJob is zero, deleteJob
func (c *Group) releaseJob(key string, ca *call) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if atomic.AddInt32(&ca.refJob, -1) == 0 {
		delete(c.single, key)
	}
}

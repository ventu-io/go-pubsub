// Copyright 2015 Ventu.io. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file

package lpoll

import (
	"errors"
	"sync"
	"time"
)

type LPoll struct {
	mx   sync.Mutex
	subs map[string]*Sub
}

func New() *LPoll {
	return &LPoll{
		subs: make(map[string]*Sub),
	}
}

func (lp *LPoll) Publish(data interface{}, topics ...string) {
	// lock for iterating over map
	lp.mx.Lock()
	defer lp.mx.Unlock()
	for _, topic := range topics {
		for _, sub := range lp.subs {
			go sub.Publish(data, topic)
		}
	}
}

func (lp *LPoll) Subscribe(timeout time.Duration, topics ...string) string {
	sub := NewSub(timeout, func(id string) {
		// lock for deletion
		lp.mx.Lock()
		delete(lp.subs, id)
		lp.mx.Unlock()
	}, topics...)
	// lock for element insertion
	lp.mx.Lock()
	lp.subs[sub.id] = sub
	lp.mx.Unlock()
	return sub.id
}

func (lp *LPoll) Get(id string, polltime time.Duration) (chan []interface{}, error) {
	// do not lock
	if sub, ok := lp.subs[id]; ok {
		return sub.Get(polltime), nil
	} else {
		return nil, errors.New("incorrect subscription id")
	}
}

func (lp *LPoll) Drop(id string) {
	if sub, ok := lp.subs[id]; ok {
		go sub.Drop()
		// lock for deletion
		lp.mx.Lock()
		delete(lp.subs, id)
		lp.mx.Unlock()
	}
}

func (lp *LPoll) Shutdown() {
	// lock for iterating over map
	lp.mx.Lock()
	defer lp.mx.Unlock()
	for id, sub := range lp.subs {
		go sub.Drop()
		delete(lp.subs, id)
	}
}

func (lp *LPoll) List() []string {
	// lock for iterating over map
	lp.mx.Lock()
	defer lp.mx.Unlock()
	var res []string
	for id, _ := range lp.subs {
		res = append(res, id)
	}
	return res
}

func (lp *LPoll) Topics() []string {
	topics := make(map[string]bool)
	// lock for iterating over map
	lp.mx.Lock()
	for _, sub := range lp.subs {
		for topic, _ := range sub.topics {
			topics[topic] = true
		}
	}
	lp.mx.Unlock()

	var res []string
	for topic, _ := range topics {
		res = append(res, topic)
	}
	return res
}

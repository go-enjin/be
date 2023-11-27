// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signaling

import "sync"

const (
	SignalServePage Signal = "serve-page"
	SignalServePath Signal = "serve-path"
)

type Signal string

type Listener func(signal Signal, tag string, data []interface{}, argv []interface{}) (stop bool)

type Signaling interface {
	SignalSupport
	EmitterSupport
}

type EmitterSupport interface {
	Emit(signal Signal, tag string, argv ...interface{}) (stopped bool)
}

type SignalSupport interface {
	Connect(signal Signal, handle string, l Listener, data ...interface{})
	Disconnect(signal Signal, handle string)
}

type cListener struct {
	H string
	L Listener
	D []interface{}
}

type CSignaling struct {
	signaling     map[Signal][]*cListener
	signalingLock *sync.RWMutex
}

func (c *CSignaling) InitSignaling() {
	c.signaling = make(map[Signal][]*cListener)
	c.signalingLock = &sync.RWMutex{}
}

func (c *CSignaling) Emit(signal Signal, tag string, argv ...interface{}) (stopped bool) {
	c.signalingLock.RLock()
	defer c.signalingLock.RUnlock()
	if num := len(c.signaling[signal]); num > 0 {
		for i := num - 1; i >= 0; i-- {
			if stopped = c.signaling[signal][i].L(signal, tag, c.signaling[signal][i].D, argv); stopped {
				return
			}
		}
	}
	return
}

func (c *CSignaling) Connect(signal Signal, handle string, l Listener, data ...interface{}) {
	c.signalingLock.Lock()
	defer c.signalingLock.Unlock()
	c.signaling[signal] = append(c.signaling[signal], &cListener{H: handle, L: l, D: data})
	return
}

func (c *CSignaling) Disconnect(signal Signal, handle string) {
	c.signalingLock.Lock()
	defer c.signalingLock.Unlock()
	var modified []*cListener
	for _, sl := range c.signaling[signal] {
		if sl.H != handle {
			modified = append(modified, sl)
		}
	}
	c.signaling[signal] = modified
	return
}

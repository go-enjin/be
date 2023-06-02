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

package be

import (
	"github.com/go-enjin/be/pkg/feature/signaling"
)

type signalListener struct {
	H string
	L signaling.Listener
	D []interface{}
}

func (e *Enjin) Emit(signal signaling.Signal, tag string, argv ...interface{}) (stopped bool) {
	e.signalingLock.RLock()
	defer e.signalingLock.RUnlock()
	if num := len(e.signaling[signal]); num > 0 {
		for i := num - 1; i >= 0; i-- {
			if stopped = e.signaling[signal][i].L(signal, tag, e.signaling[signal][i].D, argv); stopped {
				return
			}
		}
	}
	return
}

func (e *Enjin) Connect(signal signaling.Signal, handle string, l signaling.Listener, data ...interface{}) {
	e.signalingLock.Lock()
	defer e.signalingLock.Unlock()
	e.signaling[signal] = append(e.signaling[signal], &signalListener{H: handle, L: l, D: data})
	return
}

func (e *Enjin) Disconnect(signal signaling.Signal, handle string) {
	e.signalingLock.Lock()
	defer e.signalingLock.Unlock()
	var modified []*signalListener
	for _, sl := range e.signaling[signal] {
		if sl.H != handle {
			modified = append(modified, sl)
		}
	}
	e.signaling[signal] = modified
	return
}
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

package feature

type LifeCycleState uint8

const (
	StateCreated LifeCycleState = iota
	StateConstructed
	StateInitialized
	StateMade
	StateBuilt
	StateSetup
	StateStarted
	StatePostStarted
	StateShutdown
	EndLifeCycleStates
)

func (l LifeCycleState) Valid() (valid bool) {
	valid = l < EndLifeCycleStates
	return
}

func (l LifeCycleState) String() (name string) {
	switch l {
	case StateCreated:
		name = "created"
	case StateConstructed:
		name = "constructed"
	case StateInitialized:
		name = "initialized"
	case StateMade:
		name = "made"
	case StateBuilt:
		name = "built"
	case StateSetup:
		name = "setup"
	case StateStarted:
		name = "started"
	case StatePostStarted:
		name = "postStarted"
	case StateShutdown:
		name = "shutdown"
	default:
		name = "invalid"
	}
	return
}
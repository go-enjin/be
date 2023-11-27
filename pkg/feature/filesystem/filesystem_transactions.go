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

package filesystem

import (
	"github.com/go-enjin/be/pkg/log"
)

//func (f *CFeature[MakeTypedFeature]) Transaction(fn func(ff *CFeature[MakeTypedFeature])) {
//	f.BeginTransactions()
//	defer f.EndTransactions()
//	cloned, _ := f.CFeature.Clone().(*feature.CFeature)
//	fn(&CFeature[MakeTypedFeature]{
//		CFeature: *cloned,
//	})
//}

func (f *CFeature[MakeTypedFeature]) BeginTransactions() {
	f.txLock.Lock()
	log.TraceDF(1, "starting %v transactions", f.Tag())
	for _, mps := range f.MountPoints {
		for _, mp := range mps {
			if mp.RWFS != nil {
				mp.RWFS.BeginTransaction()
				log.TraceDF(1, "beginning %v transactions: %v", f.Tag(), mp.Mount)
			}
		}
	}
}

func (f *CFeature[MakeTypedFeature]) RollbackTransactions() {
	defer f.txLock.Unlock()
	for _, mps := range f.MountPoints {
		for _, mp := range mps {
			if mp.RWFS != nil {
				mp.RWFS.RollbackTransaction()
				log.TraceDF(1, "rolling back %v transactions: %v", f.Tag(), mp.Mount)
			}
		}
	}
}

func (f *CFeature[MakeTypedFeature]) CommitTransactions() {
	defer f.txLock.Unlock()
	for _, mps := range f.MountPoints {
		for _, mp := range mps {
			if mp.RWFS != nil {
				mp.RWFS.CommitTransaction()
				log.TraceDF(1, "committing %v transactions: %v", f.Tag(), mp.Mount)
			}
		}
	}
}

func (f *CFeature[MakeTypedFeature]) EndTransactions() {
	log.TraceDF(1, "ending %v transactions", f.Tag())
	if err := recover(); err != nil {
		f.RollbackTransactions()
	} else {
		f.CommitTransactions()
	}
}

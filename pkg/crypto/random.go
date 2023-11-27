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

package crypto

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	mathRand "math/rand"
)

func RandomValue(numBytes int) (value string, err error) {
	b := make([]byte, numBytes)
	if _, e := cryptoRand.Read(b); e != nil {
		err = fmt.Errorf("crypto/rand error: %w; falling back to math/rand", e)
		if _, ee := mathRand.Read(b); ee != nil {
			err = fmt.Errorf("%w; math/rand error: %w", e, ee)
		}
	}
	value = hex.EncodeToString(b)
	return
}

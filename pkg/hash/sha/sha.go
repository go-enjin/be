// Copyright (c) 2022  The Go-Enjin Authors
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

package sha

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
)

var (
	RxShasum64 = regexp.MustCompile(`^([a-f0-9]{64})$`)
	RxShasum10 = regexp.MustCompile(`^([a-f0-9]{10})$`)
)

func FileHash64(path string) (shasum string, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	shasum = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func DataHash64(data []byte) (shasum string, err error) {
	h := sha256.New()
	if _, err = h.Write(data); err != nil {
		return
	}
	shasum = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func FileHash10(path string) (shasum string, err error) {
	if shasum, err = FileHash64(path); err == nil {
		shasum = shasum[0:10]
	}
	return
}

func DataHash10(data []byte) (shasum string, err error) {
	if shasum, err = DataHash64(data); err == nil {
		shasum = shasum[0:10]
	}
	return
}

func VerifyFile64(sum, file string) (err error) {
	var fh *os.File
	if fh, err = os.Open(file); err != nil {
		return
	}
	defer func() { _ = fh.Close() }()
	shaHash := sha256.New()
	if _, err = io.Copy(shaHash, fh); err != nil {
		return
	}
	hash := fmt.Sprintf("%x", shaHash.Sum(nil))
	if hash != sum {
		err = fmt.Errorf("shasum mismatch %v", file)
		return
	}
	return
}

func DataHashSha1(data []byte) (shasum string, err error) {
	h := sha1.New()
	if _, err = h.Write(data); err != nil {
		return
	}
	shasum = fmt.Sprintf("%x", h.Sum(nil))
	return
}
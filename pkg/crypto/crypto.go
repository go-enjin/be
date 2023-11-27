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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

var (
	ErrBadCipherText = errors.New("bad ciphertext")
)

func prepare(key []byte) (shasum [32]byte, block cipher.Block, gcm cipher.AEAD, err error) {
	shasum = sha256.Sum256(key)
	if block, err = aes.NewCipher(shasum[:]); err != nil {
		return
	}

	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}

	return
}

func NewKey256() (key []byte) {
	key = make([]byte, 0, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return
}

func Encrypt(key, unencrypted []byte) (encoded string, err error) {
	//var block cipher.Block
	var gcm cipher.AEAD
	if _, _ /*block*/, gcm, err = prepare(key); err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	encrypted := gcm.Seal(nonce, nonce, unencrypted, nil)
	encoded = hex.EncodeToString(encrypted)

	return
}

func Decrypt(key []byte, encoded string) (decrypted []byte, err error) {
	var encrypted []byte
	if encrypted, err = hex.DecodeString(encoded); err != nil {
		return
	}

	var gcm cipher.AEAD
	if _, _, gcm, err = prepare(key); err != nil {
		return
	}

	if len(encrypted) < gcm.NonceSize() {
		err = ErrBadCipherText
		return
	}

	decrypted, err = gcm.Open(
		nil,
		encrypted[:gcm.NonceSize()],
		encrypted[gcm.NonceSize():],
		nil)
	return
}

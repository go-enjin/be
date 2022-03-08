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

package autocert

import (
	"context"
	"errors"

	"golang.org/x/crypto/acme/autocert"
	"gorm.io/gorm"

	"github.com/go-enjin/be/pkg/database"
)

func NewManager(email string, domains ...string) *autocert.Manager {
	return &autocert.Manager{
		Cache:      DatabaseCache(""),
		Prompt:     autocert.AcceptTOS,
		Email:      email,
		HostPolicy: autocert.HostWhitelist(domains...),
	}
}

type Model struct {
	gorm.Model

	Name string
	Cert []byte
}

type DatabaseCache string

func (dc DatabaseCache) Get(ctx context.Context, name string) (data []byte, err error) {
	var db *gorm.DB
	if db, err = database.Get(); err != nil {
		return
	}
	done := make(chan struct{})
	var record Model
	go func() {
		err = db.Find(&record, `name = ?`, name).Error
		close(done)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, autocert.ErrCacheMiss
	}
	data = record.Cert
	return
}

func (dc DatabaseCache) Put(ctx context.Context, name string, data []byte) (err error) {
	var db *gorm.DB
	if db, err = database.Get(); err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		record := Model{
			Name: name,
			Cert: data,
		}
		err = db.Create(&record).Error
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return
}

func (dc DatabaseCache) Delete(ctx context.Context, name string) (err error) {
	var db *gorm.DB
	if db, err = database.Get(); err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		var model Model
		err = db.Where(`name = ?`, name).Delete(&model).Error
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return
}
//go:build srv_fastcgi || srv || all

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

package fastcgi

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yookoala/gofast"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/serve"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
)

type service struct {
	DirIndex string
	Target   string
	Domains  []string
	EnvKeys  []string

	handler http.Handler
}

func newHandler(dirIndex, target, network, source string, envKeys []string) (s *service, err error) {
	if target, err = bePath.Abs(target); err != nil {
		return
	}
	if network == "auto" {
		if bePath.IsFile(source) {
			network = "unix"
		} else {
			network = "tcp"
		}
	}
	var h gofast.Handler
	if bePath.IsDir(target) {
		h = gofast.NewHandler(
			newPhpFS(dirIndex, target, envKeys)(gofast.BasicSession),
			gofast.SimpleClientFactory(gofast.SimpleConnFactory(network, source)),
		)
		log.DebugF("fastcgi target is a directory: %v", target)
	} else if bePath.IsFile(target) {
		h = gofast.NewHandler(
			gofast.NewFileEndpoint(target)(gofast.BasicSession),
			gofast.SimpleClientFactory(gofast.SimpleConnFactory(network, source)),
		)
		log.DebugF("fastcgi target is a file: %v", target)
	} else {
		err = fmt.Errorf("target is not a file or a directory")
		return
	}
	lh := &logger{next: h}
	s = &service{
		EnvKeys:  envKeys,
		DirIndex: dirIndex,
		Target:   target,
		handler:  lh,
	}
	return
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(s.Domains) > 0 {
		var err error
		var host string
		if host, _, err = net.SplitHostPort(r.Host); err != nil {
			host = r.Host
		}
		if !slices.Present(host, s.Domains...) {
			log.WarnF("rejecting unsupported domain: %v", r.Host)
			serve.Serve404(w, r)
			return
		}
	}

	path := filepath.Join(s.Target, r.URL.Path)
	if !strings.HasSuffix(path, ".php") && bePath.IsFile(path) {
		http.ServeFile(w, r, path)
		return
	}

	s.handler.ServeHTTP(w, r)
}
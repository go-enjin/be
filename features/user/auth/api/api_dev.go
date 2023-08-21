//go:build (user_auth_api && dev) || (user_auths && dev) || all

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

package api

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/github-com-go-pkgz-auth/provider"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
)

var DefaultDevAuthServerPort = 8086

type MakeDevAuthSupport interface {
	EnableDevAuthServer(enabled bool) MakeFeature
}

type DevAuthSupport struct {
	enableDevAuthService bool
	devAuthServerPort    int
	devAuthServerHost    string
}

func (f *CFeature) EnableDevAuthServer(enabled bool) MakeFeature {
	f.enableDevAuthService = enabled
	return f
}

func (f *CFeature) BuildDevAuthService(b feature.Buildable) (err error) {

	tag := f.Tag().String()

	b.AddFlags(
		&cli.IntFlag{
			Category: tag,
			Name:     globals.MakeFlagName(tag, "dev-auth-port"),
			EnvVars:  globals.MakeFlagEnvKeys(tag, "dev-auth-port"),
			Value:    DefaultDevAuthServerPort,
		},
		&cli.StringFlag{
			Category: tag,
			Name:     globals.MakeFlagName(tag, "dev-auth-host"),
			EnvVars:  globals.MakeFlagEnvKeys(tag, "dev-auth-host"),
			Value:    "localhost",
		},
	)

	return
}

func (f *CFeature) StartupDevAuthService(ctx *cli.Context) (err error) {

	if !f.enableDevAuthService {
		return
	}

	tag := f.Tag().String()
	port := ctx.Int(globals.MakeFlagName(tag, "dev-auth-port"))
	host := ctx.String(globals.MakeFlagName(tag, "dev-auth-host"))

	log.WarnF("!!!!! %v FEATURE INCLUDES RUNNING DEV AUTH SERVICE (%s:%d) !!!!!", f.Tag(), host, port)

	f.authService.AddDevProvider(host, port)

	go func() {

		var devAuthServer *provider.DevAuthServer
		if devAuthServer, err = f.authService.DevAuth(); err != nil {
			log.FatalF("%v feature error starting dev auth service: %v", tag, err)
			return
		}

		devAuthServer.Provider.Host = host
		devAuthServer.Provider.Port = port
		// devAuthServer.Automatic = true

		devAuthServer.GetEmailFn = func(username string) string {
			return username + "@localhost.nope"
		}

		log.InfoF("%s feature starting dev auth service on: %s:%d", tag, host, port)
		devAuthServer.Run(context.Background())

	}()

	return
}
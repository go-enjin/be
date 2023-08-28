package handler

import (
	"net/http"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	beLogContext "github.com/go-enjin/be/types/logging/context"
	"github.com/go-enjin/be/types/logging/response"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-logging-handler"

type Feature interface {
	feature.Feature
	feature.ServiceLogHandler
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) LogHandler(next http.Handler) (this http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		when := time.Now()
		logger, writer := response.NewLogger(w)
		next.ServeHTTP(writer, r)
		duration := time.Now().Sub(when)

		go func() {
			// forked to not block request any further, not doing this actually adds to the duration despite the
			// request being handled already by next.ServeHTTP
			if ctx, err := beLogContext.New(r, logger.Size(), logger.Status(), when, duration); err != nil {
				log.ErrorDF(1, "error making logger context: %v", err)
			} else {
				for _, rl := range feature.FilterTyped[feature.ServiceLogger](f.Enjin.Features().List()) {
					if err = rl.RequestLogger(ctx); err != nil {
						log.ErrorDF(1, "%v request logger error: %v", rl.Tag(), err)
					}
				}
			}
		}()

	})
}
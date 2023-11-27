package logger

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

// TODO: rewrite gorilla-handlers functions into more formal implementation
// TODO: consider using something like https://github.com/lestrrat-go/apache-logformat

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-logging-logger"

type Feature interface {
	feature.Feature
	feature.ServiceLogger
}

type MakeFeature interface {
	Make() Feature

	SetCombined(enabled bool) MakeFeature
}

type CFeature struct {
	feature.CFeature

	combined bool
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

func (f *CFeature) SetCombined(enabled bool) MakeFeature {
	f.combined = enabled
	return f
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

func (f *CFeature) RequestLogger(ctx feature.LoggerContext) (err error) {
	if f.combined {
		writeCombinedLog(log.InfoWriter(), ctx)
	} else {
		writeLog(log.InfoWriter(), ctx)
	}
	return
}

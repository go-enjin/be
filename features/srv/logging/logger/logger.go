package logger

import (
	"fmt"
	"io"

	syslog "github.com/dmachard/go-clientsyslog"

	"github.com/sirupsen/logrus"
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

const Tag feature.Tag = "srv-reqlogger"

type Feature interface {
	feature.Feature
	feature.ServiceLogger
}

type MakeFeature interface {
	Make() Feature

	SetLogging(config *log.Configuration) MakeFeature
	SetCombined(enabled bool) MakeFeature
}

type CFeature struct {
	feature.CFeature

	network string
	address string

	combined bool
	config   *log.Configuration
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) SetLogging(config *log.Configuration) MakeFeature {
	f.config = config
	return f
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
	category := f.Tag().String()
	snake := f.Tag().Snake()
	b.AddFlags(
		&cli.StringFlag{
			Name:     "reqlogger-dsn",
			Usage:    "remote syslog DSN",
			EnvVars:  b.MakeEnvKeys(snake + "_DSN"),
			Value:    "",
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	if dsnArg := ctx.String("reqlogger-dsn"); dsnArg != "" {
		// empty --reqlogger-host means do nothing
		var dsn *log.SyslogDSN
		if dsn, err = log.ParseSyslogDSN(dsnArg); err != nil {
			return
		}

		if f.config == nil {
			f.config = log.NewConfig()
		}
		f.config.LogTag = log.Config.LogTag
		f.config.AppName = log.Config.AppName
		f.config.LogHook = f.Tag().Kebab()

		var hook logrus.Hook
		if hook, err = log.NewSyslogHook(dsn, syslog.LOG_INFO, f.config.AppName); err != nil {
			return fmt.Errorf("failed to setup reqlogger (%v): %w", dsn, err)
		} else {
			f.config.LogHooks = append(f.config.LogHooks, hook)
		}

		if dsn.Options.Has("Format") {
			if lf, ok := dsn.Options.Get("Format").(log.Format); ok {
				f.config.LoggingFormat = lf
			}
		}

		f.config.Apply()
		log.DebugF("configured reqlogger: %v", dsn)
	}

	return
}

func (f *CFeature) RequestLogger(ctx feature.LoggerContext) (err error) {
	var config *log.Configuration
	var writer *io.PipeWriter
	if f.config != nil {
		config = f.config
		writer = f.config.InfoWriter()
	} else {
		config = &log.Config.Configuration
		writer = log.InfoWriter()
	}

	config.RLock()
	defer config.RUnlock()

	if config.LoggingFormat == log.FormatJson {
		if f.combined {
			writeCombinedJsonLog(writer, ctx)
		} else {
			writeJsonLog(writer, ctx)
		}
		return
	}

	if f.combined {
		writeCombinedLog(writer, ctx)
	} else {
		writeLog(writer, ctx)
	}
	return
}

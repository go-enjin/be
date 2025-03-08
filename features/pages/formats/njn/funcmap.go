package njn

import (
	"html/template"

	clContext "github.com/go-corelibs/context"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) MakeFuncMap(ctx clContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"njn": func(content string) (output template.HTML, err error) {
			return f.renderContent(ctx, content)
		},
	}
	return
}

func (f *CFeature) renderContent(input clContext.Context, content string) (output template.HTML, err error) {
	// TODO: this is a modified version of renderContent from features/funcmaps/funcmaps.go
	var ctx clContext.Context
	if input.Len() > 0 { // dynamic funcmap
		ctx = input
	} else { // static funcmap
		ctx = f.Enjin.Context(nil)
	}
	var redirect string
	if output, redirect, err = f.Process(ctx, content); err == nil {
		log.TraceF("njn rendered content format success: %v", f.Name())
	} else if redirect != "" {
		log.DebugF("njn rendered content wanted to redirect to: %q", redirect)
	}
	return
}

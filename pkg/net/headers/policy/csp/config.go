// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package csp

import (
	"fmt"

	"github.com/go-enjin/be/pkg/slices"
)

type ContentSecurityPolicyConfig struct {
	GenericSourceDirective Sources
	DefaultSrc             Sources
	ConnectSrc             Sources
	FontSrc                Sources
	FrameSrc               Sources
	ImgSrc                 Sources
	ManifestSrc            Sources
	MediaSrc               Sources
	ObjectSrc              Sources
	PrefetchSrc            Sources
	ScriptSrc              Sources
	ScriptSrcElem          Sources
	ScriptSrcAttr          Sources
	StyleSrc               Sources
	StyleSrcElem           Sources
	StyleSrcAttr           Sources
	WorkerSrc              Sources
	BaseUri                Sources
	FormAction             Sources
	FrameAncestors         Sources
}

func (c ContentSecurityPolicyConfig) Merge(other ContentSecurityPolicyConfig) (merged ContentSecurityPolicyConfig) {
	merged = ContentSecurityPolicyConfig{
		GenericSourceDirective: slices.Merge(c.GenericSourceDirective, other.GenericSourceDirective),
		DefaultSrc:             slices.Merge(c.DefaultSrc, other.DefaultSrc),
		ConnectSrc:             slices.Merge(c.ConnectSrc, other.ConnectSrc),
		FontSrc:                slices.Merge(c.FontSrc, other.FontSrc),
		FrameSrc:               slices.Merge(c.FrameSrc, other.FrameSrc),
		ImgSrc:                 slices.Merge(c.ImgSrc, other.ImgSrc),
		ManifestSrc:            slices.Merge(c.ManifestSrc, other.ManifestSrc),
		MediaSrc:               slices.Merge(c.MediaSrc, other.MediaSrc),
		ObjectSrc:              slices.Merge(c.ObjectSrc, other.ObjectSrc),
		PrefetchSrc:            slices.Merge(c.PrefetchSrc, other.PrefetchSrc),
		ScriptSrc:              slices.Merge(c.ScriptSrc, other.ScriptSrc),
		ScriptSrcElem:          slices.Merge(c.ScriptSrcElem, other.ScriptSrcElem),
		ScriptSrcAttr:          slices.Merge(c.ScriptSrcAttr, other.ScriptSrcAttr),
		StyleSrc:               slices.Merge(c.StyleSrc, other.StyleSrc),
		StyleSrcElem:           slices.Merge(c.StyleSrcElem, other.StyleSrcElem),
		StyleSrcAttr:           slices.Merge(c.StyleSrcAttr, other.StyleSrcAttr),
		WorkerSrc:              slices.Merge(c.WorkerSrc, other.WorkerSrc),
		BaseUri:                slices.Merge(c.BaseUri, other.BaseUri),
		FormAction:             slices.Merge(c.FormAction, other.FormAction),
		FrameAncestors:         slices.Merge(c.FrameAncestors, other.FrameAncestors),
	}
	return
}

func (c ContentSecurityPolicyConfig) Apply(policy Policy) (modified Policy) {
	apply := func(key string, s Sources, p Policy) (m Policy) {
		if m = p; len(s) > 0 {
			m = p.Add(&directive{name: key, sources: s})
		}
		return
	}
	modified = policy
	modified = apply("default-src", c.DefaultSrc, modified)
	modified = apply("connect-src", c.ConnectSrc, modified)
	modified = apply("font-src", c.FontSrc, modified)
	modified = apply("frame-src", c.FrameSrc, modified)
	modified = apply("img-src", c.ImgSrc, modified)
	modified = apply("manifest-src", c.ManifestSrc, modified)
	modified = apply("media-src", c.MediaSrc, modified)
	modified = apply("object-src", c.ObjectSrc, modified)
	modified = apply("prefetch-src", c.PrefetchSrc, modified)
	modified = apply("script-src", c.ScriptSrc, modified)
	modified = apply("script-src-elem", c.ScriptSrcElem, modified)
	modified = apply("script-src-attr", c.ScriptSrcAttr, modified)
	modified = apply("style-src", c.StyleSrc, modified)
	modified = apply("style-src-elem", c.StyleSrcElem, modified)
	modified = apply("style-src-attr", c.StyleSrcAttr, modified)
	modified = apply("worker-src", c.WorkerSrc, modified)
	modified = apply("base-uri", c.BaseUri, modified)
	modified = apply("form-action", c.FormAction, modified)
	modified = apply("frame-ancestors", c.FrameAncestors, modified)
	return
}

func ParseContentSecurityPolicyConfig(ctx map[string]interface{}) (cspc ContentSecurityPolicyConfig, err error) {
	var cfgErr ConfigError
	parseSource := func(key string, ctx map[string]interface{}, bucket Sources) (modified Sources) {
		if things, ok := ctx[key].([]interface{}); ok {
			for idx, thing := range things {
				if src, ok := thing.(string); ok {
					if parsed, ok := ParseSource(src); ok {
						bucket = append(
							bucket,
							parsed,
						)
					} else {
						cfgErr = cfgErr.addError(
							fmt.Sprintf(
								"failed to parse content-security-policy.%s[%d]=\"%v\"",
								key, idx, src,
							),
						)
					}
				}
			}
		}
		modified = bucket
		return
	}
	cspc.DefaultSrc = parseSource("default-src", ctx, cspc.DefaultSrc)
	cspc.ConnectSrc = parseSource("connect-src", ctx, cspc.ConnectSrc)
	cspc.FontSrc = parseSource("font-src", ctx, cspc.FontSrc)
	cspc.FrameSrc = parseSource("frame-src", ctx, cspc.FrameSrc)
	cspc.ImgSrc = parseSource("img-src", ctx, cspc.ImgSrc)
	cspc.ManifestSrc = parseSource("manifest-src", ctx, cspc.ManifestSrc)
	cspc.MediaSrc = parseSource("media-src", ctx, cspc.MediaSrc)
	cspc.ObjectSrc = parseSource("object-src", ctx, cspc.ObjectSrc)
	cspc.PrefetchSrc = parseSource("prefetch-src", ctx, cspc.PrefetchSrc)
	cspc.ScriptSrc = parseSource("script-src", ctx, cspc.ScriptSrc)
	cspc.ScriptSrcElem = parseSource("script-src-elem", ctx, cspc.ScriptSrcElem)
	cspc.ScriptSrcAttr = parseSource("script-src-attr", ctx, cspc.ScriptSrcAttr)
	cspc.StyleSrc = parseSource("style-src", ctx, cspc.StyleSrc)
	cspc.StyleSrcElem = parseSource("style-src-elem", ctx, cspc.StyleSrcElem)
	cspc.StyleSrcAttr = parseSource("style-src-attr", ctx, cspc.StyleSrcAttr)
	cspc.WorkerSrc = parseSource("worker-src", ctx, cspc.WorkerSrc)
	cspc.BaseUri = parseSource("base-uri", ctx, cspc.BaseUri)
	cspc.FormAction = parseSource("form-action", ctx, cspc.FormAction)
	cspc.FrameAncestors = parseSource("frame-ancestors", ctx, cspc.FrameAncestors)

	if !cfgErr.isEmpty() {
		err = cfgErr
	}
	return
}

// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package csp

import (
	"fmt"
	"strings"
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

type ConfigError []string

func (c ConfigError) Error() (msg string) {
	msg = strings.Join(c, "\n")
	return
}

func (c ConfigError) addError(msg string) (modified ConfigError) {
	modified = append(c, msg)
	return
}

func (c ConfigError) isEmpty() (empty bool) {
	empty = len(c) == 0
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
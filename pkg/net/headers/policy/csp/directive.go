// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

type Directive interface {
	DirectiveType() string
	Value() string
}

type SourceDirective interface {
	Directive
	Sources() (sources []Source)
	Append(sources ...Source)
}

var _ Directive = (*directive)(nil)

type directive struct {
	name    string
	sources Sources
}

func NewGenericSourceDirective(name string, sources ...Source) (d Directive) {
	d = &directive{
		name:    name,
		sources: sources,
	}
	return
}

func (d *directive) DirectiveType() string {
	return d.name
}

func (d *directive) Value() (value string) {
	value = d.name
	if len(d.sources) == 0 {
		value += " " + None.Value()
		return
	}
	var sources Sources
	if d.name == "frame-ancestors" {
		// only allow host, scheme, self and none
		sources = d.sources.
			FilterAllowedTypes(HostSourceType, SchemeSourceType, KeywordSourceType).
			FilterAllowedKeywords(None, Self).
			Collapse()
	} else {
		sources = d.sources.FilterUnsafeInline().Collapse()
	}
	for _, s := range sources.Sort() {
		value += " " + s.Value()
	}
	return
}

func (d *directive) Sources() (sources []Source) {
	sources = append(sources, d.sources...)
	return
}

func (d *directive) Append(sources ...Source) {
	var hasNone, hasMore bool
	for _, src := range sources {
		var dupe bool
		for _, dSrc := range d.sources {
			if !hasNone && dSrc.Value() == "'none'" {
				hasNone = true
			} else if !hasMore && dSrc.Value() != "'none'" {
				hasMore = true
			}
			if dupe = dSrc.Value() == src.Value(); dupe {
				break
			}
		}
		if !dupe {
			d.sources = append(d.sources, src)
		}
	}
	if hasNone && hasMore {
		var tmp Sources
		for _, dSrc := range d.sources {
			if dSrc.Value() != "'none'" {
				tmp = append(tmp, dSrc)
			}
		}
		d.sources = tmp
	}
}

func NewDefaultSrc(sources ...Source) Directive {
	return &directive{name: "default-src", sources: sources}
}

func NewConnectSrc(sources ...Source) Directive {
	return &directive{name: "connect-src", sources: sources}
}

func NewFontSrc(sources ...Source) Directive {
	return &directive{name: "font-src", sources: sources}
}

func NewFrameSrc(sources ...Source) Directive {
	return &directive{name: "frame-src", sources: sources}
}

func NewImgSrc(sources ...Source) Directive {
	return &directive{name: "img-src", sources: sources}
}

func NewManifestSrc(sources ...Source) Directive {
	return &directive{name: "manifest-src", sources: sources}
}

func NewMediaSrc(sources ...Source) Directive {
	return &directive{name: "media-src", sources: sources}
}

func NewObjectSrc(sources ...Source) Directive {
	return &directive{name: "object-src", sources: sources}
}

func NewPrefetchSrc(sources ...Source) Directive {
	return &directive{name: "prefetch-src", sources: sources}
}

func NewScriptSrc(sources ...Source) Directive {
	return &directive{name: "script-src", sources: sources}
}

func NewScriptSrcElem(sources ...Source) Directive {
	return &directive{name: "script-src-elem", sources: sources}
}

func NewScriptSrcAttr(sources ...Source) Directive {
	return &directive{name: "script-src-attr", sources: sources}
}

func NewStyleSrc(sources ...Source) Directive {
	return &directive{name: "style-src", sources: sources}
}

func NewStyleSrcElem(sources ...Source) Directive {
	return &directive{name: "style-src-elem", sources: sources}
}

func NewStyleSrcAttr(sources ...Source) Directive {
	return &directive{name: "style-src-attr", sources: sources}
}

func NewWorkerSrc(sources ...Source) Directive {
	return &directive{name: "worker-src", sources: sources}
}

func NewBaseUri(sources ...Source) Directive {
	return &directive{name: "base-uri", sources: sources}
}

func NewFormAction(sources ...Source) Directive {
	return &directive{name: "form-action", sources: sources}
}

func NewFrameAncestors(sources ...Source) Directive {
	return &directive{name: "frame-ancestors", sources: sources}
}
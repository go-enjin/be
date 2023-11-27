// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/serve"
	"github.com/go-enjin/be/pkg/request"
)

const (
	PolicyTag           request.Key = "content-security-policy"
	ReportNonceTag      request.Key = "content-security-policy-report-nonce"
	RequestNonceDataTag request.Key = "content-security-policy-request-nonce-data"
)

const (
	RequestDefaultSrcNonceTag     string = "default-src"
	RequestConnectSrcNonceTag     string = "connect-src"
	RequestFontSrcNonceTag        string = "font-src"
	RequestFrameSrcNonceTag       string = "frame-src"
	RequestImgSrcNonceTag         string = "img-src"
	RequestManifestSrcNonceTag    string = "manifest-src"
	RequestMediaSrcNonceTag       string = "media-src"
	RequestObjectSrcNonceTag      string = "object-src"
	RequestPrefetchSrcNonceTag    string = "prefetch-src"
	RequestScriptSrcNonceTag      string = "script-src"
	RequestScriptSrcElemNonceTag  string = "script-src-elem"
	RequestScriptSrcAttrNonceTag  string = "script-src-attr"
	RequestStyleSrcNonceTag       string = "style-src"
	RequestStyleSrcElemNonceTag   string = "style-src-elem"
	RequestStyleSrcAttrNonceTag   string = "style-src-attr"
	RequestWorkerSrcNonceTag      string = "worker-src"
	RequestBaseUriNonceTag        string = "base-uri"
	RequestFormActionNonceTag     string = "form-action"
	RequestFrameAncestorsNonceTag string = "frame-ancestors"
)

var (
	DefaultReportPathPrefix = "/_/csp-violation"
)

type ModifyPolicyFn = func(policy Policy, r *http.Request) (modified Policy)

type PolicyHandler struct {
	reportNonces map[string]time.Time
	requestNonce map[string]string

	sync.RWMutex
}

type RequestNonceData map[string]string

type PageContextContentSecurity struct {
	Policy Policy
	Nonces beContext.Context
}

func NewPolicyHandler() (h *PolicyHandler) {
	h = &PolicyHandler{
		reportNonces: make(map[string]time.Time),
		requestNonce: make(map[string]string),
	}
	return
}

func (h *PolicyHandler) NewReportNonce() (nonce string) {
	h.Lock()
	defer h.Unlock()
	unique, _ := uuid.NewV4()
	nonce, _ = sha.DataHash10(unique.Bytes())
	h.reportNonces[nonce] = time.Now()
	return
}

func (h *PolicyHandler) ValidateReportNonce(nonce string) (valid bool) {
	h.RLock()
	defer h.RUnlock()
	_, valid = h.reportNonces[nonce]
	return
}

func (h *PolicyHandler) PruneReportNonces() {
	h.Lock()
	defer h.Unlock()
	for nonce, stamp := range h.reportNonces {
		if time.Now().Sub(stamp) >= time.Minute {
			delete(h.reportNonces, nonce)
		}
	}
}

func (h *PolicyHandler) SetRequestPolicy(r *http.Request, policy Policy) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), PolicyTag, policy))
	return
}

func (h *PolicyHandler) GetRequestPolicy(r *http.Request) (policy Policy) {
	if p, ok := r.Context().Value(PolicyTag).(Policy); ok && p != nil {
		policy = p
	} else {
		policy = DefaultContentSecurityPolicy()
	}
	return
}

func (h *PolicyHandler) GetRequestNonce(tag string, r *http.Request) (nonce string, modified *http.Request) {
	var ok bool
	var data *RequestNonceData
	if data, ok = r.Context().Value(RequestNonceDataTag).(*RequestNonceData); ok && data != nil {
		if nonce, ok = (*data)[tag]; ok && nonce != "" {
			modified = r
			return
		}
	}
	if data == nil {
		m := make(RequestNonceData)
		data = &m
	}
	unique, _ := uuid.NewV4()
	nonce, _ = sha.DataHash64(unique.Bytes())
	(*data)[tag] = nonce
	modified = r.Clone(context.WithValue(r.Context(), RequestNonceDataTag, data))
	return
}

func (h *PolicyHandler) GetRequestNonceData(r *http.Request) (data *RequestNonceData, modified *http.Request) {
	var ok bool
	if data, ok = r.Context().Value(RequestNonceDataTag).(*RequestNonceData); ok {
		modified = r
		return
	}
	data = new(RequestNonceData)
	modified = r.Clone(context.WithValue(r.Context(), RequestNonceDataTag, data))
	return
}

func (h *PolicyHandler) PrepareRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.PruneReportNonces()
		if strings.HasPrefix(r.URL.Path, "/_/csp-violation-") {
			if r.Method == http.MethodPost {
				body, _ := io.ReadAll(r.Body)
				pathLen := len(r.URL.Path)
				nonce := r.URL.Path[pathLen-10:]
				if h.ValidateReportNonce(nonce) {
					log.WarnF("content-security-policy violation report received:\n%v", string(body))
				} else {
					log.WarnF("content-security-policy violation report received [expired]:\n%v", string(body))
				}
			}
			serve.Serve204(w, r)
			return
		}
		r = r.Clone(context.WithValue(r.Context(), ReportNonceTag, h.NewReportNonce()))
		r = h.SetRequestPolicy(r, DefaultContentSecurityPolicy())
		next.ServeHTTP(w, r)
	})
}

func (h *PolicyHandler) ModifyPolicyMiddleware(fn ModifyPolicyFn) (mw func(next http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		log.DebugF("including modify content security policy middleware")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if fn != nil {
				p := h.GetRequestPolicy(r)
				m := fn(p, r)
				r = h.SetRequestPolicy(r, m)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *PolicyHandler) PreparePageContext(config ContentSecurityPolicyConfig, ctx beContext.Context, r *http.Request) (pccs *PageContextContentSecurity, modified *http.Request) {

	prepareNonce := func(name string, r *http.Request, p Policy) (m *http.Request) {
		if p == nil || p.None(name) || (name != "script-src" && p.Unsafe(name)) {
			m = r
			return
		}
		var nonce string
		key := name + "-nonce-value"
		nonce, m = h.GetRequestNonce(key, r)
		ctx.Set(key, nonce)
		p.Add(&directive{name: name, sources: []Source{NewNonceSource(nonce)}})
		return
	}

	contentSecurityPolicy := h.GetRequestPolicy(r)
	if contentSecurityPolicy != nil {
		contentSecurityPolicy = config.Apply(contentSecurityPolicy)
	}

	modified = r

	modified = prepareNonce(RequestDefaultSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestConnectSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestFontSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestFrameSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestImgSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestManifestSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestMediaSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestObjectSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestPrefetchSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestScriptSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestScriptSrcElemNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestScriptSrcAttrNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestStyleSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestStyleSrcElemNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestStyleSrcAttrNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestWorkerSrcNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestBaseUriNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestFormActionNonceTag, modified, contentSecurityPolicy)
	modified = prepareNonce(RequestFrameAncestorsNonceTag, modified, contentSecurityPolicy)

	var data *RequestNonceData
	data, modified = h.GetRequestNonceData(modified)

	cspRequestNonces := beContext.New()
	for tag, nonce := range *data {
		cspRequestNonces.Set(tag, nonce)
	}

	pccs = &PageContextContentSecurity{
		Policy: contentSecurityPolicy,
		Nonces: cspRequestNonces,
	}
	return
}

func (h *PolicyHandler) ApplyHeaders(w http.ResponseWriter, r *http.Request) {
	if policy := h.GetRequestPolicy(r); policy != nil {
		if nonce, ok := r.Context().Value(ReportNonceTag).(string); ok {
			reportUri := DefaultReportPathPrefix + "-" + nonce
			value := policy.
				Set(NewReportUri(reportUri)).
				Set(NewReportTo("self-endpoint")).
				Collapse().
				Value()
			w.Header().Set("Reporting-Endpoints", `self-endpoint="`+reportUri+`"`)
			w.Header().Set("Content-Security-Policy", value)
			// log.DebugF("setting request content security policy header: %v", value)
		} else {
			log.ErrorF("request with content security policy missing nonce: %#+v", policy)
		}
	} else {
		log.WarnF("request missing content security policy context")
	}
}

func (h *PolicyHandler) FinalizeRequest(w http.ResponseWriter, r *http.Request) {
	h.ApplyHeaders(w, r)
	return
}

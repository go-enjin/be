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
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func StrictContentSecurityPolicy() Policy {
	return &cPolicy{
		NewDefaultSrc(Self, SchemeSource("https")),
		NewFrameAncestors(None),
		NewObjectSrc(None),
	}
}

func DefaultContentSecurityPolicy() Policy {
	return &cPolicy{
		NewDefaultSrc(Self, SchemeSource("https"), SchemeSource("data"), UnsafeInline),
		NewFrameAncestors(None),
		NewObjectSrc(None),
	}
}

const (
	PolicyTag        beContext.RequestKey = "content-security-policy"
	ReportNonceTag   beContext.RequestKey = "content-security-policy-report-nonce"
	RequestNonceTags beContext.RequestKey = "content-security-policy-request-nonce-tags"
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
	if p, ok := r.Context().Value(PolicyTag).(Policy); ok {
		policy = p
	} else {
		policy = DefaultContentSecurityPolicy()
	}
	return
}

func (h *PolicyHandler) GetRequestNonce(tag string, r *http.Request) (nonce string, modified *http.Request) {
	var ok bool
	h.RLock()
	nonce, ok = h.requestNonce[tag]
	h.RUnlock()
	modified = r
	if !ok {
		unique, _ := uuid.NewV4()
		nonce, _ = sha.DataHash64(unique.Bytes())
		h.Lock()
		h.requestNonce[tag] = nonce
		h.Unlock()
		if v, ok := r.Context().Value(RequestNonceTags).([]string); ok {
			if !beStrings.StringInSlices(tag, v) {
				v = append(v, tag)
				modified = r.Clone(context.WithValue(r.Context(), RequestNonceTags, v))
			}
		} else {
			modified = r.Clone(context.WithValue(r.Context(), RequestNonceTags, []string{tag}))
		}
	}
	return
}

func (h *PolicyHandler) PrepareRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.PruneReportNonces()
		if strings.HasPrefix(r.URL.Path, "/_/csp-violation-") {
			pathLen := len(r.URL.Path)
			nonce := r.URL.Path[pathLen-10:]
			if h.ValidateReportNonce(nonce) {
				if r.Method == "POST" {
					body, _ := io.ReadAll(r.Body)
					log.WarnF("content-security-policy violation report received: %v", string(body))
				}
			} else {
				log.WarnF("ignoring csp report, invalid nonce present: %v", r.URL.Path)
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
				r = h.SetRequestPolicy(
					r,
					fn(
						h.GetRequestPolicy(r),
						r,
					),
				)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *PolicyHandler) ApplyHeaders(w http.ResponseWriter, r *http.Request) {
	if policy := h.GetRequestPolicy(r); policy != nil {
		if nonce, ok := r.Context().Value(ReportNonceTag).(string); ok {
			reportUri := DefaultReportPathPrefix + "-" + nonce
			value := policy.
				Set(NewReportUri(reportUri)).
				Set(NewReportTo("self-endpoint")).
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
	h.Lock()
	defer h.Unlock()
	if v, ok := r.Context().Value(RequestNonceTags).([]string); ok {
		for _, tag := range v {
			delete(h.requestNonce, tag)
		}
	}
	return
}
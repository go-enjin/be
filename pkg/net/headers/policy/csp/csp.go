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

	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/serve"
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
	PolicyTag      = "content-security-policy"
	ReportNonceTag = "content-security-policy-report-nonce"
)

var (
	DefaultReportPathPrefix = "/_/csp-violation"
)

type ModifyPolicyFn = func(policy Policy, r *http.Request) (modified Policy)

type PolicyHandler struct {
	nonces map[string]time.Time

	sync.RWMutex
}

func NewPolicyHandler() (h *PolicyHandler) {
	h = &PolicyHandler{
		nonces: make(map[string]time.Time),
	}
	return
}

func (h *PolicyHandler) Create() (nonce string) {
	h.Lock()
	defer h.Unlock()
	unique, _ := uuid.NewV4()
	nonce, _ = sha.DataHash10(unique.Bytes())
	h.nonces[nonce] = time.Now()
	return
}

func (h *PolicyHandler) Validate(nonce string) (valid bool) {
	h.RLock()
	defer h.RUnlock()
	_, valid = h.nonces[nonce]
	return
}

func (h *PolicyHandler) Prune() {
	h.Lock()
	defer h.Unlock()
	for nonce, stamp := range h.nonces {
		if time.Now().Sub(stamp) >= time.Minute {
			delete(h.nonces, nonce)
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

func (h *PolicyHandler) PrepareRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Prune()
		if strings.HasPrefix(r.URL.Path, "/_/csp-violation-") {
			pathLen := len(r.URL.Path)
			nonce := r.URL.Path[pathLen-10:]
			if h.Validate(nonce) {
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
		r = r.Clone(context.WithValue(r.Context(), ReportNonceTag, h.Create()))
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

func (h *PolicyHandler) FinalizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if policy := h.GetRequestPolicy(r); policy != nil {
			if nonce, ok := r.Context().Value(ReportNonceTag).(string); ok {
				reportUri := DefaultReportPathPrefix + "-" + nonce
				value := policy.
					Set(NewReportUri(reportUri)).
					Set(NewReportTo("self-endpoint")).
					Value()
				w.Header().Set("Reporting-Endpoints", `self-endpoint="`+reportUri+`"`)
				w.Header().Set("Content-Security-Policy", value)
			}
		}
		next.ServeHTTP(w, r)
	})
}
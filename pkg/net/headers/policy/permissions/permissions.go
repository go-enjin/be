// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-enjin/be/pkg/log"
)

func NoPermissionsPolicy() Policy {
	return NewPolicy(
		NewAccelerometer(AllowNone),
		NewAmbientLightSensor(AllowNone),
		NewAutoplay(AllowNone),
		NewBattery(AllowNone),
		NewCamera(AllowNone),
		NewDisplayCapture(AllowNone),
		NewDocumentDomain(AllowNone),
		NewEncryptedMedia(AllowNone),
		NewExecutionWhileNotRendered(AllowNone),
		NewHidden(AllowNone),
		NewExecutionWhileOutOfViewport(AllowNone),
		NewFullscreen(AllowNone),
		NewGamepad(AllowNone),
		NewGamepadconnected(AllowNone),
		NewGeolocation(AllowNone),
		NewGyroscope(AllowNone),
		NewHid(AllowNone),
		NewIdleDetection(AllowNone),
		NewLocalFonts(AllowNone),
		NewMagnetometer(AllowNone),
		NewMicrophone(AllowNone),
		NewMidi(AllowNone),
		NewPayment(AllowNone),
		NewPictureInPicture(AllowNone),
		NewPublickeyCredentialsGet(AllowNone),
		NewScreenWakeLock(AllowNone),
		NewSerial(AllowNone),
		NewSpeakerSelection(AllowNone),
		NewUsb(AllowNone),
		NewWebShare(AllowNone),
		NewXrSpatialTracking(AllowNone),
	)
}

const (
	PolicyTag = "permissions-policy"
)

type ModifyPolicyFn = func(policy Policy, r *http.Request) (modified Policy)

type PolicyHandler struct {
	sync.RWMutex
}

func NewPolicyHandler() (h *PolicyHandler) {
	h = &PolicyHandler{}
	return
}

func (h *PolicyHandler) SetRequestPolicy(r *http.Request, policy Policy) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), PolicyTag, policy))
	return
}

func (h *PolicyHandler) GetRequestPolicy(r *http.Request) (policy Policy) {
	if p, ok := r.Context().Value(PolicyTag).(Policy); ok {
		policy = p
	} else {
		policy = NoPermissionsPolicy()
	}
	return
}

func (h *PolicyHandler) PrepareRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = h.SetRequestPolicy(r, NoPermissionsPolicy())
		next.ServeHTTP(w, r)
	})
}

func (h *PolicyHandler) ModifyPolicyMiddleware(fn ModifyPolicyFn) (mw func(next http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		log.DebugF("including modify permissions policy middleware")
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
			w.Header().Set("Permissions-Policy", policy.Value())
		}
		next.ServeHTTP(w, r)
	})
}
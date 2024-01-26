// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"

	"github.com/go-enjin/golang-org-x-text/message"
)

var (
	ErrNotImplemented        = errors.New("not implemented")
	ErrSignalStopped         = errors.New("signal stopped action")
	ErrNothingToDo           = errors.New("nothing to do")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrBadCookie             = errors.New("bad cookie")
	ErrBadRequest            = errors.New("bad request")
	ErrExistsAlready         = errors.New("exists already")
	ErrFileNotFound          = errors.New("file not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrGroupNotFound         = errors.New("group not found")
	ErrTokenNotFound         = errors.New("token not found")
	ErrSecretNotFound        = errors.New("secret not found")
	ErrProviderNotFound      = errors.New("provider not found")
	ErrAudienceNotFound      = errors.New("audience not found")
	ErrFileSystemLockTimeout = errors.New("filesystem lock request timed out")
	ErrSuspiciousPanic       = errors.New("suspicious activity or possible programmer error")
	ErrDataTypeNotSupported  = errors.New("data type not supported")
)

func BadRequestError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("Bad request")
	return
}

func FormExpiredError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("Form submission expired, please try again")
	return
}

func IncompleteFormError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("Incomplete form submitted, please try again")
	return
}

func IncompleteLinkError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("Incomplete link submitted, please try again")
	return
}

func UnexpectedError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("An unexpected error occurred")
	return
}

func PermissionDeniedError(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("Permission denied")
	return
}

func OtpChallengeFailed(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("OTP (one-time passcode) verification failed")
	return
}

func UserNotFound(printer *message.Printer) (msg string) {
	msg = printer.Sprintf("User not found")
	return
}

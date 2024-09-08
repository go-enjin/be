//go:build driver_email_gomail || drivers_email || drivers || all

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

package gomail

type SmtpConfig struct {
	// Host is the remote SMTP service hostname
	Host string
	// Port is the remote SMTP service port
	Port int
	// Username is the remote SMTP service account username
	Username string
	// Username is the remote SMTP service account password
	//
	// Do not hard-code this value. Use the environment variables for operational
	// security
	Password string

	// Email is the address used in the `From:` message header
	Email string
	// Display is the (optional) Email address display name
	Display string

	// Retries is the maximum number of retries after if the first attempt fails
	//
	// Negative values use the DefaultRetries and a value of zero means "try once"
	Retries int
}

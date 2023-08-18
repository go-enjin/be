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

package userbase

import (
	"github.com/go-enjin/be/pkg/feature"
)

type GroupsProvider interface {
	// IsUserInGroup returns true if the user is in the given group
	IsUserInGroup(eid string, group Group) (present bool)

	// GetUserGroups returns the user's list of groups
	GetUserGroups(eid string) (groups Groups)
}

type GroupsManager interface {
	// GetGroupActions returns the list of actions associated with group
	GetGroupActions(group Group) (actions feature.Actions)

	// UpdateGroup appends the given actions to the group, creating the group
	// if non exists
	UpdateGroup(group Group, actions ...feature.Action) (err error)

	// RemoveGroup deletes the given group from the system. Does not delete any
	// fallback groups
	RemoveGroup(group Group) (err error)

	// AddUserToGroup adds a user to the list of groups given
	AddUserToGroup(eid string, groups ...Group) (err error)

	// RemoveUserFromGroup removes a user from each of the given groups
	RemoveUserFromGroup(eid string, groups ...Group) (err error)
}
// Copyright (C) 2023  RunesGambit.com - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package feature

type Database interface {
	Feature

	// ListDB returns a sorted list of connected db tags for use with DB and
	// MustDB
	ListDB() (tags []string)

	// DB returns the database connection or an error
	DB(tag string) (db interface{}, err error)

	// MustDB returns the database connection or panics on error
	MustDB(tag string) (db interface{})
}

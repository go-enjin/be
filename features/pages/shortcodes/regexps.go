package shortcodes

import (
	"regexp"
)

var (
	keywordPattern  = `[a-zA-Z][-_a-zA-Z0-9]*[a-zA-Z0-9]*`
	rxNameOnly      = regexp.MustCompile(`^\s*(` + keywordPattern + `)\s*$`)
	rxNameValue     = regexp.MustCompile(`^\s*(` + keywordPattern + `)=(.+?)\s*$`)
	rxNameKeyValues = regexp.MustCompile(`^\s*(` + keywordPattern + `)\s*(\s*.+?=.+?)\s*$`)
	rxSpaces        = regexp.MustCompile(`\s+`)
)
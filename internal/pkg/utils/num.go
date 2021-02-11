package utils

import (
	"regexp"
)

var (
	numberRegexp *regexp.Regexp
)

func init() {
	numberRegexp = regexp.MustCompile("^[-+−]?(([0-9]+([,\\.][0-9]+)*)|([0-9]+([\\.][0-9]+)*[eE][-+−][0-9]+))$")
}

//IsNumber test is string is number
func IsNumber(s string) bool {
	return numberRegexp.Match([]byte(s))
}

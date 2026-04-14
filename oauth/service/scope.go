package service

import (
	"strings"

	"github.com/gaucho-racing/sentinel/oauth/model"
)

func ValidateScopes(scopes string) bool {
	for _, scope := range strings.Fields(scopes) {
		if _, ok := model.ValidScopes[scope]; !ok {
			return false
		}
	}
	return true
}

func ScopesContain(scopes string, target string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == target {
			return true
		}
	}
	return false
}

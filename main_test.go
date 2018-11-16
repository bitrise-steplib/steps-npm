package main

import (
	"testing"
)

func TestExtractNpmVersion(t *testing.T) {

	testCases := []struct {
		pkgJSON string
		want string
	} {
		{`{"engines":{"npm":"3.0.1"}}`,  "3.0.1"},
		{`{"engines":{"node":"3.0.1"}}`, ""},
	}

	for _, tc := range testCases {
		got := extractNpmVersion(tc.pkgJSON)
		if got != tc.want {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned %s instead of %s`, tc.pkgJSON, got, tc.want)
		}
	}
	
}
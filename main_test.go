package main

import (
	"testing"
)

func TestExtractNpmVersion(t *testing.T) {

	testCases := []struct {
		pkgJSON string
		want    string
		hasE    bool
	}{
		{`{"engines":{"npm":"3.0.1"}}`, "3.0.1", false},
		{`"engines":{"npm":"3.0.1"}}`, "", true},
		{`{"engines":{}}`, "", true},
		{`{"engines":{"npm":"a.b.c"}}`, "", true},
	}

	for _, tc := range testCases {
		got, gotE := extractNpmVersion(tc.pkgJSON)
		if got != tc.want {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned %s instead of %s`, tc.pkgJSON, got, tc.want)
		}

		if !tc.hasE && gotE != nil {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned with error %s`, tc.pkgJSON, gotE)
		}
	}
}

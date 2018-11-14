package main

import (
	"testing"
)

func Test_GetNpmVersionFromPackageJson(t *testing.T) {

	testCases := []struct {
		pkgJson string
		want string
		wantE error
	} {
		{`{"engines":{"npm":"3.0.1"}}`, "3.0.1", nil},
		{`{"engines":{"node":"3.0.1"}}`, "", ErrMissingNpmVersion},
	}

	for _, tc := range testCases {
		got, gotE := getNpmVersionFromPackageJson(tc.pkgJson)
		if got != tc.want || gotE != tc.wantE {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned %s,%s instead of %s,%s`, tc.pkgJson, got, gotE, tc.want, tc.wantE)
		}
	}
	
}
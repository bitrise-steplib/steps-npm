package main

import (
	"errors"
	"testing"
)

func TestExtractNpmVersion(t *testing.T) {

	testCases := []struct {
		pkgJSON string
		want    string
		wantE   error
	}{
		{`{"engines":{"npm":"3.0.1"}}`, "3.0.1", nil},
		{`"engines":{"npm":"3.0.1"}}`, "", errors.New("")},
		{`{"engines":{}}`, "", errors.New("")},
		{`{"engines":{"npm":"a.b.c"}}`, "", errors.New("")},
	}

	for _, tc := range testCases {
		got, gotE := extractNpmVersion(tc.pkgJSON)
		if got != tc.want {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned %s instead of %s`, tc.pkgJSON, got, tc.want)
		}

		if tc.wantE == nil && gotE != nil {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned with error %s instead of %s`, tc.pkgJSON, gotE, tc.wantE)
		}
	}

}

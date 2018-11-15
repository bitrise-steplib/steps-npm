package main

import (
	"testing"
)

func TestExtractNpmVersion(t *testing.T) {

	testCases := []struct {
		pkgJson string
		want string
	} {
		{`{"engines":{"npm":"3.0.1"}}`,  "3.0.1"},
		{`{"engines":{"node":"3.0.1"}}`, ""},
	}

	for _, tc := range testCases {
		got := extractNpmVersion(tc.pkgJson)
		if got != tc.want {
			t.Errorf(`getNpmVersionFromPackageJson(%s) returned %s instead of %s`, tc.pkgJson, got, tc.want)
		}
	}
	
}

func TestGetCommandAsSliceForPlatform (t *testing.T) {

	// todo: figure out how to check for commands, not just errors
	testCases := []struct {
		platform string
		wantE error
	} {
		{`darwin`, error(nil)},
		{`linux`, error(nil)},
		{`windows`, ErrOsNotSupported},
	}

	for _, tc := range testCases {
		_, gotE := getCommandAsSliceForPlatform(tc.platform)
		if gotE != tc.wantE {
			t.Errorf(`getCommandAsSliceForPlatform(%s) returned %s instead of %s`, tc.platform, gotE, tc.wantE)
		}
	}

}
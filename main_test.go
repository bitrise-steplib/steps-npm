package main

import "testing"

const (
	CORRECT_PACKAGE_JSON = `
	{
		"engines": {
			"npm": "3.0.1"
		}
	}
	`
	MISSING_NPM_JSON = `
	{
		"engines": {
			"node": "6.0.0"
		}
	}
	`
)

func Test_GetNpmVersionFromPackageJson(t *testing.T) {
	correctVersion := getNpmVersionFromPackageJson(CORRECT_PACKAGE_JSON)

	if correctVersion != "3.0.1" {
		t.Fail()
	}
}
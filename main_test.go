package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
)

func TestResolveCorepackSetup(t *testing.T) {
	tests := []struct {
		name           string
		packageJSON    string // written to workdir/package.json if non-empty
		npmVersion     string // explicit step input
		wantEnable     bool
		wantPrepare    string
	}{
		{
			name:        "explicit npm_version activates prepare",
			npmVersion:  "9.8.1",
			wantEnable:  true,
			wantPrepare: "9.8.1",
		},
		{
			name:        "npm_version takes priority over packageManager field",
			packageJSON: `{"packageManager":"npm@10.2.0"}`,
			npmVersion:  "9.8.1",
			wantEnable:  true,
			wantPrepare: "9.8.1",
		},
		{
			name:        "packageManager npm enables corepack shim without prepare",
			packageJSON: `{"packageManager":"npm@10.2.0"}`,
			wantEnable:  true,
		},
		{
			name:        "packageManager yarn skips corepack",
			packageJSON: `{"packageManager":"yarn@3.0.0"}`,
		},
		{
			name:        "no packageManager field skips corepack",
			packageJSON: `{"name":"my-app","version":"1.0.0"}`,
		},
		{
			name: "no package.json skips corepack",
		},
		{
			name:        "invalid package.json skips corepack",
			packageJSON: `not valid json`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.packageJSON != "" {
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(tc.packageJSON), 0600); err != nil {
					t.Fatalf("failed to write package.json: %s", err)
				}
			}

			got := resolveCorepackSetup(dir, tc.npmVersion, log.NewLogger())

			if got.enable != tc.wantEnable {
				t.Errorf("enable: got %v, want %v", got.enable, tc.wantEnable)
			}
			if got.prepareVersion != tc.wantPrepare {
				t.Errorf("prepareVersion: got %q, want %q", got.prepareVersion, tc.wantPrepare)
			}
		})
	}
}

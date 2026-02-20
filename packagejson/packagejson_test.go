package packagejson

import (
	"testing"
)

func TestParsePackageManager(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "npm exact version",
			input:       `{"packageManager":"npm@9.0.0"}`,
			wantName:    "npm",
			wantVersion: "9.0.0",
		},
		{
			name:        "npm version with hash suffix",
			input:       `{"packageManager":"npm@9.0.0+sha224.abc"}`,
			wantName:    "npm",
			wantVersion: "9.0.0",
		},
		{
			name:        "yarn version",
			input:       `{"packageManager":"yarn@3.0.0"}`,
			wantName:    "yarn",
			wantVersion: "3.0.0",
		},
		{
			name:        "field not set",
			input:       `{}`,
			wantName:    "",
			wantVersion: "",
		},
		{
			name:    "missing @ separator",
			input:   `{"packageManager":"npm9"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotVersion, err := ParsePackageManager(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}

			if gotName != tc.wantName {
				t.Errorf("name: got %q, want %q", gotName, tc.wantName)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("version: got %q, want %q", gotVersion, tc.wantVersion)
			}
		})
	}
}

package corepack

import (
	"testing"
)

func Test_versionAtLeast(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		min         string
		wantResult  bool
		wantErr     bool
	}{
		{
			name:       "equal versions",
			version:    "0.31.0",
			min:        "0.31.0",
			wantResult: true,
		},
		{
			name:       "higher patch",
			version:    "0.31.1",
			min:        "0.31.0",
			wantResult: true,
		},
		{
			name:       "lower patch",
			version:    "0.30.9",
			min:        "0.31.0",
			wantResult: false,
		},
		{
			name:       "higher minor",
			version:    "0.32.0",
			min:        "0.31.0",
			wantResult: true,
		},
		{
			name:       "lower minor",
			version:    "0.29.0",
			min:        "0.31.0",
			wantResult: false,
		},
		{
			name:       "higher major",
			version:    "1.0.0",
			min:        "0.31.0",
			wantResult: true,
		},
		{
			name:       "lower major",
			version:    "0.31.0",
			min:        "1.0.0",
			wantResult: false,
		},
		{
			name:    "invalid version",
			version: "not-a-version",
			min:     "0.31.0",
			wantErr: true,
		},
		{
			name:    "invalid min",
			version: "0.31.0",
			min:     "not-a-version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := versionAtLeast(tt.version, tt.min)
			if (err != nil) != tt.wantErr {
				t.Errorf("versionAtLeast(%q, %q) error = %v, wantErr %v", tt.version, tt.min, err, tt.wantErr)
				return
			}
			if got != tt.wantResult {
				t.Errorf("versionAtLeast(%q, %q) = %v, want %v", tt.version, tt.min, got, tt.wantResult)
			}
		})
	}
}

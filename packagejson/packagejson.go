package packagejson

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParsePackageManager parses the packageManager field from package.json content.
// Returns (name, version, nil) or ("", "", nil) if the field is not set.
// The version hash suffix (e.g. "+sha224.abc") is stripped from the version.
// Example: "npm@9.0.0" → ("npm", "9.0.0", nil)
// Example: "npm@9.0.0+sha224.abc" → ("npm", "9.0.0", nil)
func ParsePackageManager(content string) (name, version string, err error) {
	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return "", "", fmt.Errorf("failed to parse package.json: %s", err)
	}

	if pkg.PackageManager == "" {
		return "", "", nil
	}

	name, versionWithHash, found := strings.Cut(pkg.PackageManager, "@")
	if !found || name == "" || versionWithHash == "" {
		return "", "", fmt.Errorf("invalid packageManager field %q: expected format \"<name>@<version>\"", pkg.PackageManager)
	}

	// Strip optional hash suffix (e.g. "+sha224.abc123")
	version, _, _ = strings.Cut(versionWithHash, "+")

	if version == "" {
		return "", "", fmt.Errorf("invalid packageManager field %q: expected format \"<name>@<version>\"", pkg.PackageManager)
	}

	return name, version, nil
}

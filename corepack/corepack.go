package corepack

import (
	"fmt"
	"os/exec"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
)

// EnsureInstalled installs corepack globally via npm if it's not already available.
func EnsureInstalled(cmdFactory command.Factory, logger log.Logger) error {
	if _, err := exec.LookPath("corepack"); err == nil {
		return nil
	}

	cmd := cmdFactory.Create("npm", []string{"install", "-g", "corepack"}, nil)
	logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("failed to install corepack: %s", out)
	}

	return nil
}

// Enable runs `corepack enable` to set up corepack shims.
func Enable(cmdFactory command.Factory, logger log.Logger) error {
	cmd := cmdFactory.Create("corepack", []string{"enable"}, nil)
	logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("corepack enable failed: %s", out)
	}

	return nil
}

// PrepareNpm runs `corepack prepare npm@<version> --activate` to download and
// activate the specified npm version as the global default.
func PrepareNpm(version string, cmdFactory command.Factory, logger log.Logger) error {
	cmd := cmdFactory.Create("corepack", []string{"prepare", fmt.Sprintf("npm@%s", version), "--activate"}, nil)
	logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return fmt.Errorf("corepack prepare failed: %s", out)
	}

	return nil
}

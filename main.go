package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-steplib/steps-npm/corepack"
	"github.com/bitrise-steplib/steps-npm/packagejson"
	"github.com/kballard/go-shellquote"
)

// Config model
type Config struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
}

type corepackSetup struct {
	// enable means corepack should be installed and its npm shim enabled.
	enable bool
	// prepareVersion, if non-empty, additionally pins this exact npm version via corepack prepare --activate.
	// Only set when npm_version input is provided explicitly.
	prepareVersion string
}

// resolveCorepackSetup decides what corepack operations to run based on the
// explicit npm_version input and the packageManager field in package.json.
func resolveCorepackSetup(workdir, npmVersion string, logger log.Logger) corepackSetup {
	if npmVersion != "" {
		logger.Infof("Setting npm version %s via corepack (explicit step input)", npmVersion)
		return corepackSetup{enable: true, prepareVersion: npmVersion}
	}

	packageJSONPath := filepath.Join(workdir, "package.json")
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warnf("No package.json found at %s, skipping npm version setup", packageJSONPath)
		} else {
			logger.Warnf("Failed to read package.json: %s", err)
		}
		return corepackSetup{}
	}

	name, version, err := packagejson.ParsePackageManager(string(content))
	if err != nil {
		logger.Warnf("Failed to parse packageManager from package.json: %s", err)
		return corepackSetup{}
	}

	if name == "" {
		logger.Warnf("No packageManager field found in package.json.\n" +
			"To pin the npm version, add a \"packageManager\" field to package.json, e.g.:\n" +
			"  \"packageManager\": \"npm@10.2.0\"")
		return corepackSetup{}
	}

	if name != "npm" {
		logger.Warnf("packageManager is set to %q, not npm -> skipping npm version pinning", name+"@"+version)
		return corepackSetup{}
	}

	logger.Infof("Found packageManager field in package.json: npm@%s -> enabling corepack shim", version)
	return corepackSetup{enable: true}
}

func failf(logger log.Logger, f string, args ...any) {
	logger.Errorf(f, args...)
	os.Exit(1)
}

func main() {
	envRepo := env.NewRepository()
	logger := log.NewLogger()
	cmdFactory := command.NewFactory(envRepo)

	var config Config
	parser := stepconf.NewInputParser(envRepo)
	if err := parser.Parse(&config); err != nil {
		failf(logger, "Process config: %s", err)
	}
	stepconf.Print(config)

	workdir, err := normalizeWorkdir(config.Workdir)
	if err != nil {
		failf(logger, "Process config: %s", err)
	}

	npmArgs, err := shellquote.Split(config.Command)
	if err != nil {
		failf(logger, "Process config: provided npm command/arguments is not a valid CLI command: %s", err)
	}

	if strings.HasPrefix(config.Command, "install") {
		logger.Donef("\n" +
			"Info: From npm version >= v5.7.0, you can use the `npm ci` command instead of `npm install`. Using this command might speed up your workflow.\n" +
			"It does not work without `package-lock.json` so please commit it into the VCS repository. " +
			"More info: https://github.com/npm/npm/releases/tag/v5.7.0")
	}

	fmt.Println()
	setup := resolveCorepackSetup(workdir, config.NpmVersion, logger)
	if setup.enable {
		if err := corepack.EnsureInstalled(cmdFactory, logger); err != nil {
			failf(logger, "Version setup: %s", err)
		}
		if err := corepack.Enable(cmdFactory, logger); err != nil {
			failf(logger, "Version setup: %s", err)
		}
	}
	if setup.prepareVersion != "" {
		if err := corepack.PrepareNpm(setup.prepareVersion, cmdFactory, logger); err != nil {
			failf(logger, "Version setup: failed to prepare npm@%s: %s", setup.prepareVersion, err)
		}
	}

	fmt.Println()
	logger.Infof("Running user provided command")

	cmd := cmdFactory.Create("npm", npmArgs, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    workdir,
	})
	logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		failf(logger, "Run: provided npm command failed: %s", err)
	}

	fmt.Println()
	logger.Donef("Step success")
}

func normalizeWorkdir(workdir string) (string, error) {
	if workdir == "" {
		return os.Getwd()
	}

	abs, err := filepath.Abs(workdir)
	if err != nil {
		return "", fmt.Errorf("failed to normalize working directory path: %s", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("specified working directory path %q does not exist", abs)
		}
		return "", fmt.Errorf("failed to validate working directory path %q: %s", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("working directory path %q is not a directory", abs)
	}

	return abs, nil
}

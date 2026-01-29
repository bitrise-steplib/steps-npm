package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	semver "github.com/hashicorp/go-version"
	"github.com/kballard/go-shellquote"
)

// Config model
type Config struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
}

func getNpmVersionFromPackageJSON(path string) (string, error) {
	jsonStr, err := fileutil.ReadStringFromFile(path)
	if err != nil {
		return "", fmt.Errorf("package.json file read error: %s", err)
	}

	ver, err := extractNpmVersion(jsonStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse package.json: %s", err)
	}
	return ver, nil
}

func extractNpmVersion(jsonStr string) (string, error) {
	type pkgJSON struct {
		Engines struct {
			Npm string
		}
	}

	var m pkgJSON
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err)
	}

	if m.Engines.Npm == "" {
		return "", nil
	}

	v, err := semver.NewVersion(m.Engines.Npm)
	if err != nil {
		return "", fmt.Errorf("`%s` is not valid semver string: %s", m.Engines.Npm, err)
	}

	return v.String(), nil
}

func createInstallNpmCommand(cmdFactory command.Factory) (command.Command, error) {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name = "brew"
		args = []string{"install", "node"}
	case "linux":
		name = "apt-get"
		args = []string{"-y", "install", "npm"}
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmdFactory.Create(name, args, nil), nil
}

func setNpmVersion(ver string, cmdFactory command.Factory, logger log.Logger) error {
	cmd := cmdFactory.Create("npm", []string{"install", "-g", "--force", fmt.Sprintf("npm@%s", ver)}, nil)
	logger.Donef(fmt.Sprintf("$ %s", cmd.PrintableCommandArgs()))
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("npm command failed: %s", out)
		}
		return fmt.Errorf("error running npm command: %s", err)
	}

	return nil
}

func systemDefined(cmdFactory command.Factory, logger log.Logger) (string, error) {
	if path, err := exec.LookPath("npm"); err == nil {
		logger.Printf("npm found at %s", path)

		cmd := cmdFactory.Create("npm", []string{"--version"}, nil)
		logger.Donef(fmt.Sprintf("$ %s", cmd.PrintableCommandArgs()))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				return "", fmt.Errorf("npm command failed: %s", out)
			}
			return "", fmt.Errorf("error running npm command: %s", err)
		}

		return out, nil
	}

	return "", nil
}

func failf(logger log.Logger, f string, args ...interface{}) {
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

	workdir, err := pathutil.AbsPath(config.Workdir)
	if err != nil {
		failf(logger, "Process config: failed to normalize working directory path: %s", err)
	}

	exists, err := pathutil.IsDirExists(workdir)
	if err != nil {
		failf(logger, "Process config: failed to validate working directory path `%s`: %s", workdir, err)
	}
	if !exists {
		failf(logger, "Process config: specified working directory path `%s` does not exist", workdir)
	}

	npmArgs, err := shellquote.Split(config.Command)
	if err != nil {
		failf(logger, "Process config: provided npm command/arguments is not a valid CLI command: %s", err)
	}

	if strings.HasPrefix(config.Command, "install") {
		logger.Donef("\n" +
			"Info: From npm version >= v5.7.0, you can use the `npm ci` command insead of `npm install`. Using this command might speeds up your workflow.\n" +
			"It does not work without `package-lock.json` so please commit it into the VCS repository. " +
			"More info: https://github.com/npm/npm/releases/tag/v5.7.0")
	}

	toInstall := false
	toSet := config.NpmVersion

	if toSet == "" {
		fmt.Println()
		logger.Infof("Autodetecting npm version")
		logger.Printf("Checking package.json for npm version")

		path := filepath.Join(workdir, "package.json")
		exists, err := pathutil.IsPathExists(path)
		if err != nil {
			failf(logger, "Install dependencies: failed to validate package.json path: %s", err)
		}

		if exists {
			toSet, err = getNpmVersionFromPackageJSON(path)
			if err != nil {
				logger.Warnf("error getting version: %s", err)
			}
		} else {
			logger.Warnf("No package.json found at path: %s", path)
		}
	}

	if toSet == "" {
		logger.Warnf("Could not read version information from package.json")
		logger.Printf("Locating preinstalled npm")

		systemVer, err := systemDefined(cmdFactory, logger)
		if err != nil {
			failf(logger, "Install dependencies: failed to check installed npm version: %s", err)
		}
		if systemVer == "" {
			logger.Warnf("npm not found on PATH")
			toSet = "latest"
			toInstall = true
		}
		logger.Printf("Preinstalled npm version: %s", systemVer)
	}

	if toInstall {
		fmt.Println()
		logger.Infof("Ensuring npm version %s", toSet)

		cmd, err := createInstallNpmCommand(cmdFactory)
		if err != nil {
			failf(logger, "Install dependencies: %s", err)
		}
		logger.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf(logger, "Install dependencies: failed to install npm: %s", err)
		}
	}

	if toSet != "" {
		fmt.Println()
		logger.Infof("Ensuring npm version %s", toSet)

		if err := setNpmVersion(toSet, cmdFactory, logger); err != nil {
			failf(logger, "Install dependencies: failed to install npm version `%s`: %s", toSet, err)
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

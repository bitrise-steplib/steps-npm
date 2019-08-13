package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	semver "github.com/hashicorp/go-version"
	"github.com/kballard/go-shellquote"
)

// Config model
type Config struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
	UseCache   bool   `env:"cache_local_deps,opt[true,false]"`
}

func getNpmVersionFromPackageJSON(path string) (string, error) {
	jsonStr, err := fileutil.ReadStringFromFile(path)
	if err != nil {
		return "", fmt.Errorf("package.json file read error: %s", err)
	}

	ver, err := extractNpmVersion(jsonStr)
	if err != nil {
		return "", fmt.Errorf("package json content error: %s", err)
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

func createInstallNpmCommand() (*command.Model, error) {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"brew", "install", "node"}
	case "linux":
		args = []string{"apt-get", "-y", "install", "npm"}
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return command.New(args[0], args[1:]...), nil
}

func setNpmVersion(ver string) error {
	cmd := command.New("npm", "install", "-g", fmt.Sprintf("npm@%s", ver))
	log.Donef(fmt.Sprintf("$ %s", cmd.PrintableCommandArgs()))
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("error running npm install: %s", out)
		}
		return fmt.Errorf("error running npm install: %s", err)
	}

	return nil
}

func systemDefined() (string, error) {
	if path, err := exec.LookPath("npm"); err == nil {
		log.Printf("npm found at %s", path)

		cmd := command.New("npm", "--version")
		log.Donef(fmt.Sprintf("$ %s", cmd.PrintableCommandArgs()))
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				return "", fmt.Errorf("error running npm command: %s", out)
			}
			return "", fmt.Errorf("error running npm command: %s", err)
		}

		return out, nil
	}

	return "", nil
}

func failf(f string, args ...interface{}) {
	log.Errorf(f, args...)
	fmt.Println()
	os.Exit(1)
}

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		failf("error parsing step config: %s", err)
	}
	stepconf.Print(config)

	workdir, err := pathutil.AbsPath(config.Workdir)
	if err != nil {
		failf("error normalizing workdir path: %s", err)
	}

	exists, err := pathutil.IsDirExists(workdir)
	if err != nil {
		failf("error validating workdir `%s`: %s", workdir, err)
	}
	if !exists {
		failf("specified path `%s` does not exist", workdir)
	}

	toInstall := false
	toSet := config.NpmVersion

	if toSet == "" {
		fmt.Println()
		log.Infof("Autodetecting npm version")
		log.Printf("Checking package.json for npm version")

		path := filepath.Join(workdir, "package.json")
		exists, err := pathutil.IsPathExists(path)
		if err != nil {
			failf("error validating package.json path: %s", err)
		}

		if exists {
			toSet, err = getNpmVersionFromPackageJSON(path)
			if err != nil {
				log.Warnf("error getting version: %s", err)
			}
		} else {
			log.Warnf("No package.json found at path: %s", path)
		}
	}

	if toSet == "" {
		log.Warnf("Could not read version information from package.json")
		log.Printf("Locating system installed npm")

		systemVer, err := systemDefined()
		if err != nil {
			failf("error getting installed npm version: %s", err)
		}
		if systemVer == "" {
			log.Warnf("npm not found on PATH")
			toSet = "latest"
			toInstall = true
		}
		log.Printf("system npm version: %s", systemVer)
	}

	if toInstall {
		fmt.Println()
		log.Infof("Ensuring npm version %s", toSet)

		cmd, err := createInstallNpmCommand()
		if err != nil {
			failf("Error installing npm: %s", err)
		}
		log.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Error installing npm: %s", err)
		}
	}

	if toSet != "" {
		fmt.Println()
		log.Infof("Ensuring npm version %s", toSet)

		if err := setNpmVersion(toSet); err != nil {
			failf("Error setting npm version to %s: %s", toSet, err)
		}
	}

	fmt.Println()
	log.Infof("Running user provided command")

	args, err := shellquote.Split(config.Command)
	if err != nil {
		failf("error preparing command for execution: %s", err)
	}

	cmd := command.NewWithStandardOuts("npm", args...)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	cmd.SetDir(workdir)
	if err := cmd.Run(); err != nil {
		failf("Error running command %s: %s", config.Command, err)
	}

	// Only cache if npm command is install, node_modules could be included in the repository
	// Expecting command as the first argument of npm
	// npm commands: https://github.com/npm/cli/blob/36682d4482cddee0acc55e8d75b3bee6e78fff37/lib/config/cmd-list.js
	if config.UseCache &&
		(len(args) != 0) && sliceutil.IsStringInSlice(args[0], []string{"install", "isntall", "i", "add"}) {
		if err := cacheNpm(workdir); err != nil {
			log.Warnf("Failed to mark files for caching, error: %s", err)
		}
	}

	fmt.Println()
	log.Successf("Step success")
}

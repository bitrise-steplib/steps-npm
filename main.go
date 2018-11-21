package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/stepconf"
	semver "github.com/hashicorp/go-version"
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

func runAndLog(cmd *command.Model) (string, error) {
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	log.Donef(fmt.Sprintf("$ %s", cmd.PrintableCommandArgs()))
	if err != nil {
		return out, fmt.Errorf("error running npm command: %s", err)
	}

	log.Printf(out)

	return out, nil
}

func setNpmVersion(ver string) (string, error) {
	cmd := command.New("npm", "install", "-g", fmt.Sprintf("npm@%s", ver))
	out, err := runAndLog(cmd)
	if err != nil {
		return out, fmt.Errorf("error running npm install: %s", err)
	}

	return out, nil
}

func readFromPackageJSON(workdir string) (string, error) {
	path := filepath.Join(workdir, "package.json")

	return getNpmVersionFromPackageJSON(path)
}

func systemDefined() (string, error) {
	if path, err := exec.LookPath("npm"); err == nil {
		log.Printf("npm found at %s", path)

		cmd := command.New("npm", "--version")
		return runAndLog(cmd)
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
		toSet, err = readFromPackageJSON(workdir)
		if err != nil {
			failf("error reading npm version from package.json: %s", err)
		}
	}

	if toSet == "" {
		log.Warnf("No npm version found in package.json")
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
	}

	if toInstall {
		fmt.Println()
		log.Infof("Ensuring npm version %s", toSet)

		cmd, err := createInstallNpmCommand()
		if err != nil {
			failf("Error installing npm: %s", err)
		}
		if _, err := runAndLog(cmd); err != nil {
			failf("Error installing npm: %s", err)
		}
	}

	if toSet != "" {
		fmt.Println()
		log.Infof("Ensuring npm version %s", toSet)

		if _, err := setNpmVersion(toSet); err != nil {
			failf("Error setting npm version to %s: %s", toSet, err)
		}
	}

	fmt.Println()
	log.Infof("Running user provided command")

	cmd := command.New("npm", config.Command)
	cmd.SetDir(workdir)
	if _, err := runAndLog(cmd); err != nil {
		failf("Error running command %s: %s", config.Command, err)
	}

	fmt.Println()
	log.Successf("Step success")
}

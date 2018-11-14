package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"os"
	"os/exec"
	"runtime"
)

var ErrMissingNpmVersion = errors.New("Missing npm version constraint in package.json")
var ErrOsNotSupported = errors.New(fmt.Sprintf(`Operating system %s not supported`, runtime.GOOS))

type jsonModel struct {
	Engines struct {
		Npm string
	}
}

type ConfigsModel struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
}

func getNpmVersionFromPackageJson(content string) (string, error) {
	var m jsonModel
	_ = json.Unmarshal([]byte(content), &m)
	if m.Engines.Npm == "" {
		return "", ErrMissingNpmVersion
	}

	return m.Engines.Npm, nil
}

func getNpmVersionFromSystem() string {
	out, _ := command.RunCmdAndReturnTrimmedOutput(command.New("npm", "--version").GetCmd())
	return out
}

func getCommandAsSliceForPlatform(os string) ([]string, error) {
	var args []string
	switch os {
	case "darwin":
		args = []string{"brew", "install", "node"}
	case "linux":
		args = []string{"apt-get", "-y", "install", "npm"}
	default:
		return args, ErrOsNotSupported
	}
	return args, nil
}

func createInstallNpmCommand(platform string) *command.Model {
	slice, _ := getCommandAsSliceForPlatform(platform)
	cmd, _ := command.NewFromSlice(slice)
	return cmd
}

func installLatestNpm() error {
	fmt.Printf("INFO: npm binary not found on PATH, installing latest")
	
	installNpmCmd := createInstallNpmCommand(runtime.GOOS)
	if installNpmCmd == nil {
		return errors.New("FATAL ERROR: not supported OS version")
	}
	_, err := command.RunCmdAndReturnTrimmedOutput(installNpmCmd.GetCmd())
	return err
}

func failf(f string, args ...interface{}) {
	log.Errorf(f, args...)
	os.Exit(1)
}

func main() {
	var config ConfigsModel
	if err := stepconf.Parse(&config); err != nil {
		failf("Couldn't create step config: %v\n", err)
	}

	if config.NpmVersion == "" {
		content, err := fileutil.ReadStringFromFile("package.json")
		config.NpmVersion, err = getNpmVersionFromPackageJson(content)
		if config.NpmVersion == "" {
			if _, err = exec.LookPath("npm"); err == nil {
				config.NpmVersion = getNpmVersionFromSystem()

			} else {
				err := installLatestNpm()

				if err != nil {
					failf("Couldn't install npm: %v", err)
				}
			}
		}
	}

	fmt.Printf("detected npm version: %s\n", config.NpmVersion)
}

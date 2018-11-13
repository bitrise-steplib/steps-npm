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

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		Workdir:    os.Getenv("workdir"),
		Command:    os.Getenv("command"),
		NpmVersion: os.Getenv("npm_version"),
	}
}

func getNpmVersionFromPackageJson(path string) string {
	content, _ := fileutil.ReadStringFromFile("package.json")
	var m jsonModel
	_ = json.Unmarshal([]byte(content), &m)
	return m.Engines.Npm
}

func getNpmVersionFromSystem() string {
	out, _ := command.RunCmdAndReturnTrimmedOutput(command.New("npm", "--version").GetCmd())
	return out
}

func getCommandForPlatform(os string) *command.Model {
	var cmd *command.Model
	switch os {
	case "darwin":
		cmd = command.New("brew", "install", "node")
	case "linux":
		cmd = command.New("apt-get", "-y", "install", "npm")
	default:
		return nil
	}
	return cmd
}

func installLatestNpm() error {
	fmt.Printf("INFO: npm binary not found on PATH, installing latest")
	
	installNpmCmd := getCommandForPlatform(runtime.GOOS).GetCmd()
	if installNpmCmd == nil {
		return errors.New("FATAL ERROR: not supported OS version")
	}
	_, err := command.RunCmdAndReturnTrimmedOutput(installNpmCmd)
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
		config.NpmVersion = getNpmVersionFromPackageJson("package.json")
		if config.NpmVersion == "" {
			if _, err := exec.LookPath("npm"); err == nil {
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

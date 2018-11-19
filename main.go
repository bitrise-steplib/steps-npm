package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

// Config model
type Config struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
}

func getNpmVersionFromPackageJSON() (string, error) {
	jsonStr, err := fileutil.ReadStringFromFile("package.json")
	if err != nil {
		return "", fmt.Errorf("package.json file read error: %s", err)
	}

	var ver string
	if ver, err = extractNpmVersion(jsonStr); err != nil {
		return "", fmt.Errorf("package json content error: %s", err)
	}
	return ver, nil
}

func extractNpmVersion(jsonStr string) (string, error) {
	type jsonModel struct {
		Engines struct {
			Npm string
		}
	}

	var m jsonModel
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return "", fmt.Errorf("json unmarshal error: %s", err)
	}

	return m.Engines.Npm, nil
}

func createInstallNpmCommand(os string) (*command.Model, error) {
	var args []string
	switch os {
	case "darwin":
		args = []string{"brew", "install", "node"}
	case "linux":
		args = []string{"apt-get", "-y", "install", "npm"}
	}

	var cmd, err = command.NewFromSlice(args)
	if err != nil {
		return nil, fmt.Errorf("could not create npm install command: %s", err)
	}

	return cmd, nil
}

func installLatestNpm() (string, error) {
	cmd, err := createInstallNpmCommand(runtime.GOOS)
	if err != nil {
		return "", fmt.Errorf("error creating npm install command: %s", err)
	}

	var out string
	out, err = command.RunCmdAndReturnTrimmedOutput(cmd.GetCmd())

	if err != nil {
		return out, fmt.Errorf("error running npm install: %s", err)
	}

	return out, nil
}

func failf(f string, args ...interface{}) {
	log.Errorf(f, args...)
	os.Exit(1)
}

func runNpmCommand(npmCmd ...string) (string, error) {
	cmd := command.New("npm", npmCmd...)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	log.Infof(cmd.PrintableCommandArgs())
	if err != nil {
		return out, fmt.Errorf("error running npm command: %s", err)
	}
	return out, nil
}

func main() {
	var (
		config Config
		ver    string
		out    string
		err    error
	)

	if err := stepconf.Parse(&config); err != nil {
		failf("Couldn't create step config: %v\n", err)
	}
	stepconf.Print(config)

	ver = config.NpmVersion
	if ver == "" {
		log.Infof("No npm version provided as step input. Checking package.json.")
		if ver, err = getNpmVersionFromPackageJSON(); err != nil {
			log.Warnf("No npm version found in package.json! Falling back to installed npm.")
			if path, err := exec.LookPath("npm"); err != nil {
				log.Warnf("npm not found on PATH, installing latest version")
				if err := installLatestNpm(); err != nil {
					failf("Couldn't install npm: %v", err)
				}
				log.Infof("installing npm done")
			} else {
				log.Infof("using npm installation located at %s", path)
			}
		}
	}

	out, err = runNpmCommand("install", "-g", fmt.Sprintf("npm@%s", ver))
	if err != nil {
		log.Errorf(out)
		failf("error setting npm version %s: %s", ver, err)
	}

	out, err = runNpmCommand(config.Command)
	if err != nil {
		log.Errorf(out)
		failf("error running npm command %s: %s", config.Command, err)
	}

	log.Donef("$ npm %s", config.Command)
	log.Infof("npm %s output: ", out)
	log.Successf("Step success")
}

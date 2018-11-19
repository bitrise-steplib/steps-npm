package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
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
	return command.NewFromSlice(args)
}

func installLatestNpm() error {
	installNpmCmd, err := createInstallNpmCommand(runtime.GOOS)
	if err != nil {
		return err
	}

	// return the output also -- or put the relevant error message in the returned error
	// consider the exit status handling!
	_, err = command.RunCmdAndReturnTrimmedOutput(installNpmCmd.GetCmd())
	return err
}

func failf(f string, args ...interface{}) {
	log.Errorf(f, args...)
	os.Exit(1)
}

func runNpmCommand(npmCmd ...string) (out string, exitCode int, err error) {
	cmd := command.New("npm", npmCmd...)
	out, npmErr := cmd.RunAndReturnTrimmedCombinedOutput()
	log.Infof(cmd.PrintableCommandArgs())
	if npmErr != nil {
		if errorutil.IsExitStatusError(npmErr) {
			// simplify error handling -- remove exit code stuff
			exitCode, err := errorutil.CmdExitCodeFromError(npmErr)
			if err != nil {
				return out, 1, err
			}

			return out, exitCode, errors.New(out)
		}
		return out, 0, npmErr
	}
	return out, 0, nil
}

func main() {
	var (
		config   Config
		ver      string
		out      string
		exitCode int // we dont need it
		err      error
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

	// do not expose exit code in main and in log messages
	out, exitCode, err = runNpmCommand("install", "-g", fmt.Sprintf("npm@%s", ver))

	if exitCode != 0 {
		failf("npm exit code %s, error: %s", exitCode, out)
	}
	if err != nil {
		failf("could not set npm to version %s: %s", ver, err)
	}
	log.Infof("npm update output: %s", out)

	out, exitCode, err = runNpmCommand(config.Command)
	if exitCode != 0 {
		failf("npm exit code %s, error: %s", exitCode, err)
	}

	log.Donef("$ npm %s", config.Command)
	log.Infof("npm %s output: ", out)
	log.Successf("Step success")
}

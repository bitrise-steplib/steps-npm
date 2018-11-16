package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
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

type Config struct {
	Workdir    string `env:"workdir"`
	Command    string `env:"command,required"`
	NpmVersion string `env:"npm_version"`
}

func getNpmVersionFromPackageJson() string {
	jsonStr, err := fileutil.ReadStringFromFile("package.json")
	if err != nil {
		failf("No package.json file found", err)
	}

	return extractNpmVersion(jsonStr)
}

func extractNpmVersion(jsonStr string)  string {
	var m jsonModel
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return ""
	}

	return m.Engines.Npm
}

func getNpmVersionFromSystem() string {
	out, _ := command.RunCmdAndReturnTrimmedOutput(command.New("npm", "--version").GetCmd())
	return out
}

func getCommandAsSliceForPlatform(os string) []string {
	var args []string
	switch os {
	case "darwin":
		args = []string{"brew", "install", "node"}
	case "linux":
		args = []string{"apt-get", "-y", "install", "npm"}
	}
	return args
}

func createInstallNpmCommand(platform string) *command.Model {
	slice := getCommandAsSliceForPlatform(platform)
	cmd, _ := command.NewFromSlice(slice)
	return cmd
}

func installLatestNpm() error {
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

func (configs Config) print() {
	fmt.Println()
	log.Infof("Configs:")
	log.Printf(" - Workdir: %s", configs.Workdir)
	log.Printf(" - Command: %s", configs.Command)
	log.Printf(" - NpmVerion: %s", configs.NpmVersion)
	fmt.Println()
}

func setNpmVersion(ver string) {
	// todo: run npm update to specific version
	// todo: log command and output
	runNpmCommand("npm run " + ver)
}

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		failf("Couldn't create step config: %v\n", err)
	}
	config.print()

	if config.NpmVersion == "" {
		log.Infof("No npm version provided as step input. Checking package.json.")

		if ver := getNpmVersionFromPackageJson(); ver == "" {
			log.Warnf("No npm version found in package.json! Falling back to installed npm.")
			if _, err := exec.LookPath("npm"); err != nil {
				log.Warnf("npm not found on PATH, installing latest version")
				if err := installLatestNpm(); err != nil {
					failf("Couldn't install npm: %v", err)
				}
			} else {
				ver = getNpmVersionFromSystem()
			}

		} else {
			setNpmVersion(ver)
		}
	} else {
		setNpmVersion(config.NpmVersion)
	}

	cmd := command.New("npm", config.Command)
	out, npmErr := cmd.RunAndReturnTrimmedCombinedOutput()
	if npmErr != nil {
		if errorutil.IsExitStatusError(npmErr) {
			exitCode, err := errorutil.CmdExitCodeFromError(npmErr)

			failf("npm exit code %s, %s", exitCode, npmErr)
		} else {
			failf("%s failed: %s", cmd.PrintableCommandArgs(), out)
		}
	}

	log.Donef("Step success")
}

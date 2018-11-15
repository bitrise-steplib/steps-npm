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

func getNpmVersionFromPackageJson(content string) string {
	var m jsonModel
	err := json.Unmarshal([]byte(content), &m)
	if err != nil {
		return ""
	}

	return m.Engines.Npm
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

func (configs ConfigsModel) print() {
	fmt.Println()
	log.Infof("Configs:")
	log.Printf(" - Workdir: %s", configs.Workdir)
	log.Printf(" - Command: %s", configs.Command)
	log.Printf(" - NpmVerion: %s", configs.NpmVersion)
	fmt.Println()
}

func main() {
	var config ConfigsModel
	if err := stepconf.Parse(&config); err != nil {
		failf("Couldn't create step config: %v\n", err)
	}
	config.print()

	if config.NpmVersion == "" {
		content, err := fileutil.ReadStringFromFile("package.json")
		if err != nil {
			failf("No package.json file found", err)
		}

		if ver := getNpmVersionFromPackageJson(content); ver == "" {
			if path, _ := exec.LookPath("npm"); path == "" {
				config.NpmVersion = getNpmVersionFromSystem()
			} else {
				if err := installLatestNpm(); err != nil {
					failf("Couldn't install npm: %v", err)
				}
			}
		}
	}

	fmt.Printf("detected npm version: %s\n", config.NpmVersion)
}

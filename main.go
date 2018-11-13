package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
)

type jsonModel struct {
	Engines struct {
		Npm string
	}
}

type ConfigsModel struct {
	Workdir    string
	Command    string
	NpmVersion string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		Workdir:    os.Getenv("workdir"),
		Command:    os.Getenv("command"),
		NpmVersion: os.Getenv("npm_version"),
	}
}

func getNpmVersionFromPackageJson(path string) (string) {
	content, _ := fileutil.ReadStringFromFile("package.json")
	var m jsonModel;
	_ = json.Unmarshal([]byte(content), &m)
	return m.Engines.Npm
}

func getNpmVersionFromSystem() string {
	out, _ := command.RunCmdAndReturnTrimmedOutput(command.New("npm", "--version").GetCmd())
	return out
}

func main() {
	
	config, configSource := createConfigsModelFromEnvs(), "USERINPUT"
	if config.NpmVersion == "" {
		config.NpmVersion, configSource = getNpmVersionFromPackageJson("package.json"), "PACKAGEJSON"
		if config.NpmVersion == "" {
			if _, err := exec.LookPath("npm2"); err == nil {
				config.NpmVersion, configSource = getNpmVersionFromSystem(), "SYSTEM"

			} else {
				fmt.Printf("INFO: npm binary not found on PATH, installing latest")

			var cmd *command.Model
			switch (runtime.GOOS) {
			case "darwin":
				cmd = command.New("brew", "install", "node")
			case "linux":
				cmd = command.New("apt-get", "-y", "install", "npm")
			default:
				fmt.Printf("FATAL ERROR: not supported OS version")
				return
			}

				_, err := command.RunCmdAndReturnTrimmedOutput(cmd.GetCmd())
				configSource = "NONE"

				if err != nil {
					fmt.Printf("FATAL ERROR: %s\n", err)
					return
				}
			}
		}
	}

	fmt.Printf("detected npm version: %s using config source: %s\n", config.NpmVersion, configSource)
}
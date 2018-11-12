package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	_, error := exec.LookPath("npm")

	if (error != nil) {

	}

	out, _ := command.RunCmdAndReturnTrimmedOutput(command.New("npm", "--version").GetCmd())
	return out
}

func main() {
	
	config := createConfigsModelFromEnvs()
	if config.NpmVersion == "" {
		config.NpmVersion = getNpmVersionFromPackageJson("package.json")
		if config.NpmVersion == "" {
			config.NpmVersion = getNpmVersionFromSystem()
		}
	}

	fmt.Printf("detected npm version: %s\n", config.NpmVersion)
}
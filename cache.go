package main

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type cacheLevel int

const (
	cacheNone cacheLevel = iota
	cacheLocal
	cacheGlobal
)

func cacheNpm(workdir string, cacheLevel cacheLevel) error {
	npmCache := cache.New()
	switch cacheLevel {
	case cacheNone:
		return nil
	case cacheLocal:
		{
			// Cache local node_modules
			localPackageDir := filepath.Join(workdir, "node_modules")

			exist, err := pathutil.IsDirExists(localPackageDir)
			if err != nil {
				return fmt.Errorf("failed to check directory existence, error: %s", err)
			}
			if exist {
				npmCache.IncludePath(localPackageDir)
			} else {
				log.Debugf("local node_modules directory does not exist: %s", localPackageDir)
			}
		}
	case cacheGlobal:
		{
			npmInstallCommand := command.New("npm", "root", "-g")
			fmt.Println()
			log.Donef("$ %s", npmInstallCommand.PrintableCommandArgs())

			out, err := npmInstallCommand.RunAndReturnTrimmedOutput()
			if err != nil {
				if errorutil.IsExitStatusError(err) {
					return fmt.Errorf("command failed, output: %s", out)
				}
				return fmt.Errorf("failed to run command, error: %s", err)
			}

			exist, err := pathutil.IsDirExists(out)
			if err != nil {
				return fmt.Errorf("failed to check directory existence, error: %s", err)
			}
			if exist {
				npmCache.IncludePath(out)
			} else {
				log.Debugf("Global npm package directory does not exist: %s", err)
			}
		}
	}

	if err := npmCache.Commit(); err != nil {
		return fmt.Errorf("failed to mark node_modules directory to be cached, error: %s", err)
	}
	return nil
}

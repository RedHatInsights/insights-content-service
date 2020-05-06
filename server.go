/*
Copyright © 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Entry point to the insights report server
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/conf"
)

const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota

	defaultConfigFilename = "config"
)

var (
	// BuildVersion contains the major.minor version of the CLI client
	BuildVersion string = "*not set*"

	// BuildTime contains timestamp when the CLI client has been built
	BuildTime string = "*not set*"

	// BuildBranch contains Git branch used to build this application
	BuildBranch string = "*not set*"

	// BuildCommit contains Git commit used to build this application
	BuildCommit string = "*not set*"
)

// startService starts service and returns error code
func startService() int {
	return ExitStatusOK
}

func initInfoLog(msg string) {
	log.Info().Str("type", "init").Msg(msg)
}

func printVersionInfo() {
	initInfoLog("Version: " + BuildVersion)
	initInfoLog("Build time: " + BuildTime)
	initInfoLog("Branch: " + BuildBranch)
	initInfoLog("Commit: " + BuildCommit)
}

const helpMessageTemplate = `
Reporting service for insights results

Usage:

    %+v [command]

The commands are:

    <EMPTY>             starts reporting service
    start-service       starts reporting service
    help                prints help
    print-help          prints help
    print-config        prints current configuration set by files & env variables
    print-version-info  prints version info

`

func printHelp() int {
	fmt.Printf(helpMessageTemplate, os.Args[0])
	return 0
}

func printConfig() int {
	configBytes, err := json.MarshalIndent(conf.Config, "", "    ")

	if err != nil {
		log.Error().Err(err)
		return 1
	}

	fmt.Println(string(configBytes))

	return 0
}

func main() {
	err := conf.LoadConfiguration(defaultConfigFilename)
	if err != nil {
		panic(err)
	}

	command := "start-service"

	if len(os.Args) >= 2 {
		command = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	os.Exit(handleCommand(command))
}

func handleCommand(command string) int {
	switch command {
	case "start-service":
		printVersionInfo()

		errCode := startService()
		if errCode != 0 {
			return errCode
		}
		return ExitStatusOK
	case "help", "print-help":
		return printHelp()
	case "print-config":
		return printConfig()
	case "print-version-info":
		printVersionInfo()
	default:
		fmt.Printf("\nCommand '%v' not found\n", command)
		return printHelp()
	}

	return ExitStatusOK
}

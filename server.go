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

// Entry point to the insights content service
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/conf"
	"github.com/RedHatInsights/insights-content-service/groups"
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
	groupsConfigPath := conf.GetGroupsConfiguration().ConfigPath
	groups, _ := groups.ParseGroupConfigFile(groupsConfigPath)

	for key, value := range groups {
		fmt.Println("Group id ", key)
		fmt.Println("Group content: ", value)
	}

	return ExitStatusOK
}

func printInfo(msg string, val string) {
	fmt.Printf("%s\t%s\n", msg, val)
}

func printVersionInfo() {
	printInfo("Version:", BuildVersion)
	printInfo("Build time:", BuildTime)
	printInfo("Branch:", BuildBranch)
	printInfo("Commit:", BuildCommit)
}

func initInfoLog(msg string) {
	log.Info().Str("type", "init").Msg(msg)
}

func logVersionInfo() {
	initInfoLog("Version: " + BuildVersion)
	initInfoLog("Build time: " + BuildTime)
	initInfoLog("Branch: " + BuildBranch)
	initInfoLog("Commit: " + BuildCommit)
}

const helpMessageTemplate = `
Service to provide content for OCP rules

Usage:

    %+v [command]

The commands are:

    <EMPTY>             starts content service
    start-service       starts content service
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
		logVersionInfo()

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

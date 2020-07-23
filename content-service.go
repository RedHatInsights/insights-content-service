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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RedHatInsights/insights-operator-utils/metrics"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/conf"
	"github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-content-service/groups"
	"github.com/RedHatInsights/insights-content-service/server"
)

const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota

	// ExitStatusServerError is returned in case of any REST API server-related error
	ExitStatusServerError

	// ExitStatusReadContentError is returned when the static content parsing fails
	ExitStatusReadContentError

	// ExitStatusOther represents other errors that might happen
	ExitStatusOther

	defaultConfigFilename = "config"
)

var (
	serverInstance *server.HTTPServer

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
	serverCfg := conf.GetServerConfiguration()
	groupsCfg := conf.GetGroupsConfiguration()
	groups, err := groups.ParseGroupConfigFile(groupsCfg.ConfigPath)

	if err != nil {
		log.Error().Err(err).Msg("Groups init error")
		return ExitStatusServerError
	}

	metricsCfg := conf.GetMetricsConfiguration()
	if metricsCfg.Namespace != "" {
		metrics.AddAPIMetricsWithNamespace(metricsCfg.Namespace)
	}

	ruleContentDirPath := conf.GetContentPathConfiguration()
	contentDir, err := content.ParseRuleContentDir(ruleContentDirPath)

	if osPathError, ok := err.(*os.PathError); ok {
		log.Error().Err(osPathError).Msg("No rules directory")
		return ExitStatusReadContentError
	}

	serverInstance = server.New(serverCfg, groups, contentDir)

	err = serverInstance.Start()
	if err != nil {
		log.Error().Err(err).Msg("HTTP(s) start error")
		return ExitStatusServerError
	}

	return ExitStatusOK
}

func printInfo(msg string, val string) {
	fmt.Printf("%s\t%s\n", msg, val)
}

func printVersionInfo() int {
	printInfo("Version:", BuildVersion)
	printInfo("Build time:", BuildTime)
	printInfo("Branch:", BuildBranch)
	printInfo("Commit:", BuildCommit)
	return ExitStatusOK
}

func printGroups() int {
	groupsConfig := conf.GetGroupsConfiguration()
	groups, err := groups.ParseGroupConfigFile(groupsConfig.ConfigPath)

	if err != nil {
		log.Error().Err(err).Msg("Groups init error")
		return ExitStatusServerError
	}

	fmt.Println(groups)
	return ExitStatusOK
}

func printRules() int {
	log.Info().Msg("Printing rules")
	contentPath := conf.GetContentPathConfiguration()
	contentDir, err := content.ParseRuleContentDir(contentPath)

	if err != nil {
		log.Error().Err(err).Msg("Error parsing the content")
		return ExitStatusReadContentError
	}

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)

	if err := encoder.Encode(contentDir); err == nil {
		fmt.Println(buffer)
		return ExitStatusOK
	}

	return ExitStatusOther

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
    print-groups        prints current groups configuration
    print-rules         prints current parsed rules
    print-version-info  prints version info

`

func printHelp() int {
	fmt.Printf(helpMessageTemplate, os.Args[0])
	return ExitStatusOK
}

func printConfig(config conf.ConfigStruct) int {
	configBytes, err := json.MarshalIndent(config, "", "    ")

	if err != nil {
		log.Error().Err(err)
		return ExitStatusOther
	}

	fmt.Println(string(configBytes))

	return ExitStatusOK
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
		if errCode != ExitStatusOK {
			return errCode
		}
		return ExitStatusOK
	case "help", "print-help":
		return printHelp()
	case "print-config":
		return printConfig(conf.Config)
	case "print-version-info":
		return printVersionInfo()
	case "print-groups":
		return printGroups()
	case "print-rules":
		return printRules()
	default:
		fmt.Printf("\nCommand '%v' not found\n", command)
		return printHelp()
	}
}

// Copyright 2020 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-content-service/groups"
)

// groupConfigMap is a shorthand for the map used to store the group configuration.
type groupConfigMap map[string]groups.Group

var (
	groupConfigPath = "./groups_config.yaml"
	contentDirPath  = "./content/"
)

func init() {
	flag.StringVar(&groupConfigPath, "config", groupConfigPath, "Path to the group configuration YAML file.")
	flag.StringVar(&contentDirPath, "content", contentDirPath, "Path to the content directory (the one containing the 'config.yaml' file).")
	flag.Parse()
}

func main() {
	initLogger()
	groupCfg := checkGroupConfig()
	checkRuleContent(groupCfg)
}

// initLogger initializes the zerolog library to pretty-print the log messages.
func initLogger() {
	err := logger.InitZerolog(
		logger.LoggingConfiguration{
			Debug:                      true,
			LogLevel:                   "debug",
			LoggingToCloudWatchEnabled: false,
		},
		logger.CloudWatchConfiguration{},
		logger.SentryLoggingConfiguration{},
		logger.KafkaZerologConfiguration{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initialize zerolog")
	}
}

// checkGroupConfig reads the group configuration file and performs defined checks on it.
// Then it returns the config to be used by the rule content checks.
func checkGroupConfig() groupConfigMap {
	groupCfg, err := groups.ParseGroupConfigFile(groupConfigPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse group config file")
	}

	// Unique group is just a check that makes sure no two groups have the same name property.
	uniqueGroups := map[string]string{}

	// For each group defined in the group configuration file.
	for groupKey, group := range groupCfg {
		if firstGroupKey, exists := uniqueGroups[group.Name]; exists {
			log.Warn().Msgf("multiple groups with the name '%s' (first with key '%s', but also with key '%s')", group.Name, firstGroupKey, groupKey)
		} else {
			uniqueGroups[group.Name] = groupKey
		}

		// Check for duplicate tag in a single group.
		// The same tag being used by multiple groups is allowed.
		uniqueTags := map[string]struct{}{}

		// For each tag assigned to the group.
		for _, tag := range group.Tags {
			if _, exists := uniqueTags[tag]; exists {
				log.Warn().Msgf("duplicate '%s' tag reference in group '%s'", tag, group.Name)
			} else {
				uniqueTags[tag] = struct{}{}
			}
		}
	}

	return groupCfg
}

// checkRuleContent checks if rule content files are not empty
// and if the tags assigned to all error codes really exist.
func checkRuleContent(groupCfg groupConfigMap) {
	ruleContentDir, _, err := content.ParseRuleContentDir(contentDirPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to rule content directory")
	}

	// For every rule with a content available.
	for ruleName, ruleContent := range ruleContentDir.Rules {
		checkRuleAttributeNotEmpty(ruleName, "name", ruleContent.Plugin.Name)
		checkRuleAttributeNotEmpty(ruleName, "node_id", ruleContent.Plugin.NodeID)
		checkRuleAttributeNotEmpty(ruleName, "product_code", ruleContent.Plugin.ProductCode)
		checkRuleAttributeNotEmpty(ruleName, "python_module", ruleContent.Plugin.PythonModule)

		checkRuleFileNotEmpty(ruleName, "summary.md", ruleContent.Summary)

		if len(ruleContent.ErrorKeys) == 0 {
			log.Warn().Msgf("rule '%s' contains no error code", ruleName)
		}
		// For every error code of the rule.
		for errCode, errContent := range ruleContent.ErrorKeys {
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "description", errContent.Metadata.Description)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "impact", errContent.Metadata.Impact.Name)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "publish_date", errContent.Metadata.PublishDate)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "status", errContent.Metadata.Status)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "likelihood", fmt.Sprint(errContent.Metadata.Likelihood))

			checkErrorCodeFileNotEmpty(ruleName, errCode, "generic.md", errContent.Generic)

			checkErrorCodeTags(groupCfg, ruleName, errCode, errContent)
		}
	}
}

// checkErrorCodeTags checks that the tags referenced by the error code are valid.
// At the end, all assigned tags (and the groups they belong to) are printed in the form of a map.
func checkErrorCodeTags(groupCfg groupConfigMap, ruleName, errCode string, errContent content.RuleErrorKeyContent) {
	errGroups := map[string][]string{}

	// For every tag of that error code.
	for _, errTag := range errContent.Metadata.Tags {
		// Check for duplicate tags in the error code's content.
		if _, exists := errGroups[errTag]; exists {
			log.Error().Msgf("duplicate tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
		}

		// List of groups to which the tag belongs.
		tagGroups := []string{}

		// Find a group with the tag.
		for _, group := range groupCfg {
			for _, tag := range group.Tags {
				if tag == errTag {
					tagGroups = append(tagGroups, group.Name)
					break
				}
			}
		}

		// Check if at least one group with the tag was found.
		if len(tagGroups) > 0 {
			errGroups[errTag] = tagGroups
		} else {
			log.Error().Msgf("unknown tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
		}
	}

	log.Info().Msgf("%s|%s: %v", ruleName, errCode, errGroups)
}

// Base rule content checks.

func checkRuleFileNotEmpty(ruleName, fileName, value string) {
	checkStringNotEmpty(
		fmt.Sprintf("content file '%s' of rule '%s'", fileName, ruleName),
		value,
	)
}

func checkRuleAttributeNotEmpty(ruleName, attribName, value string) {
	checkStringNotEmpty(
		fmt.Sprintf("attribute '%s' of rule '%s'", attribName, ruleName),
		value,
	)
}

// Error code content checks.

func checkErrorCodeFileNotEmpty(ruleName, errorCode, fileName string, value string) {
	checkStringNotEmpty(
		fmt.Sprintf("content file '%s' of error code '%s|%s'", fileName, ruleName, errorCode),
		value,
	)
}

func checkErrorCodeAttributeNotEmpty(ruleName, errorCode, attribName, value string) {
	checkStringNotEmpty(
		fmt.Sprintf("attribute '%s' of error code '%s|%s'", attribName, ruleName, errorCode),
		value,
	)
}

// Generic check for any name:value string pair.
func checkStringNotEmpty(name, value string) {
	if strings.TrimSpace(value) == "" {
		log.Warn().Msgf("%s is empty", name)
	}
}

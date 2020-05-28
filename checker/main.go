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

	"github.com/RedHatInsights/insights-content-service/groups"
	"github.com/RedHatInsights/insights-results-aggregator/content"
	"github.com/RedHatInsights/insights-results-aggregator/logger"
	"github.com/rs/zerolog/log"
)

type groupConfigMap map[string]groups.Group

var (
	groupConfigPath string = "./groups_config.yaml"
	contentDirPath  string = "./content/"
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

func initLogger() {
	err := logger.InitZerolog(
		logger.LoggingConfiguration{
			Debug:                      true,
			LogLevel:                   "debug",
			LoggingToCloudWatchEnabled: false,
		},
		logger.CloudWatchConfiguration{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initialize zerolog")
	}
}

func checkGroupConfig() groupConfigMap {
	groupCfg, err := groups.ParseGroupConfigFile(groupConfigPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse group config file")
	}

	// Check if all tags on all groups are unique.
	// - If no groups contains the same tag multiple times.
	// - If no two groups share the same tag name.
	uniqueTags := map[string]string{}
	// Unique group is just a check that makes sure no two groups have the same name property.
	uniqueGroups := map[string]string{}

	for groupKey, group := range groupCfg {
		if firstGroupKey, exists := uniqueGroups[group.Name]; exists {
			log.Warn().Msgf("multiple groups with the name '%s' (first with key '%s', but also with key '%s')", group.Name, firstGroupKey, groupKey)
		} else {
			uniqueGroups[group.Name] = groupKey
		}

		for _, tag := range group.Tags {
			if firstGroupName, exists := uniqueTags[tag]; exists {
				log.Warn().Msgf("tag '%s' is defined multiple times (first time in group '%s', but also in group '%s')", tag, firstGroupName, group.Name)
			} else {
				uniqueTags[tag] = group.Name
			}
		}
	}

	return groupCfg
}

func checkRuleContent(groupCfg groupConfigMap) {
	ruleContentDir, err := content.ParseRuleContentDir(contentDirPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse group config file")
	}

	// For every rule.
	for ruleName, ruleContent := range ruleContentDir.Rules {
		checkRuleAttributeNotEmpty(ruleName, "name", ruleContent.Plugin.Name)
		checkRuleAttributeNotEmpty(ruleName, "node_id", ruleContent.Plugin.NodeID)
		checkRuleAttributeNotEmpty(ruleName, "product_code", ruleContent.Plugin.ProductCode)
		checkRuleAttributeNotEmpty(ruleName, "python_module", ruleContent.Plugin.PythonModule)

		checkRuleFileNotEmpty(ruleName, "more_info.md", ruleContent.MoreInfo)
		checkRuleFileNotEmpty(ruleName, "reason.md", ruleContent.Reason)
		checkRuleFileNotEmpty(ruleName, "resolution.md", ruleContent.Resolution)
		checkRuleFileNotEmpty(ruleName, "summary.md", ruleContent.Summary)

		if len(ruleContent.ErrorKeys) == 0 {
			log.Warn().Msgf("rule '%s' contains no error code", ruleName)
		}

		// For every error code of that rule.
		for errCode, errContent := range ruleContent.ErrorKeys {
			checkErrorCodeFileNotEmpty(ruleName, errCode, "generic.md", errContent.Generic)

			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "condition", errContent.Metadata.Condition)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "description", errContent.Metadata.Description)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "impact", errContent.Metadata.Impact)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "publish_date", errContent.Metadata.PublishDate)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "status", errContent.Metadata.Status)
			checkErrorCodeAttributeNotEmpty(ruleName, errCode, "likelihood", fmt.Sprint(errContent.Metadata.Likelihood))

			checkErrorCodeTags(groupCfg, ruleName, errCode, errContent)
		}
	}
}

func checkErrorCodeTags(groupCfg groupConfigMap, ruleName string, errCode string, errContent content.RuleErrorKeyContent) {
	errGroups := map[string]string{}

	// For every tag of that error code.
	for _, errTag := range errContent.Metadata.Tags {
		// Check for duplicate tags in the error code's content.
		if _, exists := errGroups[errTag]; exists {
			log.Error().Msgf("duplicate tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
		}

		// Find a group with the tag.
		for _, group := range groupCfg {
			for _, tag := range group.Tags {
				if tag == errTag {
					errGroups[errTag] = group.Name
					break
				}
			}
		}

		// Check if at least one group with the tag was found.
		if _, exists := errGroups[errTag]; !exists {
			log.Error().Msgf("invalid tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
		}
	}

	log.Info().Msgf("%s|%s: %v", ruleName, errCode, errGroups)
}

// Base rule content checks.

func checkRuleFileNotEmpty(ruleName, fileName string, value []byte) {
	checkStringNotEmpty(
		fmt.Sprintf("content file '%s' of rule '%s'", fileName, ruleName),
		string(value),
	)
}

func checkRuleAttributeNotEmpty(ruleName, attribName, value string) {
	checkStringNotEmpty(
		fmt.Sprintf("attribute '%s' of rule '%s'", attribName, ruleName),
		value,
	)
}

// Error code content checks.

func checkErrorCodeFileNotEmpty(ruleName, errorCode, fileName string, value []byte) {
	checkStringNotEmpty(
		fmt.Sprintf("content file '%s' of error code '%s|%s'", fileName, ruleName, errorCode),
		string(value),
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

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

// Package content contains logic for parsing rule content.
package content

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/RedHatInsights/insights-operator-utils/types"
	"github.com/go-yaml/yaml"
	"github.com/rs/zerolog/log"
)

// Logging messages
const (
	separator          = "------------------------------------------------------------"
	directoryAttribute = "directory"
)

type (
	// RuleContent wraps all the content available for a rule into a single structure.
	RuleContent = types.RuleContent
	// RulePluginInfo is a Go representation of the `plugin.yaml`
	// file inside of the rule content directory.
	RulePluginInfo = types.RulePluginInfo
	// RuleErrorKeyContent wraps content of a single error key.
	RuleErrorKeyContent = types.RuleErrorKeyContent
	// ErrorKeyMetadata is a Go representation of the `metadata.yaml`
	// file inside of an error key content directory.
	ErrorKeyMetadata = types.ErrorKeyMetadata
	// RuleContentDirectory contains content for all available rules in a directory.
	RuleContentDirectory = types.RuleContentDirectory
	// GlobalRuleConfig represents the file that contains
	// metadata globally applicable to any/all rule content.
	GlobalRuleConfig = types.GlobalRuleConfig
)

// readFilesIntoByteArrayPointers reads the contents of the specified files
// in the base directory and saves them via the specified byte slice pointers.
func readFilesIntoFileContent(baseDir string, filelist []string) (map[string][]byte, error) {
	var filesContent = map[string][]byte{}
	for _, name := range filelist {
		log.Info().Msgf("Parsing %s/%s", baseDir, name)
		var err error
		rawBytes, err := ioutil.ReadFile(filepath.Clean(path.Join(baseDir, name)))
		if err != nil {
			filesContent[name] = nil
			log.Error().Err(err)
		} else {
			filesContent[name] = rawBytes
		}
	}

	return filesContent, nil
}

// createErrorContents takes a mapping of files into contents and perform
// some checks about it
func createErrorContents(contentRead map[string][]byte) (*RuleErrorKeyContent, error) {
	errorContent := RuleErrorKeyContent{}

	if contentRead["generic.md"] == nil {
		return nil, &MissingMandatoryFile{FileName: "generic.md"}
	}

	errorContent.Generic = string(contentRead["generic.md"])

	if contentRead["reason.md"] == nil {
		errorContent.Reason = ""
		errorContent.HasReason = false
	} else {
		errorContent.Reason = string(contentRead["reason.md"])
		errorContent.HasReason = true
	}

	if contentRead["metadata.yaml"] == nil {
		return nil, &MissingMandatoryFile{FileName: "metadata.yaml"}
	}

	if err := yaml.Unmarshal(contentRead["metadata.yaml"], &errorContent.Metadata); err != nil {
		return nil, err
	}

	return &errorContent, nil
}

// parseErrorContents reads the contents of the specified directory
// and parses all subdirectories as error key contents.
// This implicitly checks that the directory exists,
// so it is not necessary to ever check that elsewhere.
func parseErrorContents(ruleDirPath string) (map[string]RuleErrorKeyContent, error) {
	entries, err := ioutil.ReadDir(ruleDirPath)
	if err != nil {
		return nil, err
	}

	errorContents := map[string]RuleErrorKeyContent{}

	for _, e := range entries {
		if e.IsDir() {
			name := e.Name()
			contentFiles := []string{
				"generic.md",
				"metadata.yaml",
			}

			readContents, err := readFilesIntoFileContent(path.Join(ruleDirPath, name), contentFiles)
			if err != nil {
				return errorContents, err
			}

			errContents, err := createErrorContents(readContents)
			if err != nil {
				return errorContents, err
			}
			errorContents[name] = *errContents
		}
	}

	return errorContents, nil
}

// createRuleContent
func createRuleContent(contentRead map[string][]byte, errorKeys map[string]RuleErrorKeyContent) (*RuleContent, error) {
	ruleContent := RuleContent{ErrorKeys: errorKeys}

	if contentRead["plugin.yaml"] == nil {
		return nil, &MissingMandatoryFile{FileName: "plugin.yaml"}
	}

	if err := yaml.Unmarshal(contentRead["plugin.yaml"], &ruleContent.Plugin); err != nil {
		return nil, err
	}

	// The file "summary.md" is used inconsistently by applications. Since
	// the more accurate description or generic.md fields can be used
	// instead, summary.md becomes redundant. For consistency reason it's
	// still loaded, but it's ok if it's missing completely.
	//
	// See https://issues.redhat.com/browse/CCXDEV-5052 for context.
	if contentRead["summary.md"] == nil {
		log.Info().Msg("File summary.md is missing, using empty string instead")
		ruleContent.Summary = ""
	} else {
		ruleContent.Summary = string(contentRead["summary.md"])
	}

	if contentRead["reason.md"] == nil {
		// check error keys for a reason
		ruleContent.Reason = ""
		ruleContent.HasReason = false
		log.Warn().Msgf("reason for rule [%s] is empty", ruleContent.Plugin.PythonModule)
	} else {
		ruleContent.Reason = string(contentRead["reason.md"])
		ruleContent.HasReason = true
	}

	if contentRead["resolution.md"] == nil {
		ruleContent.Resolution = ""
		log.Warn().Msgf("resolution for rule [%s] is empty", ruleContent.Plugin.PythonModule)
	} else {
		ruleContent.Resolution = string(contentRead["resolution.md"])
	}

	if contentRead["more_info.md"] == nil {
		ruleContent.MoreInfo = ""
		log.Warn().Msgf("more_info for rule [%s] is empty", ruleContent.Plugin.PythonModule)
	} else {
		ruleContent.MoreInfo = string(contentRead["more_info.md"])
	}

	return &ruleContent, nil
}

// parseRuleContent attempts to parse all available rule content from the specified directory.
func parseRuleContent(ruleDirPath string) (RuleContent, error) {
	errorContents, err := parseErrorContents(ruleDirPath)

	if err != nil {
		return RuleContent{}, err
	}

	contentFiles := []string{
		"summary.md",
		"reason.md",
		"resolution.md",
		"more_info.md",
		"plugin.yaml",
	}

	readContent, err := readFilesIntoFileContent(ruleDirPath, contentFiles)
	if err != nil {
		return RuleContent{}, err
	}

	ruleContent, err := createRuleContent(readContent, errorContents)

	if err != nil {
		return RuleContent{}, err
	}
	return *ruleContent, err
}

// parseGlobalContentConfig reads the configuration file used to store
// metadata used by all rule content, such as impact dictionary.
func parseGlobalContentConfig(configPath string) (GlobalRuleConfig, error) {
	configBytes, err := ioutil.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return GlobalRuleConfig{}, err
	}

	conf := GlobalRuleConfig{}
	err = yaml.Unmarshal(configBytes, &conf)
	return conf, err
}

// parseRulesInDir function finds all rules and their content in the specified
// directory and stores the content in the provided map.
// This function also aggregates list of rules with improper content.
func parseRulesInDir(dirPath string, contentMap *map[string]RuleContent, invalidRules *[]string) error {
	// read the whole content of specified directory
	entries, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			name := e.Name()
			subdirPath := path.Join(dirPath, name)

			// Check if this directory directly contains a rule content.
			// This check is done for the subdirectories instead of the top directory
			// upon which this function is called because the very top level directory
			// should never directly contain any rule content and because the name
			// of the directory is much easier to access here without an extra call.
			if pluginYaml, err := os.Stat(path.Join(subdirPath, "plugin.yaml")); err == nil && os.FileMode.IsRegular(pluginYaml.Mode()) {
				log.Info().Str(directoryAttribute, subdirPath).Msg("plugin.yaml found")

				// let's accumulate error report with context (in which subdir it occured)
				ruleContent, err := parseRuleContent(subdirPath)
				if err != nil {
					log.Error().Err(err).Msgf("Error trying to parse rule in dir %v", subdirPath)
					message := fmt.Sprintf("Directory: %s, Error: %v", subdirPath, err)
					*invalidRules = append(*invalidRules, message)
					continue
				}

				// TODO: Add name uniqueness check.
				(*contentMap)[name] = ruleContent
			} else {
				// Otherwise, descend into the sub-directory and see if there is any rule content.
				log.Info().Str(directoryAttribute, subdirPath).Msg("descending into sub-directory")
				if err := parseRulesInDir(subdirPath, contentMap, invalidRules); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func printInvalidRules(invalidRules []string) {
	log.Info().Msg(separator)
	log.Error().Msg("List of invalid rules")
	for i, rule := range invalidRules {
		log.Error().Int("#", i+1).Str("Error", rule).Msg("Invalid rule")
	}
}

// ParseRuleContentDir finds all rule content in a directory and parses it.
func ParseRuleContentDir(contentDirPath string) (RuleContentDirectory, error) {
	globalConfig, err := parseGlobalContentConfig(path.Join(contentDirPath, "config.yaml"))
	if err != nil {
		return RuleContentDirectory{}, err
	}

	contentDir := RuleContentDirectory{
		Config: globalConfig,
		Rules:  map[string]RuleContent{},
	}

	// parse external and internal rules separately, because there are currently more categories
	// of rules, but they just don't have content yet, so in case the content for them appears.
	// If we want to parse all of them, the full contentDirPath can be passed to parseRulesInDir without problems
	externalContentDir := path.Join(contentDirPath, "external")

	// map used to store invalid rules
	invalidRules := make([]string, 0)

	err = parseRulesInDir(externalContentDir, &contentDir.Rules, &invalidRules)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse content of external rules")
		return contentDir, err
	}
	log.Info().
		Int("invalid external rules", len(invalidRules)).
		Msg("Parsing external rules: done")

	if len(invalidRules) > 0 {
		printInvalidRules(invalidRules)
	}

	internalContentDir := path.Join(contentDirPath, "internal")

	invalidRules = make([]string, 0)

	err = parseRulesInDir(internalContentDir, &contentDir.Rules, &invalidRules)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse content of internal rules")
		return contentDir, err
	}
	log.Info().
		Int("invalid internal rules", len(invalidRules)).
		Msg("Parsing internal rules: done")

	if len(invalidRules) > 0 {
		printInvalidRules(invalidRules)
	}

	return contentDir, err
}

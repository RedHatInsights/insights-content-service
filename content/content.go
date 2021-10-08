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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	internal_types "github.com/RedHatInsights/insights-content-service/types"
	"github.com/RedHatInsights/insights-operator-utils/collections"
	"github.com/RedHatInsights/insights-operator-utils/types"
	"github.com/go-yaml/yaml"
	"github.com/rs/zerolog/log"
)

// Logging messages
const (
	separator          = "------------------------------------------------------------"
	directoryAttribute = "directory"
	// PluginYAML represents the filename of rule's plugin specification
	PluginYAML = "plugin.yaml"
	// MetadataYAML represents the filename of rule error key's metadata
	MetadataYAML = "metadata.yaml"
	// GenericMarkdown contains a generic message that should briefly describe the recommendation
	GenericMarkdown = "generic.md"
	// SummaryMarkdown contains more descriptive information about the recommendation
	SummaryMarkdown = "summary.md"
	// ReasonMarkdown contains the reason why this recommendation was triggered
	ReasonMarkdown = "reason.md"
	// ResolutionMarkdown contains resolution steps to the given recommendation/issue
	ResolutionMarkdown = "resolution.md"
	// MoreInfoMarkdown contains additional information that further describe the recommendation
	MoreInfoMarkdown = "more_info.md"

	// InternalRulesGroup is a name for a group with all internal rules
	InternalRulesGroup = "internal"

	// ExternalRulesGroup is a name for a group with all external rules
	ExternalRulesGroup = "external"
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

var (
	// MandatoryRuleWideContentFiles are mandatory files that MUST be on rule plugin or error key level
	MandatoryRuleWideContentFiles = []string{GenericMarkdown, ReasonMarkdown}

	// SharedContentFiles MAY be on either the plugin level or the error key level
	SharedContentFiles = []string{GenericMarkdown, ReasonMarkdown, SummaryMarkdown, ResolutionMarkdown, MoreInfoMarkdown}

	// RulePluginMandatoryContentFiles are mandatory on the plugin level
	RulePluginMandatoryContentFiles = []string{PluginYAML}

	// RulePluginContentFiles are all files to look for on rule plugin level
	RulePluginContentFiles = append(SharedContentFiles, RulePluginMandatoryContentFiles...)

	// ErrorKeyMandatoryContentFiles are mandatory on the error key level
	ErrorKeyMandatoryContentFiles = []string{MetadataYAML}

	// ErrorKeyContentFiles are all files to look for on error key level
	ErrorKeyContentFiles = append(SharedContentFiles, ErrorKeyMandatoryContentFiles...)

	// GlobalConfig represents configrations globally applicable to any/all rule
	GlobalConfig GlobalRuleConfig
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

// checkErrorKeysForMandatoryContent iterates over filenames defined in the mandatory files array; ensures all error keys have the attribute set
func checkErrorKeysForMandatoryContent(errorKeys map[string]RuleErrorKeyContent) (valid bool) {
	valid = true

	for _, mandatoryFile := range MandatoryRuleWideContentFiles {
		for errorKeyName, errorKey := range errorKeys {
			// all error keys must have these attributes
			switch mandatoryFile {
			case GenericMarkdown:
				if errorKey.Generic == "" {
					log.Error().Msgf("Error key `%v` is missing mandatory file %v.", errorKeyName, GenericMarkdown)
					valid = false
				}
			case ReasonMarkdown:
				if errorKey.Reason == "" {
					log.Error().Msgf("Error key `%v` is missing mandatory file %v.", errorKeyName, ReasonMarkdown)
					valid = false
				}
			default:
				log.Error().Msgf("Behaviour for mandatory file `%v` is not defined.", mandatoryFile)
				valid = false
			}
		}
	}

	return
}

func copyContentToEmptyErrorKeys(
	filename string,
	ruleContent RuleContent,
	errorKeys map[string]RuleErrorKeyContent,
) {
	for i, errorKey := range errorKeys {
		ek := errorKey

		switch filename {
		case GenericMarkdown:
			if errorKey.Generic == "" {
				ek.Generic = ruleContent.Generic
			}
		case ReasonMarkdown:
			if errorKey.Reason == "" {
				ek.Reason = ruleContent.Reason
			}
		case SummaryMarkdown:
			if errorKey.Summary == "" {
				ek.Summary = ruleContent.Summary
			}
		case ResolutionMarkdown:
			if errorKey.Resolution == "" {
				ek.Resolution = ruleContent.Resolution
			}
		case MoreInfoMarkdown:
			if errorKey.MoreInfo == "" {
				ek.MoreInfo = ruleContent.MoreInfo
			}
		default:
			log.Error().Msgf("Behaviour for copying contents of file `%v` to error keys is not defined.", filename)
		}

		errorKeys[i] = ek
	}

}

// createErrorContents takes a mapping of files into contents and perform
// some checks about it
func createErrorContents(contentRead map[string][]byte) (*RuleErrorKeyContent, error) {
	errorContent := RuleErrorKeyContent{}
	errorContentMetadata := internal_types.ReceivedErrorKeyMetadata{}

	for _, filename := range ErrorKeyContentFiles {
		if contentRead[filename] == nil {
			if mandatory := collections.StringInSlice(filename, ErrorKeyMandatoryContentFiles); mandatory {
				return nil, &MissingMandatoryFile{FileName: filename}
			}

			log.Info().Msgf("File %v is missing on error key level, using empty string instead", filename)
		}

		if filename == MetadataYAML {
			if err := yaml.Unmarshal(contentRead[MetadataYAML], &errorContentMetadata); err != nil {
				return nil, err
			}

			errorContent.Metadata = errorContentMetadata.ToErrorKeyMetadata(GlobalConfig.Impact)

			continue
		}

		val := string(contentRead[filename])

		switch filename {
		case GenericMarkdown:
			errorContent.Generic = val
		case ReasonMarkdown:
			if val == "" {
				errorContent.HasReason = false
			}
			errorContent.HasReason = true
			errorContent.Reason = val
		case SummaryMarkdown:
			errorContent.Summary = val
		case ResolutionMarkdown:
			errorContent.Resolution = val
		case MoreInfoMarkdown:
			errorContent.MoreInfo = val
		default:
			log.Error().Msgf("Behaviour for handling of error key file `%v` is not defined.", filename)
		}
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

			readContents, err := readFilesIntoFileContent(path.Join(ruleDirPath, name), ErrorKeyContentFiles)
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

func createRuleContent(contentRead map[string][]byte, errorKeys map[string]RuleErrorKeyContent) (*RuleContent, error) {
	ruleContent := RuleContent{}

	for _, filename := range RulePluginContentFiles {
		if contentRead[filename] == nil {
			if mandatory := collections.StringInSlice(filename, RulePluginMandatoryContentFiles); mandatory {
				return nil, &MissingMandatoryFile{FileName: filename}
			}

			log.Info().Msgf("File %v is missing on plugin level, using empty string instead", filename)
		}

		if filename == PluginYAML {
			if err := yaml.Unmarshal(contentRead[PluginYAML], &ruleContent.Plugin); err != nil {
				return nil, err
			}
			continue
		}

		val := string(contentRead[filename])

		switch filename {
		case GenericMarkdown:
			ruleContent.Generic = val
		case ReasonMarkdown:
			if val == "" {
				ruleContent.HasReason = false
			}
			ruleContent.HasReason = true
			ruleContent.Reason = val
		case SummaryMarkdown:
			ruleContent.Summary = val
		case ResolutionMarkdown:
			ruleContent.Resolution = val
		case MoreInfoMarkdown:
			ruleContent.MoreInfo = val
		default:
			log.Error().Msgf("Behaviour for handling of plugin file `%v` is not defined.", filename)
		}

		copyContentToEmptyErrorKeys(filename, ruleContent, errorKeys)
	}

	ruleContent.ErrorKeys = errorKeys

	valid := checkErrorKeysForMandatoryContent(ruleContent.ErrorKeys)
	if !valid {
		return nil, errors.New("some of the error keys are missing mandatory attributes")
	}

	return &ruleContent, nil
}

// parseRuleContent attempts to parse all available rule content from the specified directory.
func parseRuleContent(ruleDirPath string) (RuleContent, error) {
	errorContents, err := parseErrorContents(ruleDirPath)

	if err != nil {
		return RuleContent{}, err
	}

	readContent, err := readFilesIntoFileContent(ruleDirPath, RulePluginContentFiles)
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
	if err != nil {
		log.Error().Err(err).Msgf("Can't apply global rule configurations")
	} else {
		GlobalConfig = conf
	}

	return conf, err
}

// updateRuleContentStatus function updates a map containing results of parsing
// all rules, external and internal ones
func updateRuleContentStatus(ruleContentStatusMap map[string]types.RuleContentStatus,
	ruleType types.RuleType, name string, loaded bool, err error) {
	// fill-in value to be used in Error attribute
	var parsingError = types.RuleParsingError("")
	if err != nil {
		parsingError = types.RuleParsingError(err.Error())
	}

	// new entry to a map
	ruleContentStatus := types.RuleContentStatus{
		RuleType: ruleType,
		Loaded:   loaded,
		Error:    parsingError,
	}

	// check for a name collision
	_, found := ruleContentStatusMap[name]
	if found {
		log.Error().Str("rule name", name).Msg("Duplicate rule name found")
	}

	// update map
	ruleContentStatusMap[name] = ruleContentStatus
}

// parseRulesInDir function finds all rules and their content in the specified
// directory and stores the content in the provided map.
// This function also aggregates list of rules with improper content.
func parseRulesInDir(dirPath string, ruleType types.RuleType,
	contentMap *map[string]RuleContent, invalidRules *[]string,
	ruleContentStatusMap map[string]types.RuleContentStatus) error {
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
			if pluginYaml, err := os.Stat(path.Join(subdirPath, PluginYAML)); err == nil && os.FileMode.IsRegular(pluginYaml.Mode()) {
				log.Info().Str(directoryAttribute, subdirPath).Msgf("%v found", PluginYAML)

				// let's accumulate error report with context (in which subdir it occurred)
				ruleContent, err := parseRuleContent(subdirPath)
				if err != nil {
					log.Error().Err(err).Msgf("Error trying to parse rule in dir %v", subdirPath)
					message := fmt.Sprintf("Directory: %s, Error: %v", subdirPath, err)
					*invalidRules = append(*invalidRules, message)

					updateRuleContentStatus(ruleContentStatusMap, ruleType, name, false, err)
					continue
				}

				// TODO: Add name uniqueness check.
				(*contentMap)[name] = ruleContent

				updateRuleContentStatus(ruleContentStatusMap, ruleType, name, true, nil)
			} else {
				// Otherwise, descend into the sub-directory and see if there is any rule content.
				log.Info().Str(directoryAttribute, subdirPath).Msg("descending into sub-directory")
				if err := parseRulesInDir(subdirPath, ruleType, contentMap, invalidRules, ruleContentStatusMap); err != nil {
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
func ParseRuleContentDir(contentDirPath string) (RuleContentDirectory, map[string]types.RuleContentStatus, error) {
	// we don't know in advance how many rules we have, so let's use nil slice there
	var ruleContentStatusMap map[string]types.RuleContentStatus = make(map[string]types.RuleContentStatus)

	globalConfig, err := parseGlobalContentConfig(path.Join(contentDirPath, "config.yaml"))
	if err != nil {
		return RuleContentDirectory{}, ruleContentStatusMap, err
	}

	contentDir := RuleContentDirectory{
		Config: globalConfig,
		Rules:  map[string]RuleContent{},
	}

	// parse external and internal rules separately, because there are currently more categories
	// of rules, but they just don't have content yet, so in case the content for them appears.
	// If we want to parse all of them, the full contentDirPath can be passed to parseRulesInDir without problems
	externalContentDir := path.Join(contentDirPath, ExternalRulesGroup)

	// map used to store invalid rules
	invalidRules := make([]string, 0)

	err = parseRulesInDir(externalContentDir, ExternalRulesGroup,
		&contentDir.Rules, &invalidRules, ruleContentStatusMap)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse content of external rules")
		return contentDir, ruleContentStatusMap, err
	}
	log.Info().
		Int("invalid external rules", len(invalidRules)).
		Msg("Parsing external rules: done")

	if len(invalidRules) > 0 {
		printInvalidRules(invalidRules)
	}

	internalContentDir := path.Join(contentDirPath, InternalRulesGroup)

	invalidRules = make([]string, 0)

	err = parseRulesInDir(internalContentDir, InternalRulesGroup,
		&contentDir.Rules, &invalidRules, ruleContentStatusMap)
	if err != nil {
		log.Error().Err(err).Msg("Cannot parse content of internal rules")
		return contentDir, ruleContentStatusMap, err
	}
	log.Info().
		Int("invalid internal rules", len(invalidRules)).
		Msg("Parsing internal rules: done")

	if len(invalidRules) > 0 {
		printInvalidRules(invalidRules)
	}

	return contentDir, ruleContentStatusMap, err
}

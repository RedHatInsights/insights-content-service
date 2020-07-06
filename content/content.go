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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/go-yaml/yaml"
	"github.com/rs/zerolog/log"
)

// GlobalRuleConfig represents the file that contains
// metadata globally applicable to any/all rule content.
type GlobalRuleConfig struct {
	Impact map[string]int `yaml:"impact" json:"impact"`
}

// ErrorKeyMetadata is a Go representation of the `metadata.yaml`
// file inside of an error key content directory.
type ErrorKeyMetadata struct {
	Condition   string   `yaml:"condition" json:"condition"`
	Description string   `yaml:"description" json:"description"`
	Impact      string   `yaml:"impact" json:"impact"`
	Likelihood  int      `yaml:"likelihood" json:"likelihood"`
	PublishDate string   `yaml:"publish_date" json:"publish_date"`
	Status      string   `yaml:"status" json:"status"`
	Tags        []string `yaml:"tags" json:"tags"`
}

// RuleErrorKeyContent wraps content of a single error key.
type RuleErrorKeyContent struct {
	Generic   string           `json:"generic"`
	Metadata  ErrorKeyMetadata `json:"metadata"`
	Reason    string           `json:"reason"`
	hasReason bool
}

// RulePluginInfo is a Go representation of the `plugin.yaml`
// file inside of the rule content directory.
type RulePluginInfo struct {
	Name         string `yaml:"name" json:"name"`
	NodeID       string `yaml:"node_id" json:"node_id"`
	ProductCode  string `yaml:"product_code" json:"product_code"`
	PythonModule string `yaml:"python_module" json:"python_module"`
}

// RuleContent wraps all the content available for a rule into a single structure.
type RuleContent struct {
	Summary    string                         `json:"summary"`
	Reason     string                         `json:"reason"`
	Resolution string                         `json:"resolution"`
	MoreInfo   string                         `json:"more_info"`
	Plugin     RulePluginInfo                 `json:"plugin"`
	ErrorKeys  map[string]RuleErrorKeyContent `json:"error_keys"`
	hasReason  bool
}

// RuleContentDirectory contains content for all available rules in a directory.
type RuleContentDirectory struct {
	Config GlobalRuleConfig
	Rules  map[string]RuleContent
}

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
		errorContent.hasReason = false
	} else {
		errorContent.Reason = string(contentRead["reason.md"])
		errorContent.hasReason = true
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
				"reason.md",
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

	if contentRead["summary.md"] == nil {
		return nil, &MissingMandatoryFile{FileName: "summary.md"}
	}

	ruleContent.Summary = string(contentRead["summary.md"])

	if contentRead["reason.md"] == nil {
		// check error keys for a reason
		ruleContent.Reason = ""
		ruleContent.hasReason = false
	} else {
		ruleContent.Reason = string(contentRead["reason.md"])
		ruleContent.hasReason = true
	}

	if contentRead["resolution.md"] == nil {
		return nil, &MissingMandatoryFile{FileName: "resolution.md"}
	}

	ruleContent.Resolution = string(contentRead["resolution.md"])

	if contentRead["more_info.md"] == nil {
		return nil, &MissingMandatoryFile{FileName: "more_info.md"}
	}

	ruleContent.MoreInfo = string(contentRead["more_info.md"])

	if contentRead["plugin.yaml"] == nil {
		return nil, &MissingMandatoryFile{FileName: "plugin.yaml"}
	}

	if err := yaml.Unmarshal(contentRead["plugin.yaml"], &ruleContent.Plugin); err != nil {
		return nil, err
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
	return *ruleContent, nil
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

// parseRulesInDir finds all rules and their content in the specified
// directory and stores the content in the provided map.
func parseRulesInDir(dirPath string, contentMap *map[string]RuleContent) error {
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
				ruleContent, err := parseRuleContent(subdirPath)
				if err != nil {
					return err
				}

				allRequiredFields := checkRequiredFields(ruleContent)

				if !allRequiredFields {
					// create an appropriate error and return
					log.Warn().Msgf("Some file in dir %s is missing", subdirPath)
					return &MissingMandatoryFile{FileName: "reason.md"}
				}

				// TODO: Add name uniqueness check.
				(*contentMap)[name] = ruleContent
			} else {
				// Otherwise, descend into the sub-directory and see if there is any rule content.
				if err := parseRulesInDir(subdirPath, contentMap); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// checkRequiredFields search if all the required fields in the RuleContent are ok
// at the moment only checks for Reason field
func checkRequiredFields(rule RuleContent) bool {
	if rule.hasReason {
		return true
	}

	for _, errorKeyContent := range rule.ErrorKeys {
		if !errorKeyContent.hasReason {
			return false
		}
	}

	return true
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

	externalContentDir := path.Join(contentDirPath, "external")
	err = parseRulesInDir(externalContentDir, &contentDir.Rules)

	return contentDir, err
}

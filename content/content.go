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
	Generic  string           `json:"generic"`
	Metadata ErrorKeyMetadata `json:"metadata"`
	Reason   string           `json:"reason"`
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
}

// RuleContentDirectory contains content for all available rules in a directory.
type RuleContentDirectory struct {
	Config GlobalRuleConfig
	Rules  map[string]RuleContent
}

// readFilesIntoByteArrayPointers reads the contents of the specified files
// in the base directory and saves them via the specified byte slice pointers.
func readFilesIntoByteArray(baseDir string, filelist []string) (map[string][]byte, error) {
	var filesContent = map[string][]byte{}
	for _, name := range filelist {
		var err error
		ptr, err := ioutil.ReadFile(filepath.Clean(path.Join(baseDir, name)))
		if err != nil {
			log.Error().Err(err)
		}
		filesContent[name] = ptr
	}
	return filesContent, nil
}

// readFilesIntoStringPointers reads the content of the specified files
// in the base directory and saves them via the specified string pointer
func readFilesIntoString(baseDir string, filelist []string) (map[string]string, error) {
	var filesContent = map[string]string{}
	for _, name := range filelist {
		var err error
		rawBytes, err := ioutil.ReadFile(filepath.Clean(path.Join(baseDir, name)))
		filesContent[name] = string(rawBytes)
		if err != nil {
			log.Error().Err(err)
		}
	}
	return filesContent, nil
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

			errContent := RuleErrorKeyContent{}
			contentFiles := []string{
				"generic.md",
				"reason.md",
			}
			yamlFiles := []string{"metadata.yaml"}

			readStrings, err := readFilesIntoString(path.Join(ruleDirPath, name), contentFiles)
			if err != nil {
				return errorContents, err
			}
			errContent.Generic = readStrings["generic.md"]
			errContent.Reason = readStrings["reason.md"]

			readBytes, err := readFilesIntoByteArray(path.Join(ruleDirPath, name), yamlFiles)
			if err != nil {
				return errorContents, err
			}

			metadataBytes := readBytes["metadata.yaml"]
			if err := yaml.Unmarshal(metadataBytes, &errContent.Metadata); err != nil {
				return errorContents, err
			}

			errorContents[name] = errContent
		}
	}

	return errorContents, nil
}

// parseRuleContent attempts to parse all available rule content from the specified directory.
func parseRuleContent(ruleDirPath string) (RuleContent, error) {
	errorContents, err := parseErrorContents(ruleDirPath)
	if err != nil {
		return RuleContent{}, err
	}

	ruleContent := RuleContent{ErrorKeys: errorContents}
	contentFiles := []string{
		"summary.md",
		"reason.md",
		"resolution.md",
		"more_info.md",
	}
	yamlFiles := []string{
		"plugin.yaml",
	}

	readStrings, err := readFilesIntoString(ruleDirPath, contentFiles)
	if err != nil {
		return RuleContent{}, err
	}

	ruleContent.Summary = readStrings["summary.md"]
	ruleContent.Reason = readStrings["reason.md"]
	ruleContent.Resolution = readStrings["resolution.md"]
	ruleContent.MoreInfo = readStrings["more_info.md"]

	readBytes, err := readFilesIntoByteArray(ruleDirPath, yamlFiles)
	if err != nil {
		return RuleContent{}, err
	}

	pluginBytes := readBytes["plugin.yaml"]
	if err := yaml.Unmarshal(pluginBytes, &ruleContent.Plugin); err != nil {
		return RuleContent{}, err
	}

	return ruleContent, nil
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
					return Error()
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
	if rule.Reason != "" {
		return true
	}

	errorKeyReasonFound := false

	for _, errorKeyContent := range rule.ErrorKeys {
		if errorKeyContent.Reason != "" {
			errorKeyReasonFound = true
		} else {
			if errorKeyReasonFound {
				return false
			}
		}
	}

	return false
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

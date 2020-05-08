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

package groups

import (
	"io/ioutil"
	"path/filepath"

	"github.com/go-yaml/yaml"
)

// Group represent the relative information about a group
type Group struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// ParseGroupConfigFile parses the groups configuration file and return the read groups
func ParseGroupConfigFile(groupConfigPath string) (map[string]Group, error) {
	configBytes, err := ioutil.ReadFile(filepath.Clean(groupConfigPath))
	if err != nil {
		return nil, err
	}

	var groups map[string]Group

	err = yaml.Unmarshal(configBytes, &groups)

	if err != nil {
		return nil, err
	}

	return groups, nil
}

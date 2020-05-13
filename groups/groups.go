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
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/go-yaml/yaml"
	"github.com/rs/zerolog/log"
)

// GroupsSetup
type GroupsSetup interface {
	Init() error
	GetGroups() (map[string]Group, error)
}

// Group represent the relative information about a group
type Group struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// GroupsSetupImpl is an implementation of GroupsSetup interface
type GroupsSetupImpl struct {
	conf      Configuration
	groupsMap map[string]Group
}

// New constructs new implementation of GroupsSetup interface
func New(config Configuration) *GroupsSetupImpl {
	return &GroupsSetupImpl{
		conf: config,
	}
}

// Init load the groups definition file into memory structures
func (groupSetup *GroupsSetupImpl) Init() error {
	var err error

	groupConfigPath := groupSetup.conf.ConfigPath
	groupSetup.groupsMap, err = parseGroupConfigFile(groupConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("Groups configuration file not valid")
		return err
	}

	return nil
}

// GetGroups return the map of configured groups and its attributes
func (groupSetup *GroupsSetupImpl) GetGroups() (map[string]Group, error) {
	if groupSetup.groupsMap == nil {
		err := errors.New("No groups configuration is loaded")
		log.Error().Err(err).Msg("No groups configuration is loaded")
		return nil, err
	}

	return groupSetup.groupsMap, nil
}

// parseGroupConfigFile parses the groups configuration file and return the read groups
func parseGroupConfigFile(groupConfigPath string) (map[string]Group, error) {
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

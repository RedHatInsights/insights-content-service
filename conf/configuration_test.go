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

package conf_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/RedHatInsights/insights-content-service/conf"
)

func init() {
	err := os.Chdir("..")
	if err != nil {
		panic(err)
	}
}

func mustLoadConfiguration(t *testing.T, path string) {
	err := conf.LoadConfiguration(path)
	if err != nil {
		t.Fatal(err)
	}
}

func mustSetEnv(t *testing.T, key, val string) {
	err := os.Setenv(key, val)
	if err != nil {
		t.Fatal(err)
	}
}

func loadProperConfigFile(t *testing.T) {
	os.Clearenv()
	mustLoadConfiguration(t, "tests/config")
}

func loadConfigFileWithWrongFiles(t *testing.T) {
	os.Clearenv()
	mustLoadConfiguration(t, "tests/config_wrong_files")
}

// TestLoadConfiguration loads a configuration file for testing
func TestLoadConfiguration(t *testing.T) {
	loadProperConfigFile(t)
}

// TestLoadBrokenConfiguration loads a configuration file for testing
func TestLoadBrokenConfiguraion(t *testing.T) {
	os.Clearenv()
	err := conf.LoadConfiguration("tests/config_improper_format")
	if err == nil {
		t.Fatal("Broken configuration file should be detected")
	}
}

// TestLoadGroupsConfiguration tests loading the groups configuration sub-tree
func TestLoadGroupsConfiguration(t *testing.T) {
	loadProperConfigFile(t)

	GroupsCfg := conf.GetGroupsConfiguration()

	assert.Equal(t, "groups_config.yaml", GroupsCfg.ConfigPath)
}

// TestLoadServerConfiguration tests loading the server configuration sub-tree
func TestLoadServerConfiguration(t *testing.T) {
	loadProperConfigFile(t)

	serverCfg := conf.GetServerConfiguration()

	assert.Equal(t, ":8080", serverCfg.Address)
	assert.Equal(t, "/api/v1/", serverCfg.APIPrefix)
}

// TestLoadContentPathConfiguration tests loading the content path configuration
func TestLoadContentPathConfiguration(t *testing.T) {
	loadProperConfigFile(t)

	contentPath := conf.GetContentPathConfiguration()

	assert.Equal(t, "/rules-content", contentPath)
}

// TestLoadConfigurationEnvVariable tests loading the config. file for testing from an environment variable
func TestLoadConfigurationEnvVariable(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_CONTENT_SERVICE_CONFIG_FILE", "tests/config")

	mustLoadConfiguration(t, "foobar")
}

// TestLoadConfigurationEnvVariableNegative tests loading the config. file for testing from an environment variable
func TestLoadConfigurationEnvVariableNegative(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_CONTENT_SERVICE_CONFIG_FILE", "does not exists")

	err := conf.LoadConfiguration("foobar")
	if err == nil {
		t.Fatal("Error should be reported for non existing file")
	}
}

// TestTryToLoadNonExistingConfig checks if non existing config file causes failure or not
func TestTryToLoadNonExistingConfig(t *testing.T) {
	os.Clearenv()
	err := conf.LoadConfiguration("foobar")
	if err != nil {
		t.Fatal(err)
	}
}

// TestCheckIfFileExists tests the functionality of function checkIfFileExists
func TestCheckIfFileExists(t *testing.T) {
	err := conf.CheckIfFileExists("")
	if err == nil {
		t.Fatal("File with empty name should not exists")
	}

	err = conf.CheckIfFileExists("config.toml")
	if err != nil {
		t.Fatal("File should exists:", err)
	}

	err = conf.CheckIfFileExists("\n")
	if err == nil {
		t.Fatal("File '' should not exist")
	}

	err = conf.CheckIfFileExists(".")
	if err == nil {
		t.Fatal("File '.' is a directory")
	}

	err = conf.CheckIfFileExists("..")
	if err == nil {
		t.Fatal("File '..' is a directory")
	}
}

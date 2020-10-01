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

// Package conf contains definition of data type named ConfigStruct that
// represents configuration of Content service. This package also contains
// function named LoadConfiguration that can be used to load configuration from
// provided configuration file and/or from environment variables. Additionally
// several specific functions named GetServerConfiguration, GetGroupsConfiguration,
// GetContentPathConfiguration, GetMetricsConfiguration, GetLoggingConfiguration and
// GetCloudWatchConfiguration are to be used to return specific
// configuration options.
//
// Generated documentation is available at:
// https://godoc.org/github.com/RedHatInsights/insights-content-service/conf
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-content-service/packages/conf/configuration.html
package conf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/RedHatInsights/insights-content-service/groups"
	"github.com/RedHatInsights/insights-content-service/server"
)

const (
	configFileEnvVariableName = "INSIGHTS_CONTENT_SERVICE_CONFIG_FILE"
	defaultContentPath        = "/rules-content"
)

// MetricsConf contains the metrics configuration
type MetricsConf struct {
	Namespace string `mapstructure:"namespace" toml:"namespace"`
}

// ConfigStruct is a structure holding the whole service configuration
type ConfigStruct struct {
	Server  server.Configuration `mapstructure:"server" toml:"server"`
	Groups  groups.Configuration `mapstructure:"groups" toml:"groups"`
	Content struct {
		ContentPath string `mapstructure:"path" toml:"path"`
	} `mapstructure:"content" toml:"content"`
	Metrics    MetricsConf                    `mapstructure:"metrics" toml:"metrics"`
	Logging    logger.LoggingConfiguration    `mapstructure:"logging" toml:"logging"`
	CloudWatch logger.CloudWatchConfiguration `mapstructure:"cloudwatch" toml:"cloudwatch"`
}

// Config has exactly the same structure as *.toml file
var Config ConfigStruct

// LoadConfiguration loads configuration from defaultConfigFile, file set in
// configFileEnvVariableName or from env
func LoadConfiguration(defaultConfigFile string) error {
	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// viper is not smart enough to understand the structure of
		// config by itself
		fakeTomlConfigWriter := new(bytes.Buffer)

		err := toml.NewEncoder(fakeTomlConfigWriter).Encode(Config)
		if err != nil {
			return err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env

	const envPrefix = "INSIGHTS_CONTENT_SERVICE_"

	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	return viper.Unmarshal(&Config)
}

// GetServerConfiguration returns server configuration
func GetServerConfiguration() server.Configuration {
	err := checkIfFileExists(Config.Server.APISpecFile)
	if err != nil {
		log.Fatal().Err(err).Msg("All customer facing APIs MUST serve the current OpenAPI specification")
	}

	return Config.Server
}

// GetGroupsConfiguration returns groups configuration
func GetGroupsConfiguration() groups.Configuration {
	err := checkIfFileExists(Config.Groups.ConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("The groups configuration file is not defined")
	}

	return Config.Groups
}

// GetContentPathConfiguration get the path to the content files from the
// configuration
func GetContentPathConfiguration() string {
	if len(Config.Content.ContentPath) == 0 {
		Config.Content.ContentPath = defaultContentPath
	}

	return Config.Content.ContentPath
}

// GetMetricsConfiguration get MetricsConf from the loaded configuration
func GetMetricsConfiguration() MetricsConf {
	return Config.Metrics
}

// GetLoggingConfiguration returns logging configuration
func GetLoggingConfiguration() logger.LoggingConfiguration {
	return Config.Logging
}

// GetCloudWatchConfiguration returns cloudwatch configuration
func GetCloudWatchConfiguration() logger.CloudWatchConfiguration {
	return Config.CloudWatch
}

// checkIfFileExists returns nil if path doesn't exist or isn't a file,
// otherwise it returns corresponding error
func checkIfFileExists(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("Empty path provided")
	}
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("The following file path does not exist. Path: '%v'", path)
	} else if err != nil {
		return err
	}

	if fileMode := fileInfo.Mode(); !fileMode.IsRegular() {
		return fmt.Errorf("The following file path is not a regular file. Path: '%v'", path)
	}

	return nil
}

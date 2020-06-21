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

package helpers

import (
	"testing"

	"github.com/RedHatInsights/insights-results-aggregator-utils/tests/helpers"

	"github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-content-service/groups"
	"github.com/RedHatInsights/insights-content-service/server"
)

// APIRequest is a request to api to use in AssertAPIRequest
type APIRequest = helpers.APIRequest

// APIResponse is an expected api response to use in AssertAPIRequest
type APIResponse = helpers.APIResponse

// DefaultServerConfig is a default config used by AssertAPIRequest
var DefaultServerConfig = server.Configuration{
	Address:     ":8080",
	APIPrefix:   "/api/test/",
	APISpecFile: "openapi.json",
	Debug:       true,
	UseHTTPS:    false,
}

// AssertAPIRequest creates new server
// (which you can keep nil so it will be created automatically)
// and provided serverConfig(you can leave it empty to use the default one)
// sends api request and checks api response (see docs for APIRequest and APIResponse)
func AssertAPIRequest(
	t testing.TB,
	serverConfig *server.Configuration,
	request *APIRequest,
	expectedResponse *APIResponse,
) {
	if serverConfig == nil {
		serverConfig = &DefaultServerConfig
	}

	// TODO: it should be configurable
	groupsData := make(map[string]groups.Group)
	groupsData["foo"] = groups.Group{
		Name:        "group name: foo",
		Description: "group description: foo",
		Tags:        []string{"tag1", "tag2"},
	}
	groupsData["bar"] = groups.Group{
		Name:        "group name: bar",
		Description: "group description: bar",
		Tags:        []string{"tag3", "tag4"},
	}
	contentDir := content.RuleContentDirectory{}
	testServer := server.New(*serverConfig, groupsData, contentDir)

	helpers.AssertAPIRequest(t, testServer, serverConfig.APIPrefix, request, expectedResponse)
}

// ExecuteRequest executes http request on a testServer
var ExecuteRequest = helpers.ExecuteRequest

// CheckResponseBodyJSON checks if body is the same json as in expected
// (ignores whitespaces, newlines, etc)
// also validates both expected and body to be a valid json
var CheckResponseBodyJSON = helpers.CheckResponseBodyJSON

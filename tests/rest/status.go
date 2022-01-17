/*
Copyright © 2021, 2022 Red Hat, Inc.

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

package tests

import (
	"encoding/json"
	"errors"

	types "github.com/RedHatInsights/insights-results-types"
	"github.com/verdverm/frisby"
)

// URL to endpoint being tested there
const statusURL = apiURL + "status"

// StatusResponse represents response containing map of rules
type StatusResponse struct {
	RuleContentStatusMap map[string]types.RuleContentStatus `json:"rules"`
	Status               string                             `json:"status"`
}

// checkStatusResponseContent check the actual content received from the server
func checkStatusResponseContent(payload []byte) error {
	response := StatusResponse{}

	// check if the 'status' response has proper format
	err := json.Unmarshal(payload, &response)
	if err != nil {
		// deserialization failed
		return err
	}

	if response.Status != "ok" {
		// unexpected status detected
		return errors.New("ok status expected")
	}
	for name, value := range response.RuleContentStatusMap {
		// rudimentary check for rule name
		if name == "" {
			return errors.New("wrong rule name")
		}
		// RuleType should be either "internal" or "external", nothing else
		if value.RuleType != "internal" && value.RuleType != "external" {
			return errors.New("wrong ruleType field")
		}
		if value.Loaded {
			// loaded rules should have 'error' field empty
			if value.Error != "" {
				return errors.New("error field is not empty for loaded rule")
			}
		} else {
			// not loaded rules should have 'error' with error message
			if value.Error == "" {
				return errors.New("error field is empty for not loaded rule")
			}
		}
	}
	// everything seems to be ok
	return nil
}

// checkStatusEndpoint checks whether 'status' endpoint is handled correctly
func checkStatusEndpoint() {
	f := frisby.Create("Check the 'status' endpoint").Get(groupsURL)
	f.Send()
	f.ExpectStatus(200)
	f.ExpectHeader(contentTypeHeader, "application/json; charset=utf-8")
	f.PrintReport()

	// try to read payload
	text, err := f.Resp.Content()
	if err != nil {
		f.AddError(err.Error())
		return
	}

	// payload seems to part of response - let's check its content
	err = checkStatusResponseContent(text)
	if err != nil {
		f.AddError(err.Error())
	}
}

// checkWrongMethodsForStatusEndpoint check whether other HTTP methods are
// rejected correctly for the REST API 'status' endpoint
func checkWrongMethodsForStatusEndpoint() {
	checkGetEndpointByOtherMethods(statusURL, false)
}

/*
Copyright © 2020, 2021, 2022 Red Hat, Inc.

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
	"unicode"
	"unicode/utf8"

	"github.com/verdverm/frisby"
)

const groupsURL = apiURL + "groups"

// Group represents part of response containing list of groups
type Group struct {
	Description string   `json:"description"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
}

// GroupsResponse represents response containing list of groups
type GroupsResponse struct {
	Groups []Group `json:"groups"`
	Status string  `json:"status"`
}

/*
checkGroupsResponseContent check if the response for 'groups' endpoint has the following format:
{
    "groups": [
        {
            "description": "High utilization, proposed tuned profiles, storage issues",
            "tags": [
                "performance"
            ],
            "title": "Performance"
        },
        {
            "description": "Operator degraded, missing functionality due to misconfiguration or resource constraints.",
            "tags": [
                "service_availability"
            ],
            "title": "Service Availability"
        },
        {
            "description": "Issues related to certificates, user management, security groups, specific port usage, storage permissions, usage of kubeadmin account, exposed keys etc.",
            "tags": [
                "security"
            ],
            "title": "Security"
        },
        {
            "description": "Load balancer issues, machine api and autoscaler issues, failover issues, nodes down, cluster api/cluster provider issues.",
            "tags": [
                "fault_tolerance"
            ],
            "title": "Fault Tolerance"
        }
    ],
    "status": "ok"
}
*/
func checkGroupsResponseContent(payload []byte) error {
	response := GroupsResponse{}

	// check if the 'groups' response has proper format
	err := json.Unmarshal(payload, &response)
	if err != nil {
		// deserialization failed
		return err
	}

	if response.Status != "ok" {
		// unexpected status detected
		return errors.New("ok status expected")
	}
	for _, group := range response.Groups {
		err := checkTextAttribute(group.Title, "title")
		if err != nil {
			return err
		}
		err = checkTextAttribute(group.Description, "description")
		if err != nil {
			return err
		}
	}
	// everything seems to be ok
	return nil
}

// checkTextAttribute is a rudimentary check for text attributes
func checkTextAttribute(text, what string) error {
	// check for empty string
	if text == "" {
		return errors.New("empty " + what + " detected in a group")
	}

	// check the capitalization
	firstRune, _ := utf8.DecodeRuneInString(text)
	if !unicode.IsUpper(firstRune) {
		return errors.New(what + " should start by uppercase letter")
	}

	// everything seems to be ok
	return nil
}

// checkGroupsEndpoint checks whether 'groups' endpoint is handled correctly
func checkGroupsEndpoint() {
	f := frisby.Create("Check the 'groups' endpoint").Get(groupsURL)
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
	err = checkGroupsResponseContent(text)
	if err != nil {
		f.AddError(err.Error())
	}
}

// checkWrongMethodsForGroupsEndpoint check whether other HTTP methods are rejected correctly for the REST API 'groups' point
func checkWrongMethodsForGroupsEndpoint() {
	checkGetEndpointByOtherMethods(groupsURL, false)
}

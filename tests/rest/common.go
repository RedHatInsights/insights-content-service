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
	"encoding/base64"
	"fmt"

	"github.com/verdverm/frisby"
)

// common constants used by REST API tests
const (
	apiURL            = "http://localhost:8080/api/v1/"
	contentTypeHeader = "Content-Type"

	authHeaderName = "x-rh-identity"

	// ContentTypeJSON represents MIME type for JSON format
	ContentTypeJSON = "application/json; charset=utf-8"

	// ContentTypeText represents MIME type for plain text format
	ContentTypeText = "text/plain; charset=utf-8"
)

// test names, messages and logs-related constants
const (
	httpPostMethod    = "POST"
	httpPutMethod     = "PUT"
	httpDeleteMethod  = "DELETE"
	httpHeadMethod    = "HEAD"
	httpPatchMethod   = "PATCH"
	httpOptionsMethod = "OPTIONS"

	// reused test names
	checkWrongEndpointTest = "Check the end point %s with wrong method: %s"
)

// StatusOnlyResponse represents response containing just a status
type StatusOnlyResponse struct {
	Status string `json:"status"`
}

// setAuthHeaderForOrganization set authorization header to request
func setAuthHeaderForOrganization(f *frisby.Frisby, orgID int) {
	plainHeader := fmt.Sprintf("{\"identity\": {\"internal\": {\"org_id\": \"%d\"}}}", orgID)
	encodedHeader := base64.StdEncoding.EncodeToString([]byte(plainHeader))
	f.SetHeader(authHeaderName, encodedHeader)
}

// setAuthHeader set authorization header to request for organization 1
func setAuthHeader(f *frisby.Frisby) {
	setAuthHeaderForOrganization(f, 1)
}

// sendAndExpectStatus sends the request to the server and checks whether expected HTTP code (status) is returned
func sendAndExpectStatus(f *frisby.Frisby, expectedStatus int) {
	f.Send()
	f.ExpectStatus(expectedStatus)
	f.PrintReport()
}

// checkGetEndpointByOtherMethods checks whether a 'GET' endpoint respond correctly if other HTTP methods are used
func checkGetEndpointByOtherMethods(endpoint string, includingOptions bool) {
	f := frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpPostMethod)).Post(endpoint)
	sendAndExpectStatus(f, 405)

	f = frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpPutMethod)).Put(endpoint)
	sendAndExpectStatus(f, 405)

	f = frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpDeleteMethod)).Delete(endpoint)
	sendAndExpectStatus(f, 405)

	f = frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpPatchMethod)).Patch(endpoint)
	sendAndExpectStatus(f, 405)

	f = frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpHeadMethod)).Head(endpoint)
	sendAndExpectStatus(f, 405)

	// some endpoints accepts OPTIONS method together with GET one, so this check is fully optional
	if includingOptions {
		f = frisby.Create(fmt.Sprintf(checkWrongEndpointTest, endpoint, httpOptionsMethod)).Options(endpoint)
		sendAndExpectStatus(f, 405)
	}
}

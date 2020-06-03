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

package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
)

const (
	// MainEndpoint defines suffix of the root endpoint
	MainEndpoint = ""

	// GroupsEndpoint defines suffix of the groups request endpoint
	GroupsEndpoint = "groups"
)

func (server *HTTPServer) addEndpointsToRouter(router *mux.Router) {
	apiPrefix := server.Config.APIPrefix
	openAPIURL := apiPrefix + filepath.Base(server.Config.APISpecFile)

	// common REST API endpoints
	router.HandleFunc(apiPrefix+MainEndpoint, server.mainEndpoint).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+GroupsEndpoint, server.listOfGroups).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+"test", server.getStaticContent).Methods(http.MethodGet, http.MethodOptions)

	// OpenAPI specs
	router.HandleFunc(openAPIURL, server.serveAPISpecFile).Methods(http.MethodGet)
}

// MakeURLToEndpoint creates URL to endpoint, use constants from file endpoints.go
func MakeURLToEndpoint(apiPrefix, endpoint string, args ...interface{}) string {
	re := regexp.MustCompile(`\{[a-zA-Z_0-9]+\}`)
	endpoint = re.ReplaceAllString(endpoint, "%v")
	return apiPrefix + fmt.Sprintf(endpoint, args...)
}

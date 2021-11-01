/*
Copyright © 2020, 2021 Red Hat, Inc.

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
	"net/http"
	"path/filepath"

	httputils "github.com/RedHatInsights/insights-operator-utils/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// MainEndpoint defines suffix of the root endpoint
	MainEndpoint = ""
	// GroupsEndpoint defines suffix of the groups request endpoint
	GroupsEndpoint = "groups"
	// AllContentEndpoint defines suffix for all the content
	AllContentEndpoint = "content"
	// MetricsEndpoint returns prometheus metrics
	MetricsEndpoint = "metrics"
	// StatusEndpoint returns status of all rules that have been read and
	// parsed
	StatusEndpoint = "status"
	// InfoEndpoint returns basic information about content service
	// version, utils repository version, commit hash etc.
	InfoEndpoint = "info"
)

func (server *HTTPServer) addEndpointsToRouter(router *mux.Router) {
	apiPrefix := server.Config.APIPrefix
	openAPIURL := apiPrefix + filepath.Base(server.Config.APISpecFile)

	// common REST API endpoints
	router.HandleFunc(apiPrefix+MainEndpoint, server.mainEndpoint).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+GroupsEndpoint, server.listOfGroups).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+AllContentEndpoint, server.getStaticContent).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+StatusEndpoint, server.ruleContentStates).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+InfoEndpoint, server.infoMap).Methods(http.MethodGet, http.MethodOptions)

	// Prometheus metrics
	router.Handle(apiPrefix+MetricsEndpoint, promhttp.Handler()).Methods(http.MethodGet)

	// OpenAPI specs
	router.HandleFunc(
		openAPIURL,
		httputils.CreateOpenAPIHandler(server.Config.APISpecFile, server.Config.Debug, true),
	).Methods(http.MethodGet)
}

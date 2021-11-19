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

// Package server contains implementation of REST API server (HTTPServer) for the
// Insights content service. In current version, the following
// REST API endpoints are available:
package server

import (
	"context"
	"net/http"

	httputils "github.com/RedHatInsights/insights-operator-utils/http"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-content-service/groups"
)

// HTTPServer in an implementation of Server interface
type HTTPServer struct {
	Config     Configuration
	InfoParams map[string]string
	Groups     map[string]groups.Group
	Content    content.RuleContentDirectory
	Serv       *http.Server

	encodedContent       []byte
	groupsList           []groups.Group
	ruleContentStatusMap map[string]types.RuleContentStatus
}

// New constructs new implementation of Server interface
func New(config Configuration, groups map[string]groups.Group,
	contentDir content.RuleContentDirectory,
	ruleContentStatusMap map[string]types.RuleContentStatus) *HTTPServer {
	return &HTTPServer{
		Config:               config,
		Groups:               groups,
		Content:              contentDir,
		ruleContentStatusMap: ruleContentStatusMap,
		InfoParams:           make(map[string]string),
	}
}

// Start starts server
func (server *HTTPServer) Start() error {
	address := server.Config.Address
	log.Info().Msgf("Starting HTTP server at '%s'", address)
	router := server.Initialize()
	server.Serv = &http.Server{Addr: address, Handler: router}
	var err error

	err = server.Serv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Unable to start HTTP/S server")
		return err
	}

	return nil
}

// Stop stops server's execution
func (server *HTTPServer) Stop(ctx context.Context) error {
	return server.Serv.Shutdown(ctx)
}

// Initialize perform the server initialization
func (server *HTTPServer) Initialize() http.Handler {
	log.Info().Msgf("Initializing HTTP server at '%s'", server.Config.Address)

	router := mux.NewRouter().StrictSlash(true)
	router.Use(httputils.LogRequest)

	server.addEndpointsToRouter(router)
	log.Info().Msgf("Server has been initiliazed")

	return router
}

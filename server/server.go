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

// Package server contains implementation of REST API server (HTTPServer) for the
// Insights content service. In current version, the following
// REST API endpoints are available:
package server

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// HTTPServer in an implementation of Server interface
type HTTPServer struct {
	Config Configuration
	Serv   *http.Server
}

// New constructs new implementation of Server interface
func New(config Configuration) *HTTPServer {
	return &HTTPServer{
		Config: config,
	}
}

// Start starts server
func (server *HTTPServer) Start() error {
	address := server.Config.Address
	log.Info().Msgf("Starting HTTP server at '%s'", address)
	router := server.Initialize(address)
	server.Serv = &http.Server{Addr: address, Handler: router}
	var err error

	if server.Config.UseHTTPS {
		err = server.Serv.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = server.Serv.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Unable to start HTTP/S server")
		return err
	}

	return nil
}

// Initialize perform the server initialization
func (server *HTTPServer) Initialize(address string) http.Handler {
	log.Info().Msgf("Initializing HTTP server at '%s'", address)

	router := mux.NewRouter().StrictSlash(true)
	// router.Use(server.LogRequest)

	apiPrefix := server.Config.APIPrefix

	openAPIURL := apiPrefix + filepath.Base(server.Config.APISpecFile)

	// enable authentication, but only if it is setup in configuration
	if server.Config.Auth {
		// we have to enable authentication for all endpoints, including endpoints
		// for Prometheus metrics and OpenAPI specification, because there is not
		// single prefix of other REST API calls. The special endpoints needs to
		// be handled in middleware which is not optimal
		noAuthURLs := []string{
			openAPIURL,
			openAPIURL + "?", // to be able to test using Frisby
		}
		router.Use(func(next http.Handler) http.Handler { return server.Authentication(next, noAuthURLs) })
	}

	server.addEndpointsToRouter(router)

	return router
}

func (server *HTTPServer) addEndpointsToRouter(router *mux.Router) {
	apiPrefix := server.Config.APIPrefix
	openAPIURL := apiPrefix + filepath.Base(server.Config.APISpecFile)

	// it is possible to use special REST API endpoints in debug mode
	// if server.Config.Debug {
	// 	server.addDebugEndpointsToRouter(router)
	// }

	// common REST API endpoints
	router.HandleFunc(apiPrefix+MainEndpoint, server.mainEndpoint).Methods(http.MethodGet)

	// OpenAPI specs
	router.HandleFunc(openAPIURL, server.serveAPISpecFile).Methods(http.MethodGet)
}

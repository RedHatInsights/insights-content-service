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
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"

	"github.com/RedHatInsights/insights-operator-utils/responses"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-content-service/groups"
)

// mainEndpoint will handle the requests for / endpoint
func (server *HTTPServer) mainEndpoint(writer http.ResponseWriter, _ *http.Request) {
	err := responses.SendOK(writer, responses.BuildOkResponse())
	if err != nil {
		log.Error().Err(err).Msg(responseDataError)
		handleServerError(err)
		return
	}
}

// listOfGroups returns the list of defined groups
func (server *HTTPServer) listOfGroups(writer http.ResponseWriter, request *http.Request) {
	if server.groupsList == nil {
		server.groupsList = make([]groups.Group, 0, len(server.Groups))

		for _, group := range server.Groups {
			server.groupsList = append(server.groupsList, group)
		}
	}

	err := responses.SendOK(writer, responses.BuildOkResponseWithData("groups", server.groupsList))
	if err != nil {
		log.Error().Err(err)
		handleServerError(err)
		return
	}
}

// infoMap returns map of additional information about this service
func (server *HTTPServer) infoMap(writer http.ResponseWriter, request *http.Request) {
	if server.InfoParams == nil {
		err := errors.New("InfoParams is empty")
		log.Error().Err(err)
		handleServerError(err)
		return
	}

	err := responses.SendOK(writer, responses.BuildOkResponseWithData("info", server.InfoParams))
	if err != nil {
		log.Error().Err(err)
		handleServerError(err)
		return
	}
}

// ruleContentStates returns status of all rules that have been read and parsed
func (server *HTTPServer) ruleContentStates(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	if query == nil {
		err := errors.New("Unable to retrieve Query object (should not happen)")
		log.Error().Err(err).Msg(responseDataError)
		handleServerError(err)
		return
	}

	// apply filters if specified on command line
	ruleContentStatusMap := filterStatusMap(server.ruleContentStatusMap, query)

	// log basic info about filtering results
	log.Info().
		Int("All rule states", len(server.ruleContentStatusMap)).
		Int("Filtered rule states", len(ruleContentStatusMap)).
		Msg("Rule content states filtering results")

	// try to send response with filtered rule states to client
	err := responses.SendOK(writer, responses.BuildOkResponseWithData("rules", ruleContentStatusMap))
	if err != nil {
		log.Error().Err(err)
		handleServerError(err)
		return
	}
}

// getStaticContent returns all the parsed rules' content
func (server *HTTPServer) getStaticContent(writer http.ResponseWriter, request *http.Request) {
	if server.encodedContent == nil {
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)

		if err := encoder.Encode(server.Content); err != nil {
			log.Error().Err(err).Msg("Cannot encode rules static content")
			handleServerError(err)
			return
		}

		server.encodedContent = buffer.Bytes()
	}

	err := responses.Send(http.StatusOK, writer, server.encodedContent)
	if err != nil {
		log.Error().Err(err)
		handleServerError(err)
		return
	}
}

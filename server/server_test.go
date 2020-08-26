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

package server_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-content-service/server"
	"github.com/RedHatInsights/insights-content-service/tests/helpers"
)

var config = server.Configuration{
	Address:     ":8080",
	APIPrefix:   "/api/test/",
	APISpecFile: "openapi.json",
	Debug:       true,
}

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// we need to be in the correct directory containing server.key and server.crt
	err := os.Chdir("../")
	if err != nil {
		panic(err)
	}
}

func checkResponseCode(t testing.TB, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// TestServerStartHTTP checks if it's possible to start regular HTTP server
func TestServerStartHTTP(t *testing.T) {
	contentDir := content.RuleContentDirectory{}
	helpers.RunTestWithTimeout(t, func(t testing.TB) {
		s := server.New(server.Configuration{
			// will use any free port
			Address:   ":0",
			APIPrefix: config.APIPrefix,
			Debug:     true,
		}, nil, contentDir)

		go func() {
			for {
				if s.Serv != nil {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			// doing some request to be sure server started successfully
			req, err := http.NewRequest(http.MethodGet, config.APIPrefix, nil)
			helpers.FailOnError(t, err)

			response := helpers.ExecuteRequest(s, req).Result()
			checkResponseCode(t, http.StatusOK, response.StatusCode)

			// stopping the server
			err = s.Stop(context.Background())
			helpers.FailOnError(t, err)
		}()

		err := s.Start()
		if err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}, 5*time.Second)
}

// TestServerStartError checks how/if errors are handled in server.Start method.
func TestServerStartError(t *testing.T) {
	contentDir := content.RuleContentDirectory{}
	testServer := server.New(server.Configuration{
		Address:   "localhost:99999",
		APIPrefix: "",
	}, nil, contentDir)

	err := testServer.Start()
	if err == nil {
		t.Fatal("Error should be reported")
	}
	if err.Error() != "listen tcp: address 99999: invalid port" {
		t.Fatal("Invalid error message:", err.Error())
	}
}

// TestServeAPISpecFileOK checks whether it is possible to access openapi.json via REST API server
func TestServeAPISpecFileOK(t *testing.T) {
	fileData, err := ioutil.ReadFile(config.APISpecFile)
	helpers.FailOnError(t, err)

	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodGet,
		Endpoint: config.APISpecFile,
	}, &helpers.APIResponse{
		StatusCode: http.StatusOK,
		Body:       string(fileData),
	})
}

// TestServeAPISpecOptionsMethod checks whether it is not possible to access openapi.json via REST API server using other HTTP methods
func TestServeAPISpecOptionsMethod(t *testing.T) {
	// HTTP methods to check
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}

	// check handling of all unsupported methods
	for _, method := range methods {
		helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
			Method:   method,
			Endpoint: config.APISpecFile,
		}, &helpers.APIResponse{
			StatusCode: http.StatusMethodNotAllowed,
		})
	}
}

// TestServeAPISpecFileError checks the error tests in REST API server handler
func TestServeAPISpecFileError(t *testing.T) {
	// openapi.json is really not there
	dirName, err := ioutil.TempDir("/tmp/", "")
	helpers.FailOnError(t, err)

	err = os.Chdir(dirName)
	helpers.FailOnError(t, err)

	err = os.Remove(dirName)
	helpers.FailOnError(t, err)

	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodGet,
		Endpoint: config.APISpecFile,
	}, &helpers.APIResponse{
		StatusCode: http.StatusOK,
	})
}

// TestServeAPIWrongEndpoint checks the REST API server behaviour in case wrong endpoint is used in request
func TestServeAPIWrongEndpoint(t *testing.T) {
	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodGet,
		Endpoint: "wrong_endpoint",
	}, &helpers.APIResponse{
		StatusCode: http.StatusNotFound,
	})
}

// TestServeListOfGroups checks the REST API server behaviour for group listing endpoint
func TestServeListOfGroups(t *testing.T) {
	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodGet,
		Endpoint: "groups",
	}, &helpers.APIResponse{
		StatusCode: http.StatusOK,
	})
}

// TestServeListOfGroupsOptionsMethod checks the REST API server behaviour for group listing endpoint
func TestServeListOfGroupsOptionsMethod(t *testing.T) {
	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodOptions,
		Endpoint: "groups",
	}, &helpers.APIResponse{
		StatusCode: http.StatusOK,
	})
}

// TestServerContent checks the REST API server behavior for content endpoint
func TestServerContent(t *testing.T) {
	helpers.AssertAPIRequest(t, &config, &helpers.APIRequest{
		Method:   http.MethodGet,
		Endpoint: "content",
	}, &helpers.APIResponse{
		StatusCode: http.StatusOK,
	})
}

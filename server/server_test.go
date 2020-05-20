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
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/RedHatInsights/insights-content-service/server"
	"github.com/RedHatInsights/insights-content-service/tests/helpers"
)

var config = server.Configuration{
	Address:     ":8080",
	APIPrefix:   "/api/test/",
	APISpecFile: "openapi.json",
	Debug:       true,
	UseHTTPS:    false,
}

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// checkServerStart test if the HTTP/HTTPs server can be started properly
func checkServerStart(t *testing.T, https bool) {
	helpers.RunTestWithTimeout(t, func(t *testing.T) {
		s := server.New(server.Configuration{
			// will use any free port
			Address:   ":0",
			APIPrefix: config.APIPrefix,
			Debug:     true,
			UseHTTPS:  https,
		}, nil)

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

			response := helpers.ExecuteRequest(s, req, &config).Result()
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

// TestServerStartHTTP checks if it's possible to start regular HTTP server
func TestServerStartHTTP(t *testing.T) {
	checkServerStart(t, false)
}

// TestServerStartHTTPs checks if it's possible to start HTTPs server
func TestServerStartHTTPs(t *testing.T) {
	// we need to be in the correct directory containing server.key and server.crt
	err := os.Chdir("../")
	if err != nil {
		t.Fatal(err)
	}
	checkServerStart(t, true)
}

// TestServerStartError checks how/if errors are handled in server.Start method.
func TestServerStartError(t *testing.T) {
	testServer := server.New(server.Configuration{
		Address:   "localhost:99999",
		APIPrefix: "",
	}, nil)

	err := testServer.Start()
	if err == nil {
		t.Fatal("Error should be reported")
	}
	if err.Error() != "listen tcp: address 99999: invalid port" {
		t.Fatal("Invalid error message:", err.Error())
	}
}

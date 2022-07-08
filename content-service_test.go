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

package main_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/insights-content-service"
	"github.com/RedHatInsights/insights-content-service/conf"
)

// checkStandardOutputStatus tests whether the standard output capturing was successful
func checkStandardOutputStatus(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// checkStandardOutputNotEmpty tests if standard output capturing captured at least some text
func checkStandardOutputNotEmpty(t *testing.T, captured string) {
	if captured == "" {
		t.Fatal("Output is empty")
	}
}

// checkHelpContent tests the help message displayed on standard output
func checkHelpContent(t *testing.T, captured string) {
	checkStandardOutputNotEmpty(t, captured)
	if !strings.HasPrefix(captured, "\nService to provide content for OCP rules") {
		t.Fatal("Unexpected help text")
	}
}

// checkVersionContent tests the help version info displayed on standard output
func checkVersionContent(t *testing.T, captured string) {
	checkStandardOutputNotEmpty(t, captured)
	if !strings.HasPrefix(captured, "Version:\t") {
		t.Fatal("Unexpected version info")
	}
}

// checkConfigContent tests the configuration info displayed on standard output
func checkConfigContent(t *testing.T, captured string) {
	checkStandardOutputNotEmpty(t, captured)
}

// checkUnknownCommand tests the unknown command message displayed on standard output
func checkUnknownCommand(t *testing.T, captured string) {
	checkStandardOutputNotEmpty(t, captured)
	if !strings.HasPrefix(captured, "\nCommand ") {
		t.Fatal("Unexpected error message")
	}
}

// TestPrintHelp is dummy ATM - we'll check the actual print content etc. in integration tests
func TestPrintHelp(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.PrintHelp()
	})
	checkStandardOutputStatus(t, err)
	checkHelpContent(t, captured)
}

// TestPrintVersionInfo is dummy ATM - we'll check versions etc. in integration tests
func TestPrintVersionInfo(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.PrintVersionInfo()
	})
	checkStandardOutputStatus(t, err)
	checkVersionContent(t, captured)
}

// TestPrintConfig is dummy ATM - we'll check config output etc. in integration tests
func TestPrintConfig(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.PrintConfig(conf.Config)
	})
	checkStandardOutputStatus(t, err)
	checkConfigContent(t, captured)
}

// TestHandleCommandHelp tests if proper output is printed for commands "help" and "print-help"
func TestHandleCommandHelp(t *testing.T) {
	helpCommands := []string{"help", "print-help"}
	for _, command := range helpCommands {
		captured, err := capture.StandardOutput(func() {
			main.HandleCommand(command)
		})
		checkStandardOutputStatus(t, err)
		checkHelpContent(t, captured)
	}
}

// TestHandleCommandConfig tests if proper output is printed for command "print-config"
func TestHandleCommandConfig(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.HandleCommand("print-config")
	})
	checkStandardOutputStatus(t, err)
	checkConfigContent(t, captured)
}

// TestHandleCommandVersion tests if proper output is printed for command "print-version-info"
func TestHandleCommandVersion(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.HandleCommand("print-version-info")
	})
	checkStandardOutputStatus(t, err)
	checkVersionContent(t, captured)
}

// TestHandleCommmandPrintGroups tests if proper output is printed for command "print-groups"
func TestHandleCommandPrintGroups(t *testing.T) {
	_, err := capture.StandardOutput(func() {
		main.HandleCommand("print-groups")
	})
	checkStandardOutputStatus(t, err)
}

// TestHandleCommandUnknownInput tests if proper output is printed for unknown command
func TestHandleCommandUnknownInput(t *testing.T) {
	captured, err := capture.StandardOutput(func() {
		main.HandleCommand("foo-bar-baz")
	})
	checkStandardOutputStatus(t, err)
	checkUnknownCommand(t, captured)
}

// TestInitInfoLog check the function initInfoLog
func TestInitInfoLog(t *testing.T) {
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	expectedString := "*** message ***"
	main.InitInfoLog(expectedString)

	logContent := buf.String()
	if !strings.Contains(logContent, expectedString) {
		t.Fatal("Inconsistent log content", logContent)
	}
}

// TestLogVersionInfo check the function logVersionInfo
func TestLogVersionInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	main.LogVersionInfo()

	logContent := buf.String()
	if !strings.Contains(logContent, "Build time:") {
		t.Fatal("Inconsistent log content", logContent)
	}
}

// TestPrintGroups check the behaviour of the printGroups function when no groups are configured
func TestPrintGroups(t *testing.T) {
	retval := int(main.PrintGroups())
	assert.Equal(t, main.ExitStatusServerError, retval)
}

// TestPrintRules check the behaviour of the printRules function when no rules are configured
func TestPrintRules(t *testing.T) {
	retval := int(main.PrintRules())
	assert.Equal(t, main.ExitStatusReadContentError, retval)
}

// TestFillInInfoParams test the behaviour of function fillInInfoParams
func TestFillInInfoParams(t *testing.T) {
	// map to be used by this unit test
	m := make(map[string]string)

	// preliminary test if Go Universe is still ok
	assert.Empty(t, m, "Map should be empty at the beginning")

	// try to fill-in all info params
	main.FillInInfoParams(m)

	// preliminary test if Go Universe is still ok
	assert.Len(t, m, 6, "Map should contains exactly six items")

	// does the map contain all expected keys?
	assert.Contains(t, m, "BuildVersion")
	assert.Contains(t, m, "BuildTime")
	assert.Contains(t, m, "BuildBranch")
	assert.Contains(t, m, "BuildCommit")
	assert.Contains(t, m, "UtilsVersion")
	assert.Contains(t, m, "OCPRulesVersion")
}

/*
Copyright © 2021 Red Hat, Inc.

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
	"testing"

	"github.com/RedHatInsights/insights-content-service/server"
	"github.com/RedHatInsights/insights-operator-utils/types"
)

// prepareStatusMap is a helper function to prepare map containing variour rule
// content states
func prepareStatusMap(includeInternal bool, includeExternal bool, includeErrorStates bool) map[string]types.RuleContentStatus {
	const externalRuleType = "external"
	const internalRuleType = "internal"

	ruleContentStatusMap := make(map[string]types.RuleContentStatus)

	if includeInternal {
		ruleContentStatusMap["rule1"] = types.RuleContentStatus{
			RuleType: types.RuleType(internalRuleType),
			Loaded:   true,
			Error:    "",
		}

		if includeErrorStates {
			ruleContentStatusMap["rule2"] = types.RuleContentStatus{
				RuleType: types.RuleType(internalRuleType),
				Loaded:   false,
				Error:    "internal rule3 parsing error",
			}
		}
	}

	if includeExternal {
		ruleContentStatusMap["rule3"] = types.RuleContentStatus{
			RuleType: types.RuleType(externalRuleType),
			Loaded:   true,
			Error:    "",
		}

		if includeErrorStates {
			ruleContentStatusMap["rule4"] = types.RuleContentStatus{
				RuleType: types.RuleType(externalRuleType),
				Loaded:   false,
				Error:    "external rule4 parsing error",
			}
		}
	}

	return ruleContentStatusMap
}

// prepareQuery is a helper function to prepare map containing query parameters
func prepareQuery(includeInternal bool, includeExternal bool, ruleNames []string) map[string][]string {
	query := make(map[string][]string)
	// add flag to filter internal rules
	if includeInternal {
		query["internal"] = []string{"x"} // any value is appropriate
	}

	// add flag to filter external rules
	if includeExternal {
		query["external"] = []string{"x"} // any value is appropriate
	}

	// add flag to filter rules by their name(s)
	if ruleNames != nil && len(ruleNames) >= 1 {
		query["rule"] = ruleNames
	}
	return query
}

// TestFilterStatusMapNoFilter tests the function filterStatusMap when no
// filtering is specified
func TestFilterStatusMapNoFilter(t *testing.T) {
	states := prepareStatusMap(true, true, true)
	query := prepareQuery(false, false, nil)
	filtered := server.FilterStatusMap(states, query)

	// quick check for filtered rule content states
	if len(filtered) != 4 {
		t.Fatal("All four rule states should be included in filtered map")
	}
}

// TestFilterStatusMapInternalRules tests the function filterStatusMap when
// internal rules filtering is enabled
func TestFilterStatusMapInternalRules(t *testing.T) {
	states := prepareStatusMap(true, true, true)
	query := prepareQuery(true, false, nil)
	filtered := server.FilterStatusMap(states, query)

	// quick check for filtered rule content states
	if len(filtered) != 2 {
		t.Fatal("Just internal rule states should be included in filtered map")
	}
}

// TestFilterStatusMapExternalRules tests the function filterStatusMap when
// external rules filtering is enabled
func TestFilterStatusMapExternalRules(t *testing.T) {
	states := prepareStatusMap(true, true, true)
	query := prepareQuery(false, true, nil)
	filtered := server.FilterStatusMap(states, query)

	// quick check for filtered rule content states
	if len(filtered) != 2 {
		t.Fatal("Just external rule states should be included in filtered map")
	}
}

// TestFilterStatusMapInternalAndExternalRules tests the function
// filterStatusMap when internal and also external rules filtering is enabled
func TestFilterStatusMapInternalAndExternalRules(t *testing.T) {
	states := prepareStatusMap(true, true, true)
	query := prepareQuery(true, true, nil)
	filtered := server.FilterStatusMap(states, query)

	// quick check for filtered rule content states
	if len(filtered) != 4 {
		t.Fatal("All internal and external rule states should be included in filtered map")
	}
}

// TestFilterStatusMapRuleName tests the function filterStatusMap when
// filtering by rule name is enabled
func TestFilterStatusMapRuleName(t *testing.T) {
	states := prepareStatusMap(true, true, true)
	query := prepareQuery(false, false, []string{"rule1"})
	filtered := server.FilterStatusMap(states, query)

	// quick check for filtered rule content states
	if len(filtered) != 1 {
		t.Fatal("Just the specified rule status should be included in filtered map")
	}

	_, found := filtered["rule1"]
	if !found {
		t.Fatal("Wrong filtered result!")
	}
}

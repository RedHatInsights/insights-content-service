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

package server

import (
	"strings"

	"github.com/RedHatInsights/insights-operator-utils/collections"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// filterStatusMap function apply various filters to map with all rule content
// states
func filterStatusMap(states map[string]types.RuleContentStatus, query map[string][]string) map[string]types.RuleContentStatus {
	// retrieve all possible filters

	// we are just interested about the presence of these two parameters,
	// not their values
	_, externalRuleFilter := query["external"]
	_, internalRuleFilter := query["internal"]

	// this parameter can be specified multiple times to allow client to
	// select multiple rules
	ruleNames, ruleNameFilter := query["rule"]

	// rule names can be specified multiple times, but for logging we need
	// just one slice of names
	ruleNamesStr := strings.Join(ruleNames, ",")

	log.Info().
		Bool("external rules filter", externalRuleFilter).
		Bool("internal rules filter", internalRuleFilter).
		Bool("rule name filter", ruleNameFilter).
		Str("rule names", ruleNamesStr).
		Msg("RuleContentStates endpoint")

	// should we perform filtering?
	if externalRuleFilter || internalRuleFilter || ruleNameFilter {
		result := make(map[string]types.RuleContentStatus)

		// iterate over all states
		for name, value := range states {
			// filter external rules if such filter is defined
			if externalRuleFilter && value.RuleType == "external" {
				result[name] = value
			}
			// filter internal rules if such filter is defined
			if internalRuleFilter && value.RuleType == "internal" {
				result[name] = value
			}
			// filter by rule name if such filter is defined
			if ruleNameFilter && collections.StringInSlice(name, ruleNames) {
				result[name] = value
			}
		}
		return result
	}

	// no filtering needed -> return the original map
	return states
}

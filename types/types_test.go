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

package types_test

import (
	"testing"

	internal_types "github.com/RedHatInsights/insights-content-service/types"
	"github.com/stretchr/testify/assert"
)

func TestToErrorKeyMetadata(t *testing.T) {
	receivedErroKey := internal_types.ReceivedErrorKeyMetadata{
		Description:    "test description",
		Impact:         "test impact",
		Likelihood:     2,
		PublishDate:    "12/08/1988",
		ResolutionRisk: "Cluster Node Restart",
		Status:         "test status",
		Tags:           []string{"test tag 0", "test tag 1"},
	}

	testImpactDict := map[string]int{
		"test impact": 5,
	}

	testResolutionRiskDict := map[string]int{
		"Cluster Node Restart": 2,
	}

	testResult := receivedErroKey.ToErrorKeyMetadata(testImpactDict, testResolutionRiskDict)
	assert.Equal(t, "test description", testResult.Description)
	assert.Equal(t, 5, testResult.Impact.Impact)
	assert.Equal(t, 2, testResult.Likelihood)
	assert.Equal(t, "12/08/1988", testResult.PublishDate)
	assert.Equal(t, "test status", testResult.Status)
	assert.Equal(t, 2, testResult.ResolutionRisk)
	assert.Equal(t, "test tag 0", testResult.Tags[0])
	assert.Equal(t, "test tag 1", testResult.Tags[1])
}

func TestToErrorKeyMetadataNonexistentKeys(t *testing.T) {
	receivedErroKey := internal_types.ReceivedErrorKeyMetadata{
		Description:    "test description",
		Impact:         "non existent impact",
		Likelihood:     2,
		PublishDate:    "12/08/1988",
		ResolutionRisk: "non existent resolution_risk",
		Status:         "test status",
		Tags:           []string{"test tag 0", "test tag 1"},
	}

	testImpactDict := map[string]int{
		"test impact": 5,
	}

	testResolutionRiskDict := map[string]int{
		"Cluster Node Restart": 2,
	}

	testResult := receivedErroKey.ToErrorKeyMetadata(testImpactDict, testResolutionRiskDict)
	assert.Equal(t, "test description", testResult.Description)
	assert.Equal(t, 0, testResult.Impact.Impact)
	assert.Equal(t, 2, testResult.Likelihood)
	assert.Equal(t, "12/08/1988", testResult.PublishDate)
	assert.Equal(t, "test status", testResult.Status)
	assert.Equal(t, 0, testResult.ResolutionRisk)
	assert.Equal(t, "test tag 0", testResult.Tags[0])
	assert.Equal(t, "test tag 1", testResult.Tags[1])
}

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

// Package types contains declaration of various data types (usually structures)
// used elsewhere in the aggregator code.
package types

import (
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// OrgID represents organization ID
type OrgID uint32

// ClusterName represents name of cluster in format c8590f31-e97e-4b85-b506-c45ce1911a12
type ClusterName string

// Timestamp represents any timestamp in a form gathered from database
// TODO: need to be improved
type Timestamp string

// UserID represents type for user id
type UserID string

// ReceivedErrorKeyMetadata is ErrorKeyMetadata as received from
// the metadata.yaml file
type ReceivedErrorKeyMetadata struct {
	Description string   `yaml:"description" json:"description"`
	Impact      string   `yaml:"impact" json:"impact"`
	Likelihood  int      `yaml:"likelihood" json:"likelihood"`
	PublishDate string   `yaml:"publish_date" json:"publish_date"`
	Status      string   `yaml:"status" json:"status"`
	Tags        []string `yaml:"tags" json:"tags"`
}

// ToErrorKeyMetadata converts ReceivedErrorKeyMetadata to the type actually
// used by the pipeline, calculating impact with impactDict
func (r ReceivedErrorKeyMetadata) ToErrorKeyMetadata(impactDict map[string]int) types.ErrorKeyMetadata {
	returnVal := types.ErrorKeyMetadata{}
	returnVal.Description = r.Description
	impactNumber, found := impactDict[r.Impact]
	if !found {
		log.Error().Msgf(`impact "%v" doesn't have integer representation' (skipping)`, r.Impact)
	}
	returnVal.Impact.Impact = impactNumber
	returnVal.Impact.Name = r.Impact
	returnVal.Likelihood = r.Likelihood
	returnVal.PublishDate = r.PublishDate
	returnVal.Status = r.Status
	returnVal.Tags = r.Tags
	return returnVal
}

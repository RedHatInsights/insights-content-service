package main

import (
	"github.com/RedHatInsights/insights-content-service/groups"
	"github.com/RedHatInsights/insights-results-aggregator/content"
	"github.com/rs/zerolog/log"
)

func main() {
	groupCfg, err := groups.ParseGroupConfigFile("./groups_config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse group config file")
	}

	ruleContentDir, err := content.ParseRuleContentDir("../ccx-rules-ocp/content/")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse group config file")
	}

	// For every rule.
	for ruleName, ruleContent := range ruleContentDir.Rules {
		// For every error code of that rule.
		for errCode, errContent := range ruleContent.ErrorKeys {
			errGroups := map[string]string{}

			// For every tag of that error code.
			for _, errTag := range errContent.Metadata.Tags {
				// Check for duplicate tags in the error code's content.
				if _, exists := errGroups[errTag]; exists {
					log.Error().Msgf("duplicate tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
				}

				// Find a group with the tag.
				for _, group := range groupCfg {
					for _, tag := range group.Tags {
						if tag == errTag {
							errGroups[errTag] = group.Name
							break
						}
					}
				}

				// Check if at least one group with the tag was found.
				if _, exists := errGroups[errTag]; !exists {
					log.Error().Msgf("invalid tag '%s' in content of '%s|%s'", errTag, ruleName, errCode)
				}
			}

			log.Info().Msgf("%s|%s: %v", ruleName, errCode, errGroups)
		}
	}
}

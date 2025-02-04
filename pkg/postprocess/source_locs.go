// Copyright 2022 Snyk Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postprocess

import (
	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/models"
)

// Annotate a report with source location information
func AddSourceLocs(
	results *models.Results,
	configurations input.Loader,
) {
	for _, inputResult := range results.Results {
		// Retrieve the filepath of the input state by looking in the metadata.
		filepath, haveFilepath := inputResult.Input.Meta["filepath"].(string)
		if !haveFilepath {
			continue
		}

		// Annotate resources in input state
		for _, resources := range inputResult.Input.Resources {
			for rk, resource := range resources {
				location := getResourceSourceLoc(
					configurations,
					filepath,
					resource.Namespace,
					resource.ResourceType,
					resource.Id,
				)
				if resource.Meta == nil {
					resource.Meta = map[string]interface{}{}
				}
				if len(location) > 0 {
					resource.Meta["location"] = location
				}
				resources[rk] = resource
			}
		}

		// Annotate resources in rule results.
		for _, ruleResult := range inputResult.RuleResults {
			for _, result := range ruleResult.Results {
				addSourceLocsToRuleResult(
					configurations,
					filepath,
					result,
				)
			}
		}
	}
}

func addSourceLocsToRuleResult(
	configurations input.Loader,
	filepath string,
	result models.RuleResult,
) {
	for _, resource := range result.Resources {
		location := getResourceSourceLoc(
			configurations,
			filepath,
			resource.Namespace,
			resource.Type,
			resource.Id,
		)
		resource.Location = location

		for i := range resource.Attributes {
			attributePath := []interface{}{resource.Namespace, resource.Type, resource.Id}
			attributePath = append(attributePath, resource.Attributes[i].Path...)
			location, _ := configurations.Location(filepath, attributePath)
			if len(location) > 0 {
				loc := toLocation(location[0])
				resource.Attributes[i].Location = &loc
			}
		}
	}
}

func getResourceSourceLoc(
	configurations input.Loader,
	filepath string,
	resourceNamespace string,
	resourceType string,
	resourceId string,
) []models.SourceLocation {
	resourcePath := []interface{}{resourceNamespace, resourceType, resourceId}
	resourceLocs, _ := configurations.Location(filepath, resourcePath)
	if resourceLocs == nil {
		return nil
	}
	locations := []models.SourceLocation{}
	for _, loc := range resourceLocs {
		locations = append(locations, toLocation(loc))
	}
	return locations
}

func toLocation(loc input.Location) models.SourceLocation {
	return models.SourceLocation{
		Filepath: loc.Path,
		Line:     loc.Line,
		Column:   loc.Col,
	}
}

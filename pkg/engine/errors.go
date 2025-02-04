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

package engine

import (
	"errors"
)

// FailedToLoadRegoAPI indicates that an error occurred while initializing the snyk
// Rego API.
var FailedToLoadRegoAPI = errors.New("Failed to load the snyk Rego API")

// FailedToLoadRules indicates that an error occurred while consuming the rego and data
// producers provided to the engine.
var FailedToLoadRules = errors.New("Failed to load rules")

// FailedToCompile indicates that more than the maximum number of errors occurred during
// the compilation stage.
var FailedToCompile = errors.New("Failed to compile rules")

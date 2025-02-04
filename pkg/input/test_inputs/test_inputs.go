// Copyright 2022 Snyk Ltd
// Copyright 2021 Fugue, Inc.
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

package test_inputs

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed data
var data embed.FS

func Contents(t *testing.T, name string) []byte {
	contents, err := data.ReadFile(filepath.Join("data", name))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return contents
}

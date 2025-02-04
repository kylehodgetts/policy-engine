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

package input

import (
	"fmt"
	"strings"
)

// Type represents one or more types of inputs.
type Type struct {
	// Name is the primary name for this input type. This is the field to use when input
	// types need to be serialized to a string.
	Name string
	// Aliases are alternate, case-insensitive names for this input type.
	Aliases []string
	// Children are input types encompassed by this input type. This field can be used
	// to define aggregate input types.
	Children Types
}

// Matches returns true if the name of this input type or any of its children exactly
// match the given input type string.
func (t *Type) Matches(inputType string) bool {
	if t.Name == inputType {
		return true
	}
	for _, c := range t.Children {
		if c.Matches(inputType) {
			return true
		}
	}
	return false
}

// Returns true if this Type instance is exactly equal to other
func (t *Type) Equals(other *Type) bool {
	if t == other {
		return true
	}
	if !(t.Name == other.Name) {
		return false
	}
	if !(len(t.Aliases) == len(other.Aliases)) {
		return false
	}
	for idx, a := range t.Aliases {
		if a != other.Aliases[idx] {
			return false
		}
	}
	return t.Children.Equals(other.Children)
}

// Types is a slice of Type struct.
type Types []*Type

// Returns true if this Types instance is exactly equal to other
func (t Types) Equals(other Types) bool {
	if len(t) != len(other) {
		return false
	}
	for idx, a := range t {
		if !a.Equals(other[idx]) {
			return false
		}
	}
	return true
}

// FromString returns the first InputType where either its name or aliases match the
// given input type string. This method is case-insensitive.
func (t Types) FromString(inputType string) (*Type, error) {
	inputType = strings.ToLower(inputType)
	for _, i := range t {
		if strings.ToLower(i.Name) == inputType {
			return i, nil
		}
		for _, a := range i.Aliases {
			if strings.ToLower(a) == inputType {
				return i, nil
			}
		}
	}
	return nil, fmt.Errorf("Unrecognized input type")
}

// Arm represents Azure Resource Manager template inputs.
var Arm = &Type{
	Name: "arm",
}

// CloudFormation represents CloudFormation template inputs.
var CloudFormation = &Type{
	Name:    "cfn",
	Aliases: []string{"cloudformation"},
}

// CloudScan represents inputs from a Snyk Cloud Scan.
var CloudScan = &Type{
	Name:    "cloud_scan",
	Aliases: []string{"cloud-scan"},
	Children: Types{
		TerraformState,
	},
}

// Kubernetes represents Kubernetes manifest inputs.
var Kubernetes = &Type{
	Name:    "k8s",
	Aliases: []string{"kubernetes"},
}

// TerraformHCL represents Terraform HCL source code inputs.
var TerraformHCL = &Type{
	Name:    "tf_hcl",
	Aliases: []string{"tf-hcl"},
}

// TerraformPlan represents Terraform Plan JSON inputs.
var TerraformPlan = &Type{
	Name:    "tf_plan",
	Aliases: []string{"tf-plan"},
}

// TerraformState represents Terraform State JSON inputs.
var TerraformState = &Type{
	Name:    "tf_state",
	Aliases: []string{"tf-state"},
}

// StreamlinedState is a temporary addition until we're able to completely replace the
// old streamlined state format.
var StreamlinedState = &Type{
	Name:    "streamlined_state",
	Aliases: []string{"streamlined-state"},
}

// Terraform is an aggregate input type that encompasses all input types that contain
// Terraform resource types.
var Terraform = &Type{
	Name:    "tf",
	Aliases: []string{"terraform"},
	Children: Types{
		TerraformHCL,
		TerraformPlan,
		CloudScan,
	},
}

// Auto is an aggregate type that contains all of the IaC input types that this package
// supports.
var Auto = &Type{
	Name: "auto",
	Children: Types{
		Arm,
		CloudFormation,
		Kubernetes,
		TerraformHCL,
		TerraformPlan,
		TerraformState,
	},
}

// Any is an aggregate type that contains all known input types.
var Any = &Type{
	Name: "any",
	Children: Types{
		Arm,
		CloudFormation,
		Kubernetes,
		Terraform,
	},
}

// SupportedInputTypes contains all of the input types that this package has detectors
// for.
var SupportedInputTypes = Types{
	Auto,
	Arm,
	CloudFormation,
	Kubernetes,
	TerraformHCL,
	TerraformPlan,
	TerraformState,
	StreamlinedState,
}

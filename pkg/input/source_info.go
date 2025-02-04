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

// This file contains some utilities to deal with extracting source code
// information from generic JSON / YAML files.

package input

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type SourceInfoNode struct {
	key  *yaml.Node // Possibly nil
	body *yaml.Node
}

func LoadSourceInfoNode(contents []byte) (*SourceInfoNode, error) {
	multi, err := LoadMultiSourceInfoNode(contents)
	if err != nil {
		return nil, err
	}
	return &multi[0], nil
}

// LoadMultiSourceInfoNode parses YAML documents with multiple entries, or
// normal single YAML/JSON documents.
func LoadMultiSourceInfoNode(contents []byte) ([]SourceInfoNode, error) {
	dec := yaml.NewDecoder(bytes.NewReader(contents))
	var documents []*yaml.Node
	for {
		value := yaml.Node{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if value.Kind == yaml.DocumentNode {
			for _, doc := range value.Content {
				documents = append(documents, doc)
			}
		} else {
			documents = append(documents, &value)
		}
	}

	if len(documents) < 1 {
		return nil, fmt.Errorf("No document contents")
	}

	nodes := []SourceInfoNode{}
	for _, doc := range documents {
		nodes = append(nodes, SourceInfoNode{body: doc})
	}
	return nodes, nil
}

func (node *SourceInfoNode) GetKey(key string) (*SourceInfoNode, error) {
	if node.body.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("Expected %s but got %s", prettyKind(yaml.MappingNode), prettyKind(node.body.Kind))
	}

	for i := 0; i+1 < len(node.body.Content); i += 2 {
		if node.body.Content[i].Value == key {
			return &SourceInfoNode{key: node.body.Content[i], body: node.body.Content[i+1]}, nil
		}
	}

	return nil, fmt.Errorf("Key %s not found", key)
}

func (node *SourceInfoNode) GetIndex(index int) (*SourceInfoNode, error) {
	if node.body.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("Expected %s but got %s", prettyKind(yaml.SequenceNode), prettyKind(node.body.Kind))
	}

	if index < 0 || index >= len(node.body.Content) {
		return nil, fmt.Errorf("Index %d out of bounds for length %d", index, len(node.body.Content))
	}

	return &SourceInfoNode{body: node.body.Content[index]}, nil
}

// GetPath tries to retrieve a path as far as possible.
func (node *SourceInfoNode) GetPath(path []interface{}) (*SourceInfoNode, error) {
	if len(path) == 0 {
		return node, nil
	}

	switch node.body.Kind {
	case yaml.MappingNode:
		key, ok := path[0].(string)
		if !ok {
			return node, fmt.Errorf("Expected string key")
		}
		child, err := node.GetKey(key)
		if err != nil {
			return node, err
		} else {
			return child.GetPath(path[1:])
		}
	case yaml.SequenceNode:
		index, ok := path[0].(int)
		if !ok {
			return node, fmt.Errorf("Expected int index")
		}

		child, err := node.GetIndex(index)
		if err != nil {
			return node, err
		} else {
			return child.GetPath(path[1:])
		}
	default:
		return node, fmt.Errorf("Expected %s or %s at key %s but got %s", prettyKind(yaml.MappingNode), prettyKind(yaml.SequenceNode), path[0], prettyKind(node.body.Kind))
	}
}

func (node *SourceInfoNode) Location() (int, int) {
	if node.key != nil {
		return node.key.Line, node.key.Column
	} else {
		return node.body.Line, node.body.Column
	}
}

func prettyKind(kind yaml.Kind) string {
	switch kind {
	case yaml.MappingNode:
		return "map"
	case yaml.SequenceNode:
		return "array"
	case yaml.DocumentNode:
		return "doc"
	case yaml.AliasNode:
		return "alias"
	case yaml.ScalarNode:
		return "scalar"
	default:
		return "unknown"
	}
}

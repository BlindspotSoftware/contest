// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// JobDescFormat defines a type for the supported formats for job descriptor configurations.
type JobDescFormat int

// List of supported job descriptor formats
const (
	JobDescFormatUnknown JobDescFormat = iota
	JobDescFormatJSON
	JobDescFormatYAML
)

// ParseJobDescriptor validates a job descriptor's well-formedness, and returns a
// JSON-formatted descriptor if it was provided in a different format.
// The currently supported format are JSON and YAML.
func ParseJobDescriptor(data []byte, jobDescFormat JobDescFormat) ([]byte, error) {
	jobDesc := make(map[string]interface{})
	switch jobDescFormat {
	case JobDescFormatJSON:
		if err := json.Unmarshal(data, &jobDesc); err != nil {
			return nil, fmt.Errorf("failed to parse JSON job descriptor: %w", err)
		}
	case JobDescFormatYAML:
		if err := yaml.Unmarshal(data, &jobDesc); err != nil {
			return nil, fmt.Errorf("failed to parse YAML job descriptor: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown job descriptor format")
	}

	// then marshal the structure back to JSON
	jobDescJSON, err := json.MarshalIndent(jobDesc, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize job descriptor to JSON: %w", err)
	}
	return jobDescJSON, nil
}

func JobDescFormatFromString(jobDescFormat string) JobDescFormat {
	switch {
	case strings.ToLower(jobDescFormat) == "json":
		return JobDescFormatJSON
	case strings.ToLower(jobDescFormat) == "yaml":
		return JobDescFormatYAML
	default:
		return JobDescFormatUnknown
	}
}

func (j JobDescFormat) String() string {
	switch j {
	case JobDescFormatJSON:
		return "json"
	case JobDescFormatYAML:
		return "yaml"
	default:
		return "unknown"
	}
}

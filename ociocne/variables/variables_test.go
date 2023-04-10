// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package variables

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeVersion(t *testing.T) {
	var tests = []struct {
		input  string
		output string
	}{
		{
			"1-2-4",
			"1-2-4",
		},
		{
			"v1.25.8",
			"v1-25-8",
		},
		{
			"v1.24.11+el1",
			"v1-24-11-el1",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Replace %s", tt.input), func(t *testing.T) {
			assert.Equal(t, tt.output, sanitizeK8sVersion(tt.input))
		})
	}
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package variables

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashString(t *testing.T) {
	vars := &Variables{
		ControlPlaneVolumeGbs: DefaultVolumeGbs,
		ControlPlaneMemoryGbs: DefaultMemoryGbs,
		NodeVolumeGbs:         DefaultVolumeGbs,
		NodeMemoryGbs:         DefaultMemoryGbs,
	}
	varsHashPresent := vars
	varsHashPresent.Hash = "xyz"
	var tests = []struct {
		name  string
		va    *Variables
		vb    *Variables
		equal bool
	}{
		{
			"Equal hashes when equal objects",
			vars,
			vars,
			true,
		},
		{
			"Equal hashes when hash already computed on equal objects",
			varsHashPresent,
			vars,
			true,
		},
		{
			"Different hashes when different objects",
			vars,
			&Variables{
				ControlPlaneVolumeGbs: DefaultVolumeGbs,
				ControlPlaneMemoryGbs: DefaultMemoryGbs,
				NodeVolumeGbs:         DefaultVolumeGbs,
				NodeMemoryGbs:         128,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha, err := tt.va.HashSum()
			assert.NoError(t, err)
			hb, err := tt.vb.HashSum()
			assert.NoError(t, err)
			if tt.equal {
				assert.Equal(t, ha, hb)
			} else {
				assert.NotEqual(t, ha, hb)
			}
		})
	}
}

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
				ControlPlaneReplicas:  1,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.va.SetHashes()
			tt.vb.SetHashes()
			if tt.equal {
				assert.Equal(t, tt.va.ControlPlaneHash, tt.vb.ControlPlaneHash)
			} else {
				assert.NotEqual(t, tt.va.ControlPlaneHash, tt.vb.ControlPlaneHash)
			}
		})
	}
}

func TestParseNodePools(t *testing.T) {
	v := &Variables{
		RawNodePools: []string{
			"{\"name\":\"np-1\",\"replicas\":2,\"memory\":16,\"ocpus\":1,\"volumeSize\":50,\"shape\":\"VM.Standard.E4.Flex\"}",
			"{\"name\":\"np-2\",\"replicas\":4,\"memory\":64,\"ocpus\":8,\"volumeSize\":250,\"shape\":\"VM.Standard.E4.Flex\"}",
		},
	}

	nps, err := v.ParseNodePools()
	assert.NoError(t, err)
	assert.Len(t, nps, 2)

	np1 := nps[0]
	np2 := nps[1]
	assert.Equal(t, np1.Name, "np-1")
	assert.Equal(t, np2.Name, "np-2")
}

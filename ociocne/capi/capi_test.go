// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"testing"
)

func TestRenderObjects(t *testing.T) {
	v := variables.Variables{
		Name:                    "xyz",
		CompartmentID:           "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ImageID:                 "ocid1.image.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		VCNID:                   "ocid1.vcn.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WorkerNodeSubnet:        "ocid1.subnet.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ControlPlaneSubnet:      "ocid1.subnet.oc1.iad.yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
		LoadBalancerSubnet:      "ocid1.subnet.oc1.iad.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		SSHPublicKey:            "ssh-rsa aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa foo@foo-mac",
		ControlPlaneReplicas:    1,
		NodeReplicas:            1,
		NodePVTransitEncryption: true,
		NodeShape:               "VM.Standard.E4.Flex",
		ControlPlaneShape:       "VM.Standard.E4.Flex",
		KubernetesVersion:       "v1.24.8",
		TigeraTag:               variables.DefaultTigeraTag,
		CalicoRegistry:          variables.DefaultRegistryCNE,
		CCMImage:                variables.DefaultCCMImage,
		NodeOCPUs:               1,
		ControlPlaneOCPUs:       1,
		NodeMemoryGbs:           16,
		ControlPlaneMemoryGbs:   16,
		PodCIDR:                 "192.168.0.0/16",
		ClusterCIDR:             "10.0.0.0/12",
		Fingerprint:             "fingerprint",
		PrivateKey:              "xyz",
		PrivateKeyPassphrase:    "",
		Region:                  "xyz",
		Tenancy:                 "xyz",
		User:                    "xyz",
		ProviderId:              variables.ProviderId,
	}

	for _, o := range createObjects(&v) {
		u, err := loadTextTemplate(o, v)
		assert.NoError(t, err)
		assert.NotNil(t, u)
	}
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package version

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

const testCMData = `v1.24.8:
  Release: "1.5"
  container-images:
    calico: v3.25.0
    coredns: 1.8.6
    etcd: 3.5.3
    tigera-operator: v1.29.0
v1.25.7:
  Release: "1.6"
  container-images:
    calico: v3.25.0
    coredns: v1.9.3
    etcd: 3.5.6
    tigera-operator: v1.29.0`

func TestLoadDefaults(t *testing.T) {
	ki := fake.NewSimpleClientset(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ocneConfigMapName,
			Namespace: ocneConfigMapNamespace,
		},
		Data: map[string]string{
			"mapping": testCMData,
		},
	})

	defaults, err := LoadDefaults(context.TODO(), ki)
	assert.NoError(t, err)
	assert.Equal(t, "v1.25.7", defaults.KubernetesVersion)
	assert.Equal(t, "v1.9.3", defaults.ContainerImages.CoreDNS)
	assert.Equal(t, "v1.29.0", defaults.ContainerImages.TigeraOperator)
	assert.Equal(t, "3.5.6", defaults.ContainerImages.ETCD)
}

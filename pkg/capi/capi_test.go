// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fake2 "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"math/rand"
	"testing"
	"time"
)

const (
	testName = "test"

	testMachine = `apiVersion: cluster.x-k8s.io/v1beta1
kind: Machine
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: test
    cluster.x-k8s.io/control-plane: ""
    cluster.x-k8s.io/control-plane-name: test
  namespace: test`
	testKey = `aaa
bbb
ccc
ddd`
)

var (
	testMachineGVR = schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "machines",
	}
	testCAPIClient = &CAPIClient{
		capiTimeout:         0 * time.Second,
		capiPollingInterval: 0 * time.Second,
	}

	testVariables = &variables.Variables{
		Name:              testName,
		Namespace:         testName,
		CloudCredentialId: "cattle-global-data:admin-creds",
		Tenancy:           "t",
		User:              "u",
		PrivateKey:        "k",
		Region:            "r",

		InstallCalico:     true,
		InstallCCM:        true,
		InstallVerrazzano: true,
	}
)

func TestNewCAPIClient(t *testing.T) {
	c := NewCAPIClient()
	z := 0 * time.Second
	assert.Greater(t, c.capiTimeout, z)
	assert.Greater(t, c.capiPollingInterval, z)
	assert.Greater(t, c.verrazzanoTimeout, z)
	assert.Greater(t, c.verrazzanoPollingInterval, z)
}

func TestCreateOrUpdateAllObjects(t *testing.T) {
	ki := fake.NewSimpleClientset()
	di := createTestDIWithClusterAndMachine()
	_, err := testCAPIClient.CreateOrUpdateAllObjects(context.TODO(), ki, di, testVariables)
	assert.NoError(t, err)
}

func TestRenderObjects(t *testing.T) {
	v := variables.Variables{
		Name:                    "xyz",
		Namespace:               "xyz",
		CompartmentID:           "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ImageID:                 "ocid1.image.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		VCNID:                   "ocid1.vcn.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WorkerNodeSubnet:        "ocid1.subnet.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ControlPlaneSubnet:      "ocid1.subnet.oc1.iad.yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
		LoadBalancerSubnet:      "ocid1.subnet.oc1.iad.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		SSHPublicKey:            "ssh-rsa aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa foo@foo-mac",
		ControlPlaneReplicas:    1,
		NodePVTransitEncryption: true,
		ControlPlaneShape:       "VM.Standard.E4.Flex",
		KubernetesVersion:       "v1.24.8",
		TigeraTag:               "123",
		CNEPath:                 variables.DefaultCNEPath,
		ControlPlaneOCPUs:       1,
		ControlPlaneMemoryGbs:   16,
		PodCIDR:                 "192.168.0.0/16",
		ClusterCIDR:             "10.0.0.0/12",
		Fingerprint:             "fingerprint",
		PrivateKey:              testKey,
		PrivateKeyPassphrase:    "",
		Region:                  "xyz",
		Tenancy:                 "xyz",
		User:                    "xyz",
		Hash:                    "xyz",
		ProviderId:              variables.ProviderId,

		NodePools: []variables.NodePool{
			{
				Name:       "np-1",
				Replicas:   1,
				Memory:     32,
				Ocpus:      4,
				VolumeSize: 100,
				Shape:      "VM.E4.Standard.Flex",
			},
			{
				Name:       "np-2",
				Replicas:   2,
				Memory:     64,
				Ocpus:      8,
				VolumeSize: 250,
				Shape:      "xyz",
			},
		},

		InstallVerrazzano: true,
		InstallCCM:        true,
		InstallCalico:     true,
	}

	os := object.CreateObjects(&v)
	for _, o := range os {
		u, err := loadTextTemplate(o, v)
		assert.NoError(t, err)
		assert.NotNil(t, u)
	}
}

func TestDeleteCluster(t *testing.T) {
	cluster := createTestCluster(testVariables, true, true, clusterPhaseProvisioned)
	var tests = []struct {
		name string
		di   dynamic.Interface
	}{
		{
			"delete no cluster",
			fake2.NewSimpleDynamicClient(runtime.NewScheme()),
		},
		{
			"delete with cluster",
			fake2.NewSimpleDynamicClient(runtime.NewScheme(), cluster),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteCluster(context.TODO(), tt.di, testVariables)
			assert.NoError(t, err)
		})
	}
}

func TestWaitForCAPIClusterReady(t *testing.T) {
	runningMachine := createTestMachine(testVariables, machinePhaseRunning)
	notRunningMachine := createTestMachine(testVariables, "Error")
	cluster := createTestCluster(testVariables, true, true, clusterPhaseProvisioned)
	clusterCPNotReady := createTestCluster(testVariables, false, true, clusterPhaseProvisioned)
	clusterINotReady := createTestCluster(testVariables, true, false, clusterPhaseProvisioned)
	clusterError := createTestCluster(testVariables, true, true, "Error")

	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(listGVK(runningMachine), &unstructured.UnstructuredList{})
	var tests = []struct {
		name  string
		di    dynamic.Interface
		error bool
	}{
		{
			"ready when cluster and machine are ready",
			createTestDIWithClusterAndMachine(),
			false,
		},
		{
			"not ready when cluster ready but not all machines not ready",
			fake2.NewSimpleDynamicClient(scheme, cluster, runningMachine, notRunningMachine),
			true,
		},
		{
			"not ready when cluster controlplane not ready",
			fake2.NewSimpleDynamicClient(scheme, clusterCPNotReady, runningMachine),
			true,
		},
		{
			"not ready when cluster infrastructure not ready",
			fake2.NewSimpleDynamicClient(scheme, clusterINotReady, runningMachine),
			true,
		},
		{
			"not ready when cluster phase not ready",
			fake2.NewSimpleDynamicClient(scheme, clusterError, runningMachine),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testCAPIClient.WaitForCAPIClusterReady(context.TODO(), tt.di, testVariables)
			if (err != nil) != tt.error {
				t.Error(err)
			}
		})
	}
}

func listGVK(u *unstructured.Unstructured) schema.GroupVersionKind {
	gvk := u.GroupVersionKind()
	gvk.Kind = gvk.Kind + "List"
	return gvk
}

func createTestMachine(v *variables.Variables, phase string) *unstructured.Unstructured {
	machine, err := loadTextTemplate(object.Object{
		Text: testMachine,
	}, *v)
	if err != nil {
		panic(err)
	}
	machine[0].Object["status"] = map[string]interface{}{
		"phase": phase,
	}
	machine[0].Object["metadata"].(map[string]interface{})["name"] = fmt.Sprintf("m-%d", rand.Intn(10000))
	return &machine[0]
}

func createTestCluster(v *variables.Variables, cReady, iReady bool, phase string) *unstructured.Unstructured {
	cluster, err := loadTextTemplate(object.CAPICluster, *v)
	if err != nil {
		panic(err)
	}
	cluster[0].Object["status"] = map[string]interface{}{
		"controlPlaneReady":   cReady,
		"infrastructureReady": iReady,
		"phase":               phase,
	}
	return &cluster[0]
}

func createTestDIWithClusterAndMachine() dynamic.Interface {
	cluster := createTestCluster(testVariables, true, true, clusterPhaseProvisioned)
	machine := createTestMachine(testVariables, machinePhaseRunning)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(listGVK(machine), &unstructured.UnstructuredList{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   gvr.MachineDeployment.Group,
		Version: gvr.MachineDeployment.Version,
		Kind:    "MachineDeploymentList",
	}, &unstructured.UnstructuredList{})
	di := fake2.NewSimpleDynamicClient(scheme, cluster, machine)
	return di
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func createObjects(v *variables.Variables) []object {
	var res []object

	// Create addons if they are enabled
	if v.InstallCalico {
		res = append(res, cni...)
	}
	if v.InstallCCM {
		res = append(res, ccm...)
	}
	if v.InstallCSI {
		res = append(res, csi...)
	}
	if v.InstallVerrazzano {
		res = append(res, vpo...)
	}
	res = append(res, capi...)
	return res
}

type object struct {
	gvr  schema.GroupVersionResource
	text string
}

var vpo = []object{
	{gvr.ConfigMap, templates.VPOConfigMap},
	{gvr.ClusterResourceSet, templates.VPOResourceSet},
}

var csi = []object{
	{gvr.ConfigMap, templates.CSIConfigMap},
	{gvr.ClusterResourceSet, templates.CSIResourceSet},
}

var ccm = []object{
	{gvr.ConfigMap, templates.CCMConfigMap},
	{gvr.ClusterResourceSet, templates.CCMResourceSet},
}

var cni = []object{
	{gvr.ConfigMap, templates.CalicoConfigMap},
	{gvr.ClusterResourceSet, templates.CalicoResourceSet},
}

var capi = []object{
	{gvr.Cluster, templates.Cluster},
	{gvr.ClusterIdentity, templates.ClusterIdentity},
	{gvr.OCICluster, templates.OCICluster},
	{gvr.KubeadmConfigTemplate, templates.OCNEConfigTemplate},
	{gvr.KubeadmControlPlane, templates.OCNEControlPlane},
	{gvr.MachineDeployment, templates.MachineDeployment},
	{gvr.OCIMachineTemplate, templates.OCIMachineTemplate},
	{gvr.OCIMachineTemplate, templates.OCIControlPlaneMachineTemplate},
}

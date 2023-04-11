// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func createObjects(v *variables.Variables) []Object {
	var res []Object

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

type Object struct {
	GVR          schema.GroupVersionResource
	Text         string
	LockedFields map[string]bool
}

var vpo = []Object{
	{GVR: gvr.ConfigMap, Text: templates.VPOConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.VPOResourceSet},
}

var csi = []Object{
	{GVR: gvr.ConfigMap, Text: templates.CSIConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CSIResourceSet},
}

var ccm = []Object{
	{GVR: gvr.ConfigMap, Text: templates.CCMConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CCMResourceSet},
}

var cni = []Object{
	{GVR: gvr.ConfigMap, Text: templates.CalicoConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDFinalConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDInitialConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraaOperatorConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDFinalResourceSet},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraOperatorResourceSet},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoResourceSet},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDInitialResourceSet},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDFinalResourceSet},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraOperatorResourceSet},
}

var capi = []Object{
	{GVR: gvr.Cluster, Text: templates.Cluster},
	{GVR: gvr.ClusterIdentity, Text: templates.ClusterIdentity},
	{GVR: gvr.OCICluster, Text: templates.OCICluster},
	{GVR: gvr.OCNEConfigTemplate, Text: templates.OCNEConfigTemplate},
	{GVR: gvr.OCNEControlPlane, Text: templates.OCNEControlPlane},
	{GVR: gvr.MachineDeployment, Text: templates.MachineDeployment},
	{GVR: gvr.OCIMachineTemplate, Text: templates.OCIMachineTemplate},
	{GVR: gvr.OCIMachineTemplate, Text: templates.OCIControlPlaneMachineTemplate},
}

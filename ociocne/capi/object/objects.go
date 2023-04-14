// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package object

import (
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func CreateObjects(v *variables.Variables) []Object {
	return objectList(v, include{
		workers:      true,
		controlplane: true,
		capi:         true,
	})
}

func UpdateObjects(v *variables.Variables) []Object {
	return objectList(v, include{
		workers:      false,
		controlplane: false,
		capi:         true,
	})
}

func objectList(v *variables.Variables, i include) []Object {
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
	if i.capi {
		res = append(res, capi...)
	}
	if i.controlplane {
		res = append(res, ControlPlane...)
	}
	if i.workers {
		res = append(res, Workers...)
	}
	return res
}

type Object struct {
	GVR          schema.GroupVersionResource
	Text         string
	LockedFields map[string]bool
}

type include struct {
	workers      bool
	controlplane bool
	capi         bool
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
	// Tigera CRDs
	{GVR: gvr.ConfigMap, Text: templates.CalicoTigeraCRDInitialConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDInitialResourceSet},
	{GVR: gvr.ConfigMap, Text: templates.CalicoTigeraCRDFinalConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraCRDFinalResourceSet},
	// Tigera Operator
	{GVR: gvr.ConfigMap, Text: templates.CalicoTigeraaOperatorConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoTigeraOperatorResourceSet},
	// Calico resources
	{GVR: gvr.ConfigMap, Text: templates.CalicoConfigMap},
	{GVR: gvr.ClusterResourceSet, Text: templates.CalicoResourceSet},
}

var ControlPlane = []Object{
	{GVR: gvr.OCNEControlPlane, Text: templates.OCNEControlPlane},
	{GVR: gvr.OCIMachineTemplate, Text: templates.OCIControlPlaneMachineTemplate},
}

var Workers = []Object{
	{GVR: gvr.MachineDeployment, Text: templates.MachineDeployment},
	{GVR: gvr.OCIMachineTemplate, Text: templates.OCIMachineTemplate},
}

var capi = []Object{
	{GVR: gvr.Cluster, Text: templates.Cluster},
	{GVR: gvr.ClusterIdentity, Text: templates.ClusterIdentity},
	{GVR: gvr.OCICluster, Text: templates.OCICluster},
	{GVR: gvr.OCNEConfigTemplate, Text: templates.OCNEConfigTemplate},
}

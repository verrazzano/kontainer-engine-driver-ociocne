// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package object

import (
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

//GVR attempts to find the GVR for an unstructured object

func GVR(u *unstructured.Unstructured) schema.GroupVersionResource {
	gvk := u.GroupVersionKind()

	kind := strings.ToLower(gvk.Kind)
	var resource string
	if kind[len(kind)-1] == 'y' {
		resource = strings.TrimSuffix(kind, "y") + "ies"
	} else {
		resource = kind + "s"
	}
	return schema.GroupVersionResource{
		Group:   gvk.Group,
		Version: gvk.Version,
		// e.g., "Verrazzano" becomes "verrazzanos"
		Resource: resource,
	}
}

func NestedField(o interface{}, fields ...string) (interface{}, error) {
	if len(fields) < 1 {
		return o, nil
	}
	field, remainingFields := fields[0], fields[1:]

	oMap, isMap := o.(map[string]interface{})
	if !isMap {
		return nil, fmt.Errorf("%v is not a map", o)
	}
	oNew, ok := oMap[field]
	if !ok {
		return nil, fmt.Errorf("field %s not found", field)
	}
	return NestedField(oNew, remainingFields...)
}

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
	Text         string
	LockedFields map[string]bool
}

type include struct {
	workers      bool
	controlplane bool
	capi         bool
}

var vpo = []Object{
	{Text: templates.VPOConfigMap},
	{Text: templates.VPOResourceSet},
}

var csi = []Object{
	{Text: templates.CSIConfigMap},
	{Text: templates.CSIResourceSet},
}

var ccm = []Object{
	{Text: templates.CCMConfigMap},
	{Text: templates.CCMResourceSet},
}

var cni = []Object{
	// Tigera CRDs
	{Text: templates.CalicoTigeraCRDInitialConfigMap},
	{Text: templates.CalicoTigeraCRDInitialResourceSet},
	{Text: templates.CalicoTigeraCRDFinalConfigMap},
	{Text: templates.CalicoTigeraCRDFinalResourceSet},
	// Tigera Operator
	{Text: templates.CalicoTigeraaOperatorConfigMap},
	{Text: templates.CalicoTigeraOperatorResourceSet},
	// Calico resources
	{Text: templates.CalicoConfigMap},
	{Text: templates.CalicoResourceSet},
}

var ControlPlane = []Object{
	{Text: templates.OCNEControlPlane},
	{Text: templates.OCIControlPlaneMachineTemplate},
}

var Workers = []Object{
	{Text: templates.MachineDeployment},
	{Text: templates.OCIMachineTemplate},
}

var capi = []Object{
	CAPICluster,
	{Text: templates.ClusterIdentity},
	{Text: templates.OCICluster},
	{Text: templates.OCNEConfigTemplate},
}

var CAPICluster = Object{Text: templates.Cluster}

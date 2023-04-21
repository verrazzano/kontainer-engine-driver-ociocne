// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package gvr

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ControlPlaneXK8sIO   = "controlplane.cluster.x-k8s.io"
	InfrastructureXK8sIO = "infrastructure.cluster.x-k8s.io"
	ClusterXK8sIO        = "cluster.x-k8s.io"
	V1Beta1Version       = "v1beta1"
	V1Alpha1Version      = "v1alpha1"
)

var MachineDeployment = schema.GroupVersionResource{
	Group:    ClusterXK8sIO,
	Version:  V1Beta1Version,
	Resource: "machinedeployments",
}

var Cluster = schema.GroupVersionResource{
	Group:    ClusterXK8sIO,
	Version:  V1Beta1Version,
	Resource: "clusters",
}

var OCICluster = schema.GroupVersionResource{
	Group:    InfrastructureXK8sIO,
	Version:  V1Beta1Version,
	Resource: "ociclusters",
}

var ClusterIdentity = schema.GroupVersionResource{
	Group:    InfrastructureXK8sIO,
	Version:  V1Beta1Version,
	Resource: "ociclusteridentities",
}

var Machine = schema.GroupVersionResource{
	Group:    ClusterXK8sIO,
	Version:  V1Beta1Version,
	Resource: "machines",
}

var OCNEControlPlane = schema.GroupVersionResource{
	Group:    ControlPlaneXK8sIO,
	Version:  V1Alpha1Version,
	Resource: "ocnecontrolplanes",
}

var OCIMachineTemplate = schema.GroupVersionResource{
	Group:    InfrastructureXK8sIO,
	Version:  V1Beta1Version,
	Resource: "ocimachinetemplates",
}

var OCNEConfigTemplate = schema.GroupVersionResource{
	Group:    "bootstrap.cluster.x-k8s.io",
	Version:  V1Alpha1Version,
	Resource: "ocneconfigtemplates",
}

var ClusterResourceSet = schema.GroupVersionResource{
	Group:    "addons.cluster.x-k8s.io",
	Version:  V1Beta1Version,
	Resource: "clusterresourcesets",
}

var ConfigMap = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "configmaps",
}

var Verrazzano = schema.GroupVersionResource{
	Group:    "install.verrazzano.io",
	Version:  V1Beta1Version,
	Resource: "verrazzanos",
}

var VerrazzanoManagedCluster = schema.GroupVersionResource{
	Group:    "clusters.verrazzano.io",
	Version:  V1Alpha1Version,
	Resource: "verrazzanomanagedclusters",
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

const (
	ClusterName           = "name"
	DisplayName           = "display-name"
	KubernetesVersion     = "kubernetes-version"
	NodePublicKeyContents = "node-public-key-contents"

	CompartmentID       = "compartment-id"
	VcnID               = "vcn-id"
	WorkerNodeSubnet    = "worker-node-subnet"
	ControlPlaneSubnet  = "control-plane-subnet"
	LoadBalancerSubnet  = "load-balancer-subnet"
	UsePVNodeEncryption = "use-node-pv-encryption"
	PodCIDR             = "pod-cidr"
	ClusterCIDR         = "cluster-cidr"
	ImageDisplayName    = "image-display-name"
	ImageId             = "image-id"

	RawNodePools = "node-pools"
	ApplyYAMLs   = "apply-yamls"

	ControlPlaneOCPUs     = "control-plane-ocpus"
	NumControlPlaneNodes  = "num-control-plane-nodes"
	ControlPlaneMemoryGbs = "control-plane-memory-gbs"
	ControlPlaneShape     = "control-plane-shape"
	ControlPlaneVolumeGbs = "control-plane-volume-gbs"

	ControlPlaneRegistry = "control-plane-registry"
	CalicoRegistry       = "calico-image-registry"
	CalicoImagePath      = "calico-image-path"
	// TigeraTag used to determine version of tigera operator
	TigeraTag     = "tigera-image-tag"
	CCMImage      = "ccm-image"
	OCICSIImage   = "oci-csi-image"
	CSIRegistry   = "csi-registry"
	InstallCalico = "install-calico"
	InstallCCM    = "install-ccm"
	InstallCSI    = "install-csi"

	InstallVerrazzano  = "install-verrazzano"
	VerrazzanoResource = "verrazzano-resource"
	VerrazzanoImage    = "verrazzano-image"

	ProxyEndpoint = "proxy-endpoint"

	PreOCNECommands  = "pre-ocne-commands"
	PostOCNECommands = "post-ocne-commands"
	SkipOCNEInstall  = "skip-ocne-install"

	CloudCredentialId = "cloud-credential-id"
	Region            = "region"
)

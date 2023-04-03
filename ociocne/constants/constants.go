// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

const (
	ClusterName           = "name"
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

	NumWorkerNodes = "num-worker-nodes"
	NodeShape      = "node-shape"
	NodeOCPUs      = "node-ocpus"
	NodeMemoryGbs  = "node-memory-gbs"

	ControlPlaneOCPUs     = "control-plane-ocpus"
	NumControlPlaneNodes  = "num-control-plane-nodes"
	ControlPlaneMemoryGbs = "control-plane-memory-gbs"
	ControlPlaneShape     = "control-plane-shape"

	ETCDImageTag         = "etcd-image-tag"
	CoreDNSImageTag      = "core-dns-image-tag"
	ControlPlaneRegistry = "control-plane-registry"
	CalicoRegistry       = "calico-image-registry"
	CalicoTag            = "calico-image-tag"
	CCMImage             = "ccm-image"

	ProxyEndpoint = "proxy-endpoint"

	PreOCNECommands  = "pre-ocne-commands"
	PostOCNECommands = "post-ocne-commands"

	CloudCredentialId  = "cloud-credential-id"
	SecretName         = "secret-name"
	Fingerprint        = "fingerprint"
	PrivateKeyContents = "private-key-contents"
	TenancyId          = "tenancy-id"
	UserId             = "user-id"
	Region             = "region"
)

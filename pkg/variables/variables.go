// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package variables

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/rancher/kontainer-engine/drivers/options"
	"github.com/rancher/kontainer-engine/store"
	"github.com/rancher/kontainer-engine/types"
	driverconst "github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/constants"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/k8s"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/oci"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	DefaultOCICPUs                 = 2
	DefaultMemoryGbs               = 16
	DefaultVolumeGbs               = 100
	DefaultNodePVTransitEncryption = true
	DefaultVMShape                 = "VM.Standard.E4.Flex"
	ProviderId                     = `oci://{{ ds["id"] }}`

	DefaultCNEPath            = "olcne"
	DefaultVerrazzanoResource = `apiVersion: install.verrazzano.io/v1beta1
kind: Verrazzano
metadata:
  name: managed
  namespace: default
spec:
  profile: managed-cluster`
)

const (
	kubeconfigName   = "%s-kubeconfig"
	CAPIOCINamespace = "verrazzano-capi"

	loadBalancerSubnetRole         = "service-lb"
	controlPlaneSubnetRole         = "control-plane"
	controlPlaneEndpointSubnetRole = "control-plane-endpoint"
	workerSubnetRole               = "worker"
)

type Subnet struct {
	Id   string
	Role string
	Name string
	CIDR string
	Type string
}

type NodePool struct {
	Name       string `json:"name"`
	Replicas   int64  `json:"replicas"`
	Memory     int64  `json:"memory"`
	Ocpus      int64  `json:"ocpus"`
	VolumeSize int64  `json:"volumeSize"`
	Shape      string `json:"shape"`
}

var OCIClientGetter = func(v *Variables) (oci.Client, error) {
	return oci.NewClient(v.GetConfigurationProvider())
}

type (
	//Variables are parameters for cluster lifecycle operations
	Variables struct {
		Name             string
		DisplayName      string
		Namespace        string
		Hash             string
		ControlPlaneHash string
		NodePoolHash     string

		QuickCreateVCN     bool
		VCNID              string
		WorkerNodeSubnet   string
		ControlPlaneSubnet string
		LoadBalancerSubnet string
		// Parsed subnets
		Subnets       []Subnet `json:"subnets,omitempty"`
		PodCIDR       string
		ClusterCIDR   string
		ProxyEndpoint string

		// Cluster topology and configuration
		KubernetesVersion       string
		OCNEVersion             string
		SSHPublicKey            string
		ControlPlaneShape       string
		ControlPlaneReplicas    int64
		ControlPlaneOCPUs       int64
		ControlPlaneMemoryGbs   int64
		ControlPlaneVolumeGbs   int64
		NodePVTransitEncryption bool
		RawNodePools            []string
		ApplyYAMLS              []string
		// Parsed node pools
		NodePools []NodePool

		// ImageID is looked up by display name
		ImageDisplayName string
		ImageID          string
		ActualImage      string

		PreOCNECommands  []string
		PostOCNECommands []string
		SkipOCNEInstall  bool

		// Addons, images, and registries
		InstallVerrazzano bool
		// Set to true during Updates
		UninstallVerrazzano bool
		VerrazzanoResource  string
		VerrazzanoVersion   string
		VerrazzanoTag       string
		InstallCalico       bool
		InstallCCM          bool
		CNEPath             string
		TigeraTag           string
		ETCDImageTag        string
		CoreDNSImageTag     string

		// Private registry
		PrivateRegistry string

		// OCI Credentials
		CAPIOCINamespace     string
		CAPICredentialName   string
		CloudCredentialId    string
		CompartmentID        string
		Fingerprint          string
		PrivateKey           string
		PrivateKeyPassphrase string
		Region               string
		Tenancy              string
		User                 string

		// Supplied for templating
		ProviderId string
	}
)

// NewFromOptions creates a new Variables given *types.DriverOptions
func NewFromOptions(ctx context.Context, driverOptions *types.DriverOptions) (*Variables, error) {
	v := &Variables{
		Name:              options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ClusterName).(string),
		DisplayName:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.DisplayName, "displayName").(string),
		KubernetesVersion: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.KubernetesVersion, "kubernetesVersion").(string),
		OCNEVersion:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.OCNEVersion, "ocneVersion").(string),

		// User and authentication
		SSHPublicKey:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodePublicKeyContents, "nodePublicKeyContents").(string),
		CloudCredentialId: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CloudCredentialId, "cloudCredentialId").(string),
		Region:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.Region, "region").(string),
		CompartmentID:     options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CompartmentID, "compartmentId").(string),

		// Networking
		QuickCreateVCN:     options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.QuickCreateVCN, "quickCreateVcn").(bool),
		VCNID:              options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VcnID, "vcnId").(string),
		WorkerNodeSubnet:   options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.WorkerNodeSubnet, "workerNodeSubnet").(string),
		LoadBalancerSubnet: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.LoadBalancerSubnet, "loadBalancerSubnet").(string),
		ControlPlaneSubnet: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneSubnet, "controlPlaneSubnet").(string),
		PodCIDR:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.PodCIDR, "podCidr").(string),
		ClusterCIDR:        options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ClusterCIDR, "clusterCidr").(string),

		// VM settings
		ImageDisplayName:        options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ImageDisplayName, "imageDisplayName").(string),
		NodePVTransitEncryption: options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.UsePVNodeEncryption, "useNodePVEncryption").(bool),
		ControlPlaneReplicas:    options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumControlPlaneNodes, "numControlPlaneNodes").(int64),
		ControlPlaneShape:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneShape, "controlPlaneShape").(string),
		ControlPlaneOCPUs:       options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.ControlPlaneOCPUs, "controlPlaneOcpus").(int64),
		ControlPlaneMemoryGbs:   options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.ControlPlaneMemoryGbs, "controlPlaneMemoryGbs").(int64),
		ControlPlaneVolumeGbs:   options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.ControlPlaneVolumeGbs, "controlPlaneVolumeGbs").(int64),
		RawNodePools:            options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.RawNodePools, "nodePools").(*types.StringSlice).Value,
		ApplyYAMLS:              options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.ApplyYAMLs, "applyYamls").(*types.StringSlice).Value,

		// Image settings
		CNEPath:         options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CNEPath, "cnePath").(string),
		TigeraTag:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.TigeraTag, "tigeraImageTag").(string),
		ETCDImageTag:    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ETCDTag, "etcdImageTag").(string),
		CoreDNSImageTag: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CoreDNSTag, "corednsImageTag").(string),
		InstallCalico:   options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCalico, "installCalico").(bool),
		InstallCCM:      options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCCM, "installCcm").(bool),

		// Private Registry
		PrivateRegistry: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.PrivateRegistry, "privateRegistry").(string),

		// Verrazzano settings
		VerrazzanoTag:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VerrazzanoTag, "verrazzanoTag").(string),
		VerrazzanoVersion:  options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VerrazzanoVersion, "verrazzanoVersion").(string),
		VerrazzanoResource: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VerrazzanoResource, "verrazzanoResource").(string),
		InstallVerrazzano:  options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallVerrazzano, "installVerrazzano").(bool),

		// Other
		ProxyEndpoint:    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ProxyEndpoint, "proxyEndpoint").(string),
		ImageID:          options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ImageId, "imageId").(string),
		SkipOCNEInstall:  options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.SkipOCNEInstall, "skipOcneInstall").(bool),
		PreOCNECommands:  options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PreOCNECommands, "preOcneCommands").(*types.StringSlice).Value,
		PostOCNECommands: options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PostOCNECommands, "postOcneCommands").(*types.StringSlice).Value,
		ProviderId:       ProviderId,
		CAPIOCINamespace: CAPIOCINamespace,
	}
	v.Namespace = v.Name

	if err := v.SetDynamicValues(ctx); err != nil {
		return v, err
	}
	return v, nil
}

// SetUpdateValues are the values potentially changed during an update operation
func (v *Variables) SetUpdateValues(ctx context.Context, vNew *Variables) error {
	// Uninstall Verrazzano if the new state has no Verrazzano
	v.UninstallVerrazzano = false
	if v.InstallVerrazzano && !vNew.InstallVerrazzano {
		v.UninstallVerrazzano = true
	}
	v.KubernetesVersion = vNew.KubernetesVersion
	v.ControlPlaneReplicas = vNew.ControlPlaneReplicas
	v.ImageDisplayName = vNew.ImageDisplayName
	v.ControlPlaneOCPUs = vNew.ControlPlaneOCPUs
	v.ControlPlaneMemoryGbs = vNew.ControlPlaneMemoryGbs
	v.ControlPlaneVolumeGbs = vNew.ControlPlaneVolumeGbs
	v.RawNodePools = vNew.RawNodePools
	v.SSHPublicKey = vNew.SSHPublicKey
	v.DisplayName = vNew.DisplayName
	v.SkipOCNEInstall = vNew.SkipOCNEInstall
	v.ImageID = vNew.ImageID
	v.ApplyYAMLS = vNew.ApplyYAMLS
	v.TigeraTag = vNew.TigeraTag
	v.ETCDImageTag = vNew.ETCDImageTag
	v.CoreDNSImageTag = vNew.CoreDNSImageTag
	v.PrivateRegistry = vNew.PrivateRegistry
	v.InstallVerrazzano = vNew.InstallVerrazzano
	v.VerrazzanoTag = vNew.VerrazzanoTag
	v.VerrazzanoVersion = vNew.VerrazzanoVersion
	v.VerrazzanoResource = vNew.VerrazzanoResource
	return v.SetDynamicValues(ctx)
}

// SetDynamicValues sets dynamic values
func (v *Variables) SetDynamicValues(ctx context.Context) error {
	// deserialize node pools
	nodePools, err := v.ParseNodePools()
	if err != nil {
		return err
	}
	v.NodePools = nodePools

	// setup OCI client for dynamic values
	ki, err := k8s.InjectedInterface()
	if err != nil {
		return err
	}
	if err := SetupOCIAuth(ctx, ki, v); err != nil {
		return err
	}
	ociClient, err := OCIClientGetter(v)

	if err != nil {
		return err
	}
	// get image OCID from OCI
	if err := v.setImageId(ctx, ociClient); err != nil {
		return err
	}
	// get subnet metadata from OCI
	if err := v.setSubnets(ctx, ociClient); err != nil {
		return err
	}

	// set hashes for controlplane updates
	v.SetHashes()
	return nil
}

// GetConfigurationProvider creates a new configuration provider from Variables
func (v *Variables) GetConfigurationProvider() common.ConfigurationProvider {
	var passphrase *string
	if len(v.PrivateKeyPassphrase) > 0 {
		passphrase = &v.PrivateKeyPassphrase
	}
	privateKey := strings.TrimSpace(v.PrivateKey)
	return common.NewRawConfigurationProvider(v.Tenancy, v.User, v.Region, v.Fingerprint, privateKey, passphrase)
}

// GetCAPIClusterKubeConfig fetches the cluster's kubeconfig
func (v *Variables) GetCAPIClusterKubeConfig(ctx context.Context) (*store.KubeConfig, error) {
	client, err := k8s.InjectedInterface()
	if err != nil {
		return nil, err
	}
	kubeconfigSecretName := fmt.Sprintf(kubeconfigName, v.Name)
	secret, err := client.CoreV1().Secrets(v.Namespace).Get(ctx, kubeconfigSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	kubeconfig := &store.KubeConfig{}
	err = yaml.Unmarshal(secret.Data["value"], kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubeconfig, nil
}

// NodeCount is the sum of worker and control plane nodes
func (v *Variables) NodeCount() (*types.NodeCount, error) {
	nps, err := v.ParseNodePools()
	if err != nil {
		return nil, err
	}
	v.NodePools = nps
	return &types.NodeCount{
		Count: v.ControlPlaneReplicas + v.workerNodeCount(),
	}, nil
}

func (v *Variables) IsSingleNodeCluster() bool {
	return v.workerNodeCount() == 0
}

func (v *Variables) workerNodeCount() int64 {
	var count int64 = 0
	for _, np := range v.NodePools {
		count = count + np.Replicas
	}
	return count
}

// Version is the cluster Kubernetes version
func (v *Variables) Version() *types.KubernetesVersion {
	return &types.KubernetesVersion{
		Version: v.KubernetesVersion,
	}
}

func (v *Variables) ParseNodePools() ([]NodePool, error) {
	var nodePools []NodePool

	for _, rawNodePool := range v.RawNodePools {
		nodePool := NodePool{}
		if err := json.Unmarshal([]byte(rawNodePool), &nodePool); err != nil {
			return nil, err
		}
		nodePools = append(nodePools, nodePool)
	}

	return nodePools, nil
}

func (v *Variables) setImageId(ctx context.Context, client oci.Client) error {
	// if user is bringing their own image, skip the dynamic image lookup
	if v.SkipOCNEInstall {
		v.ActualImage = v.ImageID
	} else {
		imageId, err := client.GetImageIdByName(ctx, v.ImageDisplayName, v.CompartmentID)
		if err != nil {
			return err
		}
		v.ActualImage = imageId
	}

	return nil
}

func (v *Variables) setSubnets(ctx context.Context, client oci.Client) error {
	var subnets []Subnet
	subnetCache := map[string]*Subnet{}

	addSubnetForRole := func(subnetId, role string) error {
		var err error
		subnet := subnetCache[subnetId]
		if subnet == nil && subnetId != "" {
			subnet, err = getSubnetById(ctx, client, subnetId, role)
			if err != nil {
				return err
			}
		}
		if subnet != nil {
			subnets = append(subnets, *subnet)
		}
		return nil
	}

	if err := addSubnetForRole(v.LoadBalancerSubnet, loadBalancerSubnetRole); err != nil {
		return err
	}
	if err := addSubnetForRole(v.ControlPlaneSubnet, controlPlaneSubnetRole); err != nil {
		return err
	}
	if err := addSubnetForRole(v.ControlPlaneSubnet, controlPlaneEndpointSubnetRole); err != nil {
		return err
	}
	if err := addSubnetForRole(v.WorkerNodeSubnet, workerSubnetRole); err != nil {
		return err
	}
	v.Subnets = subnets
	return nil
}

func getSubnetById(ctx context.Context, client oci.Client, subnetId, role string) (*Subnet, error) {
	sn, err := client.GetSubnetById(ctx, subnetId)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet %s", subnetId)
	}

	return &Subnet{
		Id:   subnetId,
		CIDR: *sn.CidrBlock,
		Type: oci.SubnetAccess(*sn),
		Name: role,
		Role: role,
	}, nil
}

// SetupOCIAuth dynamically loads OCI authentication
func SetupOCIAuth(ctx context.Context, client kubernetes.Interface, v *Variables) error {
	ccName, ccNamespace := v.cloudCredentialNameAndNamespace()
	cc, err := client.CoreV1().Secrets(ccNamespace).Get(ctx, ccName, metav1.GetOptions{})
	// Failed to retrieve cloud credentials
	if err != nil {
		return err
	}

	v.CAPICredentialName = ccName
	v.User = string(cc.Data["ocicredentialConfig-userId"])
	v.Fingerprint = string(cc.Data["ocicredentialConfig-fingerprint"])
	v.Tenancy = string(cc.Data["ocicredentialConfig-tenancyId"])
	v.PrivateKeyPassphrase = string(cc.Data["ocicredentialConfig-passphrase"])
	v.PrivateKey = string(cc.Data["ocicredentialConfig-privateKeyContents"])
	return nil
}

func (v *Variables) SetQuickCreateVCNInfo(ctx context.Context, di dynamic.Interface) error {
	// Only set Quick Create VCN Info if using Quick Create VCN, and the VCN info is unset.
	if v.QuickCreateVCN && v.isNetworkingUnset() {
		ociCluster, err := di.Resource(gvr.OCICluster).Namespace(v.Name).Get(ctx, v.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Get VCN id and VCN subnets from the OCI Cluster resource
		vcnID, found, err := unstructured.NestedString(ociCluster.Object, "spec", "networkSpec", "vcn", "id")
		if err != nil || !found {
			return errors.New("waiting for VCN to be created")
		}
		v.VCNID = vcnID
		subnets, found, err := unstructured.NestedSlice(ociCluster.Object, "spec", "networkSpec", "vcn", "subnets")
		if err != nil || !found {
			return errors.New("waiting for subnets to be created")
		}

		// For each subnet in the VCN subnet list, identify its role and populate the subnet id in the cluster state
		for _, subnet := range subnets {
			subnetObject, ok := subnet.(map[string]interface{})
			if !ok {
				return errors.New("subnet is creating")
			}

			// Get nested subnet Role and id from the subnet object
			subnetRole, found, err := unstructured.NestedString(subnetObject, "role")
			if err != nil || !found {
				return errors.New("waiting for subnet role to be populated")
			}
			subnetId, found, err := unstructured.NestedString(subnetObject, "id")
			if err != nil {
				return errors.New("waiting for subnet id to be populated")
			}

			// Populate the subnet id depending on the subnet role
			switch subnetRole {
			case "control-plane":
				v.ControlPlaneSubnet = subnetId
			case "service-lb":
				v.LoadBalancerSubnet = subnetId
			case "worker":
				v.WorkerNodeSubnet = subnetId
			default: // we are not interested in any other subnets
				continue
			}
		}
	}

	return nil
}

func (v *Variables) isNetworkingUnset() bool {
	return len(v.VCNID) < 1 || len(v.ControlPlaneSubnet) < 1 || len(v.LoadBalancerSubnet) < 1 || len(v.WorkerNodeSubnet) < 1
}

func (v *Variables) cloudCredentialNameAndNamespace() (string, string) {
	split := strings.Split(v.CloudCredentialId, ":")

	if len(split) == 1 {
		return "cattle-global-data", split[0]
	}
	return split[1], split[0]
}

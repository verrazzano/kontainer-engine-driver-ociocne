// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package variables

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/rancher/kontainer-engine/drivers/options"
	"github.com/rancher/kontainer-engine/store"
	"github.com/rancher/kontainer-engine/types"
	driverconst "github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/constants"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/k8s"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/oci"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/version"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
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
	DefaultRegistry           = "container-registry.oracle.com"
	DefaultTigeraTag          = "v1.29.0"
	DefaultCCMImage           = "ghcr.io/oracle/cloud-provider-oci:v1.24.0"
	DefaultOCICSIImage        = "ghcr.io/oracle/cloud-provider-oci:v1.24.0"
	DefaultCSIRegistry        = "k8s.gcr.io/sig-storage"
	DefaultVerrazzanoImage    = "ghcr.io/verrazzano/verrazzano-platform-operator:v1.5.2-20230315235330-0326ee67"
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
	CAPIOCINamespace = "cluster-api-provider-oci-system"

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

type (
	//Variables are parameters for cluster lifecycle operations
	Variables struct {
		Name        string
		DisplayName string
		Namespace   string
		Hash        string

		ImageID                 string
		ImageDisplayName        string
		VCNID                   string
		WorkerNodeSubnet        string
		ControlPlaneSubnet      string
		LoadBalancerSubnet      string
		SSHPublicKey            string
		ControlPlaneReplicas    int64
		NodeReplicas            int64
		NodePVTransitEncryption bool
		NodeShape               string
		ControlPlaneShape       string
		KubernetesVersion       string
		NodeOCPUs               int64
		ControlPlaneOCPUs       int64
		NodeMemoryGbs           int64
		NodeVolumeGbs           int64
		ControlPlaneMemoryGbs   int64
		ControlPlaneVolumeGbs   int64
		PodCIDR                 string
		ClusterCIDR             string
		ProxyEndpoint           string
		PreOCNECommands         []string
		PostOCNECommands        []string
		ControlPlaneRegistry    string
		CalicoRegistry          string
		CalicoImagePath         string
		TigeraTag               string
		ETCDImageTag            string
		CoreDNSImageTag         string
		CCMImage                string
		CSIRegistry             string
		OCICSIImage             string
		ProviderId              string

		InstallVerrazzano  bool
		VerrazzanoResource string
		VerrazzanoImage    string

		InstallCalico bool
		InstallCCM    bool
		InstallCSI    bool

		CAPIOCINamespace   string
		CAPICredentialName string

		CloudCredentialId    string
		CompartmentID        string
		Fingerprint          string
		PrivateKey           string
		PrivateKeyPassphrase string
		Region               string
		Tenancy              string
		User                 string

		// Parsed subnets
		Subnets []Subnet `json:"subnets,omitempty"`
	}
)

// NewFromOptions creates a new Variables given *types.DriverOptions
func NewFromOptions(ctx context.Context, driverOptions *types.DriverOptions) (*Variables, error) {
	v := &Variables{
		Name:              options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ClusterName).(string),
		DisplayName:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.DisplayName, "displayName").(string),
		KubernetesVersion: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.KubernetesVersion, "kubernetesVersion").(string),

		// User and authentication
		SSHPublicKey:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodePublicKeyContents, "nodePublicKeyContents").(string),
		CloudCredentialId: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CloudCredentialId, "cloudCredentialId").(string),
		Region:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.Region, "region").(string),
		CompartmentID:     options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CompartmentID, "compartmentId").(string),

		// Networking
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
		NodeReplicas:            options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumWorkerNodes, "numWorkerNodes").(int64),
		NodeShape:               options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodeShape, "nodeShape").(string),
		NodeOCPUs:               options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NodeOCPUs, "nodeOcpus").(int64),
		NodeMemoryGbs:           options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NodeMemoryGbs, "nodeMemoryGbs").(int64),
		NodeVolumeGbs:           options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NodeVolumeGbs, "nodeVolumeGbs").(int64),

		// Image settings
		ControlPlaneRegistry: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneRegistry, "controlPlaneRegistry").(string),
		CalicoRegistry:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CalicoRegistry, "calicoImageRegistry").(string),
		CalicoImagePath:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CalicoImagePath, "calicoImagePath").(string),
		TigeraTag:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.TigeraTag, "tigeraImageTag").(string),
		CCMImage:             options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CCMImage, "ccmImage").(string),
		OCICSIImage:          options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.OCICSIImage, "ociCsiImage").(string),
		CSIRegistry:          options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CSIRegistry, "csiRegistry").(string),
		InstallCalico:        options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCalico, "installCalico").(bool),
		InstallCCM:           options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCCM, "installCcm").(bool),
		InstallCSI:           options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCSI, "installCsi").(bool),

		// Verrazzano settings
		VerrazzanoImage:    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VerrazzanoImage, "verrazzanoImage").(string),
		VerrazzanoResource: options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VerrazzanoResource, "verrazzanoResource").(string),
		InstallVerrazzano:  options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.InstallCalico, "installVerrazzano").(bool),

		// Other
		ProxyEndpoint:    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ProxyEndpoint, "proxyEndpoint").(string),
		PreOCNECommands:  options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PreOCNECommands, "preOcneCommands").(*types.StringSlice).Value,
		PostOCNECommands: options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PostOCNECommands, "postOcneCommands").(*types.StringSlice).Value,
		ProviderId:       ProviderId,
		CAPIOCINamespace: CAPIOCINamespace,
	}
	v.Namespace = v.Name

	if err := v.SetVersionMapping(); err != nil {
		return v, err
	}

	if err := v.SetDynamicValues(ctx); err != nil {
		return v, err
	}
	return v, nil
}

func (v *Variables) SetUpdateValues(ctx context.Context, vNew *Variables) error {
	v.KubernetesVersion = vNew.KubernetesVersion
	if err := v.SetVersionMapping(); err != nil {
		return err
	}
	v.NodeReplicas = vNew.NodeReplicas
	v.ControlPlaneReplicas = vNew.ControlPlaneReplicas
	v.ImageDisplayName = vNew.ImageDisplayName
	v.NodeOCPUs = vNew.NodeOCPUs
	v.ControlPlaneOCPUs = vNew.ControlPlaneOCPUs
	v.NodeMemoryGbs = vNew.NodeMemoryGbs
	v.ControlPlaneMemoryGbs = vNew.ControlPlaneMemoryGbs
	v.NodeVolumeGbs = vNew.NodeVolumeGbs
	v.ControlPlaneVolumeGbs = vNew.ControlPlaneVolumeGbs
	v.SSHPublicKey = vNew.SSHPublicKey
	v.DisplayName = vNew.DisplayName
	return v.SetDynamicValues(ctx)
}

// SetDynamicValues sets dynamic values from OCI in the Variables
func (v *Variables) SetDynamicValues(ctx context.Context) error {
	hash, err := v.HashSum()
	if err != nil {
		return err
	}
	v.Hash = hash
	client, err := k8s.InjectedInterface()
	if err != nil {
		return err
	}
	if err := SetupOCIAuth(ctx, client, v); err != nil {
		return err
	}
	ociClient, err := oci.NewClient(v.GetConfigurationProvider())
	if err != nil {
		return err
	}
	if err := v.setImageId(ctx, ociClient); err != nil {
		return err
	}
	return v.setSubnets(ctx, ociClient)
}

// GetConfigurationProvider creates a new configuration provider from Variables
func (v *Variables) GetConfigurationProvider() common.ConfigurationProvider {
	var passphrase *string
	if len(v.PrivateKeyPassphrase) > 0 {
		passphrase = &v.PrivateKeyPassphrase
	}
	privateKey := strings.TrimSpace(strings.ReplaceAll(v.PrivateKey, "\\n", "\n"))
	return common.NewRawConfigurationProvider(v.Tenancy, v.User, v.Region, v.Fingerprint, privateKey, passphrase)
}

// GetCAPIClusterKubeConfig fetches the cluster's kubeconfig
func (v *Variables) GetCAPIClusterKubeConfig(ctx context.Context, state *Variables) (*store.KubeConfig, error) {
	client, err := k8s.InjectedInterface()
	if err != nil {
		return nil, err
	}
	kubeconfigSecretName := fmt.Sprintf(kubeconfigName, v.Name)
	secret, err := client.CoreV1().Secrets(state.Namespace).Get(ctx, kubeconfigSecretName, metav1.GetOptions{})
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
func (v *Variables) NodeCount() *types.NodeCount {
	return &types.NodeCount{
		Count: v.NodeReplicas + v.ControlPlaneReplicas,
	}
}

// Version is the cluster Kubernetes version
func (v *Variables) Version() *types.KubernetesVersion {
	return &types.KubernetesVersion{
		Version: v.KubernetesVersion,
	}
}

func (v *Variables) SetVersionMapping() error {
	props, ok := version.Mapping[v.KubernetesVersion]
	if !ok {
		return fmt.Errorf("unknown kubernetes version %s", v.KubernetesVersion)
	}
	v.ETCDImageTag = props.ETCDImageTag
	v.CoreDNSImageTag = props.CoreDNSImageTag
	v.TigeraTag = props.TigeraTag
	return nil
}

// Validate asserts values are acceptable for cluster lifecycle operations
func (v *Variables) Validate() error {
	return nil
}

func (v *Variables) HashSum() (string, error) {
	vCopy := v
	vCopy.Hash = ""
	bytes, err := json.Marshal(vCopy)
	if err != nil {
		return "", err
	}
	sha := sha256.New()
	sha.Write(bytes)

	encoded := base32.StdEncoding.EncodeToString(sha.Sum(nil))
	return strings.ToLower(encoded[0:5]), nil
}

func (v *Variables) setImageId(ctx context.Context, client oci.Client) error {
	imageId, err := client.GetImageIdByName(ctx, v.ImageDisplayName, v.CompartmentID)
	if err != nil {
		return err
	}
	v.ImageID = imageId
	return nil
}

func (v *Variables) setSubnets(ctx context.Context, client oci.Client) error {
	var subnets []Subnet
	subnetCache := map[string]*Subnet{}

	addSubnetForRole := func(subnetId, role string) error {
		var err error
		subnet := subnetCache[subnetId]
		if subnet == nil {
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
		return nil, err
	}
	if sn == nil {
		return nil, nil
	}

	ip, _, err := net.ParseCIDR(*sn.CidrBlock)
	if err != nil {
		return nil, err
	}
	var addressType = "public"
	if ip.IsPrivate() {
		addressType = "private"
	}
	return &Subnet{
		Id:   subnetId,
		CIDR: *sn.CidrBlock,
		Type: addressType,
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
	// TODO: Support private key passphrase in cloud credentials
	v.PrivateKeyPassphrase = ""
	v.PrivateKey = string(cc.Data["ocicredentialConfig-privateKeyContents"])
	v.PrivateKey = strings.Replace(v.PrivateKey, "\n", "\\n", -1)
	return nil
}

func (v *Variables) cloudCredentialNameAndNamespace() (string, string) {
	split := strings.Split(v.CloudCredentialId, ":")

	if len(split) == 1 {
		return "cattle-global-data", split[0]
	}
	return split[1], split[0]
}

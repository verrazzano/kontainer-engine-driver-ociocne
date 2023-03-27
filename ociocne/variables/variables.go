// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package variables

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/rancher/kontainer-engine/drivers/options"
	"github.com/rancher/kontainer-engine/store"
	"github.com/rancher/kontainer-engine/types"
	driverconst "github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/constants"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/k8s"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/oci"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"strings"
)

const (
	DefaultOCICPUs                 = 2
	DefaultMemoryGB                = 16
	DefaultNodePVTransitEncryption = true
	DefaultVMShape                 = "VM.Standard.E4.Flex"
	ProviderId                     = `oci://{{ ds["id"] }}`

	DefaultRegistryCNE     = "container-registry.oracle.com/olcne"
	DefaultETCDImageTag    = "3.5.3"
	DefaultCoreDNSImageTag = "1.8.6"
	DefaultCalicoTag       = "v3.25.0"
	DefaultCCMImage        = "ghcr.io/oracle/cloud-provider-oci:v1.24.0"
)

const (
	kubeconfigName = "%s-kubeconfig"
	namespace      = "cluster-api-provider-oci-system"
	name           = "capoci-auth-config"

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
		Name                    string
		Namespace               string
		CompartmentID           string
		ImageID                 string
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
		ControlPlaneMemoryGbs   int64
		PodCIDR                 string
		ClusterCIDR             string
		ProxyEndpoint           string
		PreOCNECommands         []string
		PostOCNECommands        []string
		ControlPlaneRegistry    string
		CalicoRegistry          string
		CalicoTag               string
		ETCDImageTag            string
		CoreDNSImageTag         string
		CCMImage                string
		CSIRegistry             string

		ProviderId string
		// OCI Auth is loaded from the CAPI Provider
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
		Name:                    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ClusterName).(string),
		CompartmentID:           options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CompartmentID, "compartmentId").(string),
		ImageID:                 options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodeImage, "nodeImage").(string),
		VCNID:                   options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.VcnID, "vcnId").(string),
		WorkerNodeSubnet:        options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.WorkerNodeSubnet, "workerNodeSubnet").(string),
		LoadBalancerSubnet:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.LoadBalancerSubnet, "loadBalancerSubnet").(string),
		ControlPlaneSubnet:      options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneSubnet, "controlPlaneSubnet").(string),
		SSHPublicKey:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodePublicKeyContents, "nodePublicKeyContents").(string),
		ControlPlaneReplicas:    options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumControlPlaneNodes, "numControlPlaneNodes").(int64),
		NodeReplicas:            options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumWorkerNodes, "numWorkerNodes").(int64),
		NodePVTransitEncryption: options.GetValueFromDriverOptions(driverOptions, types.BoolType, driverconst.UsePVNodeEncryption, "useNodePVEncryption").(bool),
		NodeShape:               options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.NodeShape, "nodeShape").(string),
		ControlPlaneShape:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneShape, "controlPlaneShape").(string),
		KubernetesVersion:       options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.KubernetesVersion, "kubernetesVersion").(string),
		NodeOCPUs:               options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NodeOCPUs, "nodeOcpus").(int64),
		ControlPlaneOCPUs:       options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.ControlPlaneOCPUs, "controlPlaneOcpus").(int64),
		NodeMemoryGbs:           options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NodeMemoryGbs, "nodeMemoryGbs").(int64),
		ControlPlaneMemoryGbs:   options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.ControlPlaneMemoryGbs, "controlPlaneMemoryGbs").(int64),
		PodCIDR:                 options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.PodCIDR, "podCidr").(string),
		ClusterCIDR:             options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ClusterCIDR, "clusterCidr").(string),
		ControlPlaneRegistry:    options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ControlPlaneRegistry, "controlPlaneRegistry").(string),
		CalicoRegistry:          options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CalicoRegistry, "calicoImageRegistry").(string),
		CalicoTag:               options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CalicoTag, "calicoImageTag").(string),
		CCMImage:                options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CCMImage, "ccmImage").(string),
		ETCDImageTag:            options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ETCDImageTag, "etcdImageTag").(string),
		CoreDNSImageTag:         options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.CoreDNSImageTag, "coreDnsImageTag").(string),
		ProxyEndpoint:           options.GetValueFromDriverOptions(driverOptions, types.StringType, driverconst.ProxyEndpoint, "proxyEndpoint").(string),
		PreOCNECommands:         options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PreOCNECommands, "preOcneCommands").(*types.StringSlice).Value,
		PostOCNECommands:        options.GetValueFromDriverOptions(driverOptions, types.StringSliceType, driverconst.PostOCNECommands, "postOcneCommands").(*types.StringSlice).Value,
		ProviderId:              ProviderId,
	}
	v.Namespace = v.Name
	if err := v.SetDynamicValues(ctx); err != nil {
		return v, err
	}
	return v, nil
}

// SetUpdateOptions sets update options in the Variables
func (v *Variables) SetUpdateOptions(driverOptions *types.DriverOptions) {
	v.ControlPlaneReplicas = options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumControlPlaneNodes, "numControlPlaneNodes").(int64)
	v.NodeReplicas = options.GetValueFromDriverOptions(driverOptions, types.IntType, driverconst.NumWorkerNodes, "numWorkerNodes").(int64)
}

// SetDynamicValues sets dynamic values from OCI in the Variables
func (v *Variables) SetDynamicValues(ctx context.Context) error {
	client, err := k8s.NewInterface()
	if err != nil {
		return err
	}
	if err := LoadOCIAuth(ctx, client, v); err != nil {
		return err
	}
	ociClient, err := oci.NewClient(v.GetConfigurationProvider())
	if err != nil {
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
	client, err := k8s.NewInterface()
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

// Validate asserts values are acceptable for cluster lifecycle operations
func (v *Variables) Validate() error {
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

// LoadOCIAuth dynamically loads OCI authentication
func LoadOCIAuth(ctx context.Context, client kubernetes.Interface, v *Variables) error {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	v.Fingerprint = string(secret.Data["fingerprint"])
	v.PrivateKey = strings.Replace(string(secret.Data["key"]), "\n", "\\n", -1)
	v.PrivateKeyPassphrase = string(secret.Data["passphrase"])
	v.Region = string(secret.Data["region"])
	v.Tenancy = string(secret.Data["tenancy"])
	v.User = string(secret.Data["user"])
	return nil
}

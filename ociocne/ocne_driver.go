// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ociocne

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rancher/kontainer-engine/types"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/capi"
	driverconst "github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/constants"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/k8s"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	metadataKey = "state"
)

type OCIOCNEDriver struct {
	Logger             *zap.SugaredLogger
	driverCapabilities types.Capabilities
}

func NewDriver() types.Driver {
	driver := &OCIOCNEDriver{
		driverCapabilities: types.Capabilities{
			Capabilities: make(map[int64]bool),
		},
	}

	driver.driverCapabilities.AddCapability(types.GetVersionCapability)
	driver.driverCapabilities.AddCapability(types.SetVersionCapability)
	driver.driverCapabilities.AddCapability(types.GetClusterSizeCapability)
	driver.driverCapabilities.AddCapability(types.SetClusterSizeCapability)

	return driver
}

func (d *OCIOCNEDriver) Remove(ctx context.Context, info *types.ClusterInfo) error {
	d.Logger.Infof("capi.driver.Remove(...) called")
	state, err := d.loadVariables(info)
	if err != nil {
		return err
	}
	client, err := k8s.InjectedDynamic()
	if err != nil {
		return err
	}
	return capi.DeleteCluster(ctx, client, state)
}

// GetDriverCreateOptions implements driver interface
func (d *OCIOCNEDriver) GetDriverCreateOptions(ctx context.Context) (*types.DriverFlags, error) {
	d.Logger.Infof("capi.driver.GetDriverCreateOptions(...) called")

	driverFlag := types.DriverFlags{
		Options: make(map[string]*types.Flag),
	}
	driverFlag.Options[driverconst.ClusterName] = &types.Flag{
		Type:  types.StringType,
		Usage: "The name of the OCNE Cluster",
	}
	driverFlag.Options[driverconst.PodCIDR] = &types.Flag{
		Type:  types.StringType,
		Usage: "The Kubernetes Pod CIDR block",
		Default: &types.Default{
			DefaultString: "192.168.0.0/16",
		},
	}
	driverFlag.Options[driverconst.ClusterCIDR] = &types.Flag{
		Type:  types.StringType,
		Usage: "The Kubernetes Clister CIDR block",
		Default: &types.Default{
			DefaultString: "10.0.0.0/12",
		},
	}
	driverFlag.Options[driverconst.ControlPlaneShape] = &types.Flag{
		Type:  types.StringType,
		Usage: "The shape of the control plane nodes",
		Default: &types.Default{
			DefaultString: variables.DefaultVMShape,
		},
	}
	driverFlag.Options[driverconst.NodeShape] = &types.Flag{
		Type:  types.StringType,
		Usage: "The shape of the worker nodes",
		Default: &types.Default{
			DefaultString: variables.DefaultVMShape,
		},
	}
	driverFlag.Options[driverconst.ProxyEndpoint] = &types.Flag{
		Type:  types.StringType,
		Usage: "The proxy endpoint to configure on control plane and worker nodes",
	}
	driverFlag.Options[driverconst.ControlPlaneRegistry] = &types.Flag{
		Type:  types.StringType,
		Usage: "The registry to use for control plane images",
		Default: &types.Default{
			DefaultString: variables.DefaultRegistryCNE,
		},
	}
	driverFlag.Options[driverconst.CSIRegistry] = &types.Flag{
		Type:  types.StringType,
		Usage: "The registry to use for CSI images",
		Default: &types.Default{
			DefaultString: variables.DefaultCSIRegistry,
		},
	}
	driverFlag.Options[driverconst.OCICSIImage] = &types.Flag{
		Type:  types.StringType,
		Usage: "The OCI CSI Provider Image",
		Default: &types.Default{
			DefaultString: variables.DefaultOCICSIImage,
		},
	}
	driverFlag.Options[driverconst.CalicoRegistry] = &types.Flag{
		Type:  types.StringType,
		Usage: "The registry to use for calico cni images",
		Default: &types.Default{
			DefaultString: variables.DefaultRegistryCNE,
		},
	}
	driverFlag.Options[driverconst.CalicoTag] = &types.Flag{
		Type:  types.StringType,
		Usage: "The image tag for calico images",
		Default: &types.Default{
			DefaultString: variables.DefaultCalicoTag,
		},
	}
	driverFlag.Options[driverconst.CCMImage] = &types.Flag{
		Type:  types.StringType,
		Usage: "The image for OCI cloud-controller-manager",
		Default: &types.Default{
			DefaultString: variables.DefaultCCMImage,
		},
	}
	driverFlag.Options[driverconst.InstallCalico] = &types.Flag{
		Type:  types.BoolType,
		Usage: "Install Calico addon",
		Default: &types.Default{
			DefaultBool: true,
		},
	}
	driverFlag.Options[driverconst.InstallCCM] = &types.Flag{
		Type:  types.BoolType,
		Usage: "Install CCM addon",
		Default: &types.Default{
			DefaultBool: true,
		},
	}
	driverFlag.Options[driverconst.InstallCSI] = &types.Flag{
		Type:  types.BoolType,
		Usage: "Install CSI addon",
		Default: &types.Default{
			DefaultBool: true,
		},
	}
	driverFlag.Options[driverconst.VerrazzanoImage] = &types.Flag{
		Type:  types.StringType,
		Usage: "The Verrazzano Platform Operator Image",
		Default: &types.Default{
			DefaultString: variables.DefaultVerrazzanoImage,
		},
	}
	driverFlag.Options[driverconst.VerrazzanoResource] = &types.Flag{
		Type:  types.StringType,
		Usage: "The Verrazzano resource to install on the managed cluster",
		Default: &types.Default{
			DefaultString: variables.DefaultVerrazzanoResource,
		},
	}
	driverFlag.Options[driverconst.InstallVerrazzano] = &types.Flag{
		Type:  types.BoolType,
		Usage: "Install Verrazzano addon",
		Default: &types.Default{
			DefaultBool: true,
		},
	}
	driverFlag.Options[driverconst.ETCDImageTag] = &types.Flag{
		Type:  types.StringType,
		Usage: "The image tag to use for ETCD",
		Default: &types.Default{
			DefaultString: variables.DefaultETCDImageTag,
		},
	}
	driverFlag.Options[driverconst.CoreDNSImageTag] = &types.Flag{
		Type:  types.StringType,
		Usage: "The image tag to use for CoreDNS",
		Default: &types.Default{
			DefaultString: variables.DefaultCoreDNSImageTag,
		},
	}
	driverFlag.Options[driverconst.KubernetesVersion] = &types.Flag{
		Type:  types.StringType,
		Usage: "The Kubernetes version that will be used for your master and worker nodes e.g. v1.11.9, v1.12.7",
	}
	driverFlag.Options[driverconst.NodeOCPUs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Optional number of OCPUs for worker nodes",
		Default: &types.Default{
			DefaultInt: variables.DefaultOCICPUs,
		},
	}
	driverFlag.Options[driverconst.ControlPlaneOCPUs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Optional number of OCPUs for control plane nodes",
		Default: &types.Default{
			DefaultInt: variables.DefaultOCICPUs,
		},
	}
	driverFlag.Options[driverconst.NodeMemoryGbs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Optional amount of memory (in GBs) for worker nodes",
		Default: &types.Default{
			DefaultInt: variables.DefaultMemoryGbs,
		},
	}
	driverFlag.Options[driverconst.ControlPlaneMemoryGbs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Optional amount of memory (in GBs) for control plane nodes",
		Default: &types.Default{
			DefaultInt: variables.DefaultMemoryGbs,
		},
	}
	driverFlag.Options[driverconst.ControlPlaneVolumeGbs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Volume size of control plane nodes in Gbs",
		Default: &types.Default{
			DefaultInt: variables.DefaultVolumeGbs,
		},
	}
	driverFlag.Options[driverconst.NodePublicKeyContents] = &types.Flag{
		Type:  types.StringType,
		Usage: "The contents of the SSH public key to use for the nodes",
	}
	driverFlag.Options[driverconst.NumControlPlaneNodes] = &types.Flag{
		Type:  types.IntType,
		Usage: "Number of control plane nodes, default 1",
		Default: &types.Default{
			DefaultInt: 1,
		},
	}
	driverFlag.Options[driverconst.NumWorkerNodes] = &types.Flag{
		Type:  types.IntType,
		Usage: "Number of worker nodes, default 3.",
		Default: &types.Default{
			DefaultInt: 3,
		},
	}
	driverFlag.Options[driverconst.NodeVolumeGbs] = &types.Flag{
		Type:  types.IntType,
		Usage: "Volume size of worker nodes in Gbs",
		Default: &types.Default{
			DefaultInt: variables.DefaultVolumeGbs,
		},
	}
	driverFlag.Options[driverconst.ImageDisplayName] = &types.Flag{
		Type:  types.StringType,
		Usage: "Image for cluster nodes",
	}
	driverFlag.Options[driverconst.CloudCredentialId] = &types.Flag{
		Type:  types.StringType,
		Usage: "The cloud credential id",
	}
	driverFlag.Options[driverconst.Region] = &types.Flag{
		Type:  types.StringType,
		Usage: "The cloud provider region",
	}
	driverFlag.Options[driverconst.CompartmentID] = &types.Flag{
		Type:  types.StringType,
		Usage: "The OCID of the compartment in which to create resrouces (VCN, worker nodes, etc.)",
	}
	driverFlag.Options[driverconst.VcnID] = &types.Flag{
		Type:  types.StringType,
		Usage: "The OCID of an existing virtual network to be used for cluster creation",
	}
	driverFlag.Options[driverconst.UsePVNodeEncryption] = &types.Flag{
		Type:  types.BoolType,
		Usage: "Whether to use PV Node encryption",
		Default: &types.Default{
			DefaultBool: variables.DefaultNodePVTransitEncryption,
		},
	}
	driverFlag.Options[driverconst.WorkerNodeSubnet] = &types.Flag{
		Type:  types.StringType,
		Usage: "OCID for node pool subnet",
	}
	driverFlag.Options[driverconst.ControlPlaneSubnet] = &types.Flag{
		Type:  types.StringType,
		Usage: "OCID for control plane subnet",
	}
	driverFlag.Options[driverconst.LoadBalancerSubnet] = &types.Flag{
		Type:  types.StringType,
		Usage: "OCID for load balancer subnet",
	}
	driverFlag.Options[driverconst.PreOCNECommands] = &types.Flag{
		Type:  types.StringSliceType,
		Usage: "Commands to run before OCNE initialization",
		Default: &types.Default{
			DefaultStringSlice: &types.StringSlice{Value: []string{}}, // avoid nil value for init
		},
	}
	driverFlag.Options[driverconst.PostOCNECommands] = &types.Flag{
		Type:  types.StringSliceType,
		Usage: "Commands to run after OCNE initialization",
		Default: &types.Default{
			DefaultStringSlice: &types.StringSlice{Value: []string{}}, // avoid nil value for init
		},
	}

	return &driverFlag, nil
}

// GetDriverUpdateOptions implements driver interface
func (d *OCIOCNEDriver) GetDriverUpdateOptions(ctx context.Context) (*types.DriverFlags, error) {
	d.Logger.Infof("capi.driver.GetDriverUpdateOptions(...) called")

	driverFlag := types.DriverFlags{
		Options: make(map[string]*types.Flag),
	}
	driverFlag.Options["quantity-per-subnet"] = &types.Flag{
		Type:  types.IntType,
		Usage: "The updated number of worker nodes in each subnet / availability-domain to update. 1 (default) means no updates",
	}
	driverFlag.Options["kubernetes-version"] = &types.Flag{
		Type:  types.StringType,
		Usage: "The updated Kubernetes version",
	}
	return &driverFlag, nil
}

// Create implements driver interface
func (d *OCIOCNEDriver) Create(ctx context.Context, opts *types.DriverOptions, _ *types.ClusterInfo) (*types.ClusterInfo, error) {
	d.Logger.Infof("capi.driver.Create(...) called")
	state, err := variables.NewFromOptions(ctx, opts)
	if err != nil {
		d.Logger.Errorf("error creating state %v", err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}

	/*
	* The ClusterInfo includes the following information Version, ServiceAccountToken,Endpoint, username, password, etc
	 */
	clusterInfo := &types.ClusterInfo{}
	err = storeVariables(clusterInfo, state)
	if err != nil {
		d.Logger.Errorf("error storing state %v", err)
		return clusterInfo, err
	}

	if err := doCreateOrUpdate(ctx, state); err != nil {
		d.Logger.Errorf("Driver.Create: %v", err)
		return clusterInfo, err
	}
	return clusterInfo, nil
}

// Update implements driver interface
func (d *OCIOCNEDriver) Update(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions) (*types.ClusterInfo, error) {
	d.Logger.Infof("capi.driver.Update(...) called")

	state, err := d.loadVariables(clusterInfo)
	if err != nil {
		return nil, err
	}
	state.SetUpdateOptions(opts)
	if err := state.Validate(); err != nil {
		return clusterInfo, err
	}
	if err := storeVariables(clusterInfo, state); err != nil {
		return clusterInfo, err
	}
	if err := doCreateOrUpdate(ctx, state); err != nil {
		d.Logger.Errorf("Driver.Update: %v", err)
		return clusterInfo, err
	}
	return clusterInfo, nil
}

func (d *OCIOCNEDriver) PostCheck(ctx context.Context, info *types.ClusterInfo) (*types.ClusterInfo, error) {
	d.Logger.Infof("capi.driver.PostCheck(...) called")

	state, err := d.loadVariables(info)
	if err != nil {
		return info, err
	}
	capiClusterKubeConfig, err := state.GetCAPIClusterKubeConfig(ctx, state)
	if err != nil {
		return info, err
	}

	info.Version = state.KubernetesVersion
	info.Username = ""
	info.Password = ""
	info.ClientCertificate = ""
	info.ClientKey = ""
	info.NodeCount = state.NodeReplicas + state.ControlPlaneReplicas
	info.Metadata["nodePool"] = state.Name + "-1"
	if len(capiClusterKubeConfig.Clusters) > 0 {
		cluster := capiClusterKubeConfig.Clusters[0].Cluster
		info.Endpoint = cluster.Server
		info.RootCaCertificate = cluster.CertificateAuthorityData
	}

	// Use as a temporary token while we generate a service account.
	if len(capiClusterKubeConfig.Users) > 0 {
		if capiClusterKubeConfig.Users[0].User.Token != "" {
			info.ServiceAccountToken = capiClusterKubeConfig.Users[0].User.Token
		}
		// TODO handle info.ExecCredential when it is supported by Rancher
		// https://github.com/rancher/rancher/issues/24135
	}

	kubeConfigBytes, err := yaml.Marshal(&capiClusterKubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshaling internalConfig: %v", err)
	}

	ki, err := k8s.NewInterfaceForKubeconfig(kubeConfigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset for managed cluster %s: %v", state.Name, err)
	}

	d.Logger.Infof("Creating service account token for cluster %v", state.Name)
	info.ServiceAccountToken, err = d.generateServiceAccountToken(ctx, ki)
	if err != nil {
		return nil, fmt.Errorf("could not generate service account token: %v", err)
	}

	di, err := k8s.NewDynamicForKubeconfig(kubeConfigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic clientset for managed cluster %s: %v", state.Name, err)
	}

	d.Logger.Infof("Installing Verrazzano on cluster %v", state.Name)
	if err := capi.InstallVerrazzano(ctx, ki, di, state); err != nil {
		return nil, fmt.Errorf("failed to install Verrazzano on managed cluster %s: %v", state.Name, err)
	}
	d.Logger.Infof("+++ returning from PostCheck +++")
	return info, nil
}

func (d *OCIOCNEDriver) GetClusterSize(_ context.Context, info *types.ClusterInfo) (*types.NodeCount, error) {
	v, err := d.loadVariables(info)
	if err != nil {
		return nil, err
	}
	return v.NodeCount(), nil
}

func (d *OCIOCNEDriver) GetVersion(_ context.Context, info *types.ClusterInfo) (*types.KubernetesVersion, error) {
	v, err := d.loadVariables(info)
	if err != nil {
		return nil, err
	}
	return v.Version(), nil
}

func (d *OCIOCNEDriver) SetClusterSize(ctx context.Context, info *types.ClusterInfo, count *types.NodeCount) error {
	d.Logger.Infof("capi.driver.SetClusterSize(...) called")
	state, err := d.loadVariables(info)
	if err != nil {
		return err
	}

	state.NodeReplicas = count.Count
	if err := storeVariables(info, state); err != nil {
		d.Logger.Errorf("Failed to save new node group size: %v", err)
		return err
	}
	client, err := k8s.InjectedDynamic()
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}
	err = capi.CreateOrUpdateNodeGroup(ctx, client, state)
	if err != nil {
		return fmt.Errorf("failed to create objects: %v", err)
	}
	err = capi.WaitForCAPIClusterReady(ctx, client, state)
	if err != nil {
		return fmt.Errorf("failed to create CAPI Cluster: %v", err)
	}
	return nil
}

// SetVersion sets the Kubernetes Version of cluster
func (d *OCIOCNEDriver) SetVersion(ctx context.Context, info *types.ClusterInfo, version *types.KubernetesVersion) error {
	d.Logger.Infof("capi.driver.SetVersion(...) called")
	_, err := d.loadVariables(info)
	if err != nil {
		return err
	}
	// update version is currently asynchronous
	d.Logger.Info("Cluster version (masters and node pools) updated successfully")
	return nil
}

func (d *OCIOCNEDriver) GetCapabilities(_ context.Context) (*types.Capabilities, error) {
	d.Logger.Infof("capi.driver.GetCapabilities(...) called")
	return &d.driverCapabilities, nil
}

func (d *OCIOCNEDriver) ETCDSave(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) error {
	d.Logger.Infof("capi.driver.ETCDSave(...) called")
	return fmt.Errorf("ETCD backup operations are not implemented")
}

func (d *OCIOCNEDriver) ETCDRestore(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) (*types.ClusterInfo, error) {
	d.Logger.Infof("capi.driver.ETCDRestore(...) called")
	return nil, fmt.Errorf("ETCD backup operations are not implemented")
}

func (d *OCIOCNEDriver) ETCDRemoveSnapshot(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) error {
	d.Logger.Infof("capi.driver.ETCDRemoveSnapshot(...) called")
	return fmt.Errorf("ETCD backup operations are not implemented")
}

func (d *OCIOCNEDriver) GetK8SCapabilities(ctx context.Context, options *types.DriverOptions) (*types.K8SCapabilities, error) {
	d.Logger.Infof("capi.driver.GetK8SCapabilities(...) called")
	capabilities := &types.K8SCapabilities{
		L4LoadBalancer: &types.LoadBalancerCapabilities{
			Enabled:              true,
			Provider:             "OCILB",
			ProtocolsSupported:   []string{"TCP", "HTTP/1.0", "HTTP/1.1"},
			HealthCheckSupported: true,
		},
	}
	return capabilities, nil
}

func (d *OCIOCNEDriver) RemoveLegacyServiceAccount(ctx context.Context, info *types.ClusterInfo) error {
	d.Logger.Infof("capi.driver.RemoveLegacyServiceAccount(...) called")
	return nil
}

func storeVariables(info *types.ClusterInfo, v *variables.Variables) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not marshal state: %v", err)
	}

	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}

	info.Metadata[metadataKey] = string(bytes)
	return nil
}

func (d *OCIOCNEDriver) loadVariables(info *types.ClusterInfo) (*variables.Variables, error) {
	d.Logger.Infof("capi.driver.GetState(...) called")
	state := &variables.Variables{}
	err := json.Unmarshal([]byte(info.Metadata[metadataKey]), &state)
	return state, err
}

// GenerateServiceAccountToken generate a serviceAccountToken for clusterAdmin given a clientset
func (d *OCIOCNEDriver) generateServiceAccountToken(ctx context.Context, clientset kubernetes.Interface) (string, error) {

	token := ""
	namespace := "default"
	name := "kontainer-engine-olcne"

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	d.Logger.Debugf("[oraclecontainerengine] Kubernetes server version: %s", serverVersion)

	serviceAccount := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: name}}

	// Create new service account, if it does not exist already
	_, err = clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return "", err
		}
	}

	serviceAccount, err = clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Template for an authentication token secret bound to the service account
	secretTemplate := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccount.Name + "-token",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "ServiceAccount",
					Name:       serviceAccount.Name,
					UID:        serviceAccount.UID,
				},
			},
			Annotations: map[string]string{
				v1.ServiceAccountNameKey: serviceAccount.Name,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secretTemplate, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return "", err
		}
	}
	// wait a few seconds for authentication token to populate
	time.Sleep(10 * time.Second)

	secretObj, err := clientset.CoreV1().Secrets(namespace).Get(ctx, serviceAccount.Name+"-token", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Create new cluster-role-bindings, if it does not exist already
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, APIGroup: "", Name: name, Namespace: namespace}},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return "", err
		}
	}

	// Look up cluster role binding
	_, err = clientset.RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBinding.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting cluster role binding: %v", err)
	}

	// get the bearer token from the token-secret
	if byteToken, ok := secretObj.Data[v1.ServiceAccountTokenKey]; ok {
		token = string(byteToken)
		return token, nil
	}

	return "", fmt.Errorf("error getting authentication token from secret: %s", secretObj.Name)
}

func doCreateOrUpdate(ctx context.Context, state *variables.Variables) error {
	dynamicInterface, err := k8s.InjectedDynamic()
	kubernetesInterface, err := k8s.InjectedInterface()
	if err != nil {
		return fmt.Errorf("failed to get dynamicInterface: %v", err)
	}
	err = capi.CreateOrUpdateAllObjects(ctx, kubernetesInterface, dynamicInterface, state)
	if err != nil {
		return fmt.Errorf("failed to create objects: %v", err)
	}
	err = capi.WaitForCAPIClusterReady(ctx, dynamicInterface, state)
	if err != nil {
		return fmt.Errorf("failed to create CAPI Cluster: %v", err)
	}
	return nil
}

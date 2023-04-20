// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	gvr "github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

const (
	ociTenancyField              = "tenancy"
	ociUserField                 = "user"
	ociFingerprintField          = "fingerprint"
	ociRegionField               = "region"
	ociPassphraseField           = "passphrase"
	ociKeyField                  = "key"
	ociUseInstancePrincipalField = "useInstancePrincipal"
)

const (
	clusterPhaseProvisioned = "Provisioned"
	machinePhaseRunning     = "Running"
)

type CAPIClient struct {
	capiTimeout         time.Duration
	capiPollingInterval time.Duration

	verrazzanoTimeout         time.Duration
	verrazzanoPollingInterval time.Duration
}

func NewCAPIClient() *CAPIClient {
	return &CAPIClient{
		capiTimeout:               1 * time.Hour,
		capiPollingInterval:       30 * time.Second,
		verrazzanoTimeout:         5 * time.Minute,
		verrazzanoPollingInterval: 10 * time.Second,
	}
}

func (c *CAPIClient) DeleteHangingResources(ctx context.Context, p dynamic.Interface, cruResult *CreateOrUpdateResult, namespace string) error {
	return deleteWorkerObjects(ctx, p, cruResult, namespace)
}

// CreateOrUpdateAllObjects creates or updates all cluster result
func (c *CAPIClient) CreateOrUpdateAllObjects(ctx context.Context, kubernetesInterface kubernetes.Interface, dynamicInterface dynamic.Interface, v *variables.Variables) (*CreateOrUpdateResult, error) {
	if err := createOrUpdateCAPISecret(ctx, v, kubernetesInterface); err != nil {
		return nil, fmt.Errorf("failed to create CAPI credentials: %v", err)
	}
	cruResult, err := createOrUpdateObjects(ctx, dynamicInterface, object.CreateObjects(v), v)
	if err != nil {
		return cruResult, err
	}
	return cruResult, c.WaitForCAPIClusterReady(ctx, dynamicInterface, v)
}

// createOrUpdateCAPISecret creates the CAPI secret if it does not already exist
// if the secret exists, update it in place with the new credentials
func createOrUpdateCAPISecret(ctx context.Context, v *variables.Variables, client kubernetes.Interface) error {
	data := map[string][]byte{
		ociTenancyField:              []byte(v.Tenancy),
		ociUserField:                 []byte(v.User),
		ociFingerprintField:          []byte(v.Fingerprint),
		ociRegionField:               []byte(v.Region),
		ociPassphraseField:           []byte(v.PrivateKeyPassphrase),
		ociKeyField:                  []byte(strings.TrimSpace(strings.ReplaceAll(v.PrivateKey, "\\n", "\n"))),
		ociUseInstancePrincipalField: []byte("false"),
	}
	current, err := client.CoreV1().Secrets(v.CAPIOCINamespace).Get(ctx, v.CAPICredentialName, metav1.GetOptions{})
	if err != nil {
		// Create if not exists
		if apierrors.IsNotFound(err) {
			_, err := client.CoreV1().Secrets(v.CAPIOCINamespace).Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: v.CAPICredentialName,
					Labels: map[string]string{
						"cluster.x-k8s.io/provider": "infrastructure-oci",
					},
				},
				Data: data,
			}, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// update secret in place
	current.Data = data
	_, err = client.CoreV1().Secrets(v.CAPIOCINamespace).Update(ctx, current, metav1.UpdateOptions{})
	return err
}

func createOrUpdateObjects(ctx context.Context, dynamicInterface dynamic.Interface, objects []object.Object, v *variables.Variables) (*CreateOrUpdateResult, error) {
	cruResult := NewCreateOrUpdateResult()
	for _, o := range objects {
		partialResult, err := createOrUpdateObject(ctx, dynamicInterface, o, v)
		if err != nil {
			return cruResult, fmt.Errorf("failed to create Object %s/%s/%s: %v", o.GVR.Group, o.GVR.Version, o.GVR.Resource, err)
		}
		cruResult.Merge(partialResult)
	}
	return cruResult, nil
}

func createOrUpdateObject(ctx context.Context, client dynamic.Interface, o object.Object, v *variables.Variables) (*CreateOrUpdateResult, error) {
	return cruObject(ctx, client, o, v, true)
}

// cruObject create or update an object
func cruObject(ctx context.Context, client dynamic.Interface, o object.Object, v *variables.Variables, update bool) (*CreateOrUpdateResult, error) {
	cruResult := NewCreateOrUpdateResult()
	toCreateObject, err := loadTextTemplate(o, *v)
	if err != nil {
		return cruResult, err
	}

	for _, u := range toCreateObject {
		// Try to fetch existing object
		existingObject, err := client.Resource(o.GVR).Namespace(u.GetNamespace()).Get(ctx, u.GetName(), metav1.GetOptions{})
		if err != nil {
			// if object doesn't exist, try to create it
			if apierrors.IsNotFound(err) {
				if err := createIfNotExists(ctx, client, o.GVR, &u); err != nil {
					return cruResult, err
				}
			} else {
				return cruResult, err
			}
		} else if update { // If the Object exists, merge with existingObject and do an update
			mergedObject := mergeUnstructured(existingObject, &u, o.LockedFields)
			if err != nil {
				return cruResult, err
			}
			_, err = client.Resource(o.GVR).Namespace(mergedObject.GetNamespace()).Update(context.TODO(), mergedObject, metav1.UpdateOptions{})
			if err != nil {
				return cruResult, err
			}
		}

		cruResult.Add(o.GVR.Resource, &u)
	}

	return cruResult, nil
}

func createIfNotExists(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, u *unstructured.Unstructured) error {
	_, err := client.Resource(gvr).Namespace(u.GetNamespace()).Create(ctx, u, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// DeleteCluster deletes the cluster
func DeleteCluster(ctx context.Context, client dynamic.Interface, v *variables.Variables) error {
	deleteFromTmpl := func(o object.Object) error {
		us, err := loadTextTemplate(o, *v)
		if err != nil {
			return err
		}
		return deleteUnstructureds(ctx, client, o.GVR, us)
	}
	return deleteFromTmpl(object.Object{
		GVR:  gvr.Cluster,
		Text: templates.Cluster,
	})
}

func deleteUnstructureds(ctx context.Context, di dynamic.Interface, gvr schema.GroupVersionResource, us []unstructured.Unstructured) error {
	for _, u := range us {
		return deleteIfExists(ctx, di, gvr, u.GetName(), u.GetNamespace())
	}
	return nil
}

func deleteIfExists(ctx context.Context, di dynamic.Interface, gvr schema.GroupVersionResource, name, namespace string) error {
	err := di.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func deleteWorkerObjects(ctx context.Context, di dynamic.Interface, cruResult *CreateOrUpdateResult, namespace string) error {
	fieldSelector := fmt.Sprintf("metadata.namespace=%s", namespace)
	// cleanup machine deployments
	mds, err := di.Resource(gvr.MachineDeployment).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return err
	}

	// Delete unused machinedeployments
	for _, md := range mds.Items {
		_, err := deleteIfContains(ctx, di, cruResult, gvr.MachineDeployment, &md)
		if err != nil {
			return err
		}
		// delete associated ocimachinetemplate if it exists
		templateName, err := object.NestedField(md.Object, "spec", "template", "spec", "infrastructureRef", "name")
		if ociMachineTemplate, ok := templateName.(string); ok && err == nil {
			if err := deleteIfExists(ctx, di, gvr.OCIMachineTemplate, ociMachineTemplate, namespace); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteIfContains(ctx context.Context, di dynamic.Interface, cruResult *CreateOrUpdateResult, gvr schema.GroupVersionResource, u *unstructured.Unstructured) (bool, error) {
	if cruResult.Contains(gvr.Resource, u) {
		return true, deleteUnstructureds(ctx, di, gvr, []unstructured.Unstructured{*u})
	}
	return false, nil
}

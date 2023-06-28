// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"errors"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	"k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	verrazzanoInstallNamespace = "verrazzano-install"
	verrazzanoMCNamespace      = "verrazzano-mc"
	verrazzanoPlatformOperator = "verrazzano-platform-operator"
	verrazzanoModuleOperator   = "verrazzano-module-operator"
)

func (c *CAPIClient) InstallModules(ctx context.Context, ki kubernetes.Interface, di dynamic.Interface, v *variables.Variables) error {
	if err := c.waitForModuleOperatorReady(ctx, ki); err != nil {
		return err
	}
	_, err := createOrUpdateObjects(ctx, di, object.Modules(v), v)
	return err
}

func (c *CAPIClient) InstallAndRegisterVerrazzano(ctx context.Context, ki kubernetes.Interface, di, adminDi dynamic.Interface, v *variables.Variables) error {
	if !v.InstallVerrazzano || v.VerrazzanoResource == "" {
		return nil
	}
	if err := c.waitForVerrazzanoPlatformOperator(ctx, ki); err != nil {
		return err
	}

	// Create the Verrazzano Resource
	if err := createOrUpdateVerrazzano(ctx, di, v); err != nil {
		return fmt.Errorf("verrazzano install/update error: %v", err)
	}
	// Create the Verrazzano Managed Cluster Resource
	if _, err := createOrUpdateObject(ctx, adminDi, object.Object{
		Text: templates.VMC,
	}, v); err != nil && !meta.IsNoMatchError(err) {
		// IsNoMatchError ignored in case cluster-operator not installed, and the VMC CRD is not present
		return fmt.Errorf("vmc registration error: %v", err)
	}
	return nil
}

// DeleteVerrazzanoResources deletes the Verrazzano resource on the managed cluster, and the VerrazzanoManagedCluster on the admin cluster
func (c *CAPIClient) DeleteVerrazzanoResources(ctx context.Context, managedDi, adminDi dynamic.Interface, v *variables.Variables) error {
	if !v.InstallVerrazzano {
		return nil
	}
	if err := deleteVMC(ctx, adminDi, v); err != nil {
		return err
	}
	return c.deleteVZ(ctx, managedDi, v)
}

func deleteVMC(ctx context.Context, adminDi dynamic.Interface, v *variables.Variables) error {
	// Clean up the admin cluster VMC
	err := adminDi.Resource(gvr.VerrazzanoManagedCluster).Namespace(verrazzanoMCNamespace).Delete(ctx, v.Name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		// IsNoMatchError ignored in case cluster-operator not installed, and the VMC CRD is not present
		return fmt.Errorf("failed to delete Verrazzano Managed cluster: %v", err)
	}

	return nil
}

func (c *CAPIClient) deleteVZ(ctx context.Context, managedDi dynamic.Interface, v *variables.Variables) error {
	// Load the VZ from template and clean the managed cluster VZ
	us, err := loadTextTemplate(object.Object{
		Text: v.VerrazzanoResource,
	}, *v)
	if err != nil {
		return err
	}
	if len(us) != 1 {
		return fmt.Errorf("expected 1 Verrazzano resource from template, got %d", len(us))
	}
	vz := us[0]
	err = managedDi.Resource(gvr.Verrazzano).Namespace(vz.GetNamespace()).Delete(ctx, vz.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete Verrazzano resource: %v", err)
	}

	//
	return errors.New("deleting Verrazzano resource")
}

func createOrUpdateVerrazzano(ctx context.Context, di dynamic.Interface, v *variables.Variables) error {
	// Only add the verrazzano version if it has changed
	if _, err := cruObject(ctx, di, object.Object{
		Text: v.VerrazzanoResource,
	}, v, func(u *unstructured.Unstructured) error {
		versionString, b, err := unstructured.NestedString(u.Object, "status", "version")
		// if no version found, don't force version update
		if err != nil || !b {
			return nil
		}
		if versionString == v.VerrazzanoVersion {
			return nil
		}
		return unstructured.SetNestedField(u.Object, v.VerrazzanoVersion, "spec", "version")
	}); err != nil {
		return err
	}

	return nil
}

func (c *CAPIClient) waitForDeployment(ctx context.Context, ki kubernetes.Interface, namespace, name string) error {
	endTime := time.Now().Add(c.verrazzanoTimeout)
	for {
		time.Sleep(c.verrazzanoPollingInterval)
		vpoDeployment, err := ki.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if isDeploymentReady(vpoDeployment) {
			return nil
		}
		if time.Now().After(endTime) {
			return fmt.Errorf("timed out waiting for deployment %s/%s to be ready", namespace, name)
		}
	}
}

func (c *CAPIClient) waitForModuleOperatorReady(ctx context.Context, ki kubernetes.Interface) error {
	return c.waitForDeployment(ctx, ki, verrazzanoModuleOperator, verrazzanoModuleOperator)
}

func (c *CAPIClient) waitForVerrazzanoPlatformOperator(ctx context.Context, ki kubernetes.Interface) error {
	return c.waitForDeployment(ctx, ki, verrazzanoInstallNamespace, verrazzanoPlatformOperator)
}

func isDeploymentReady(deployment *v1.Deployment) bool {
	if deployment == nil || deployment.Spec.Replicas == nil {
		return false
	}
	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas && deployment.Status.AvailableReplicas == *deployment.Spec.Replicas
}

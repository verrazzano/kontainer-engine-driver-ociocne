// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	"k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	verrazzanoInstallNamespace = "verrazzano-install"
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

	// Create the Verrazzano Resource if not exists
	if _, err := cruObject(ctx, di, object.Object{
		Text: v.VerrazzanoResource,
	}, v, false); err != nil {
		return fmt.Errorf("install error: %v", err)
	}

	// Create the Verrazzano Managed Cluster Resource if not exists
	if _, err := cruObject(ctx, adminDi, object.Object{
		Text: templates.VMC,
	}, v, false); err != nil {
		return fmt.Errorf("registration error: %v", err)
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

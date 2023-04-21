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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	verrazzanoInstallNamespace = "verrazzano-install"
	verrazzanoPlatformOperator = "verrazzano-platform-operator"
)

func (c *CAPIClient) InstallAndRegisterVerrazzano(ctx context.Context, ki kubernetes.Interface, di, adminDi dynamic.Interface, v *variables.Variables) error {
	if !v.InstallVerrazzano || v.VerrazzanoResource == "" {
		return nil
	}
	if err := c.waitForVerrazzanoPlatformOperator(ctx, ki); err != nil {
		return err
	}

	// Create the Verrazzano Resource if not exists
	if _, err := cruObject(ctx, di, object.Object{
		GVR:  gvr.Verrazzano,
		Text: v.VerrazzanoResource,
	}, v, false); err != nil {
		return fmt.Errorf("install error: %v", err)
	}

	// Create the Verrazzano Managed Cluster Resource if not exists
	if _, err := cruObject(ctx, adminDi, object.Object{
		GVR:  gvr.VerrazzanoManagedCluster,
		Text: templates.VMC,
	}, v, false); err != nil {
		return fmt.Errorf("registration error: %v", err)
	}
	return nil
}

func (c *CAPIClient) waitForVerrazzanoPlatformOperator(ctx context.Context, ki kubernetes.Interface) error {
	endTime := time.Now().Add(c.verrazzanoTimeout)
	for {
		time.Sleep(c.verrazzanoPollingInterval)
		vpoDeployment, err := ki.AppsV1().Deployments(verrazzanoInstallNamespace).Get(ctx, verrazzanoPlatformOperator, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if isVPOReady(vpoDeployment) {
			return nil
		}
		if time.Now().After(endTime) {
			return errors.New("timed out waiting for Verrazzano Platform Operator to be ready")
		}
	}
}

func isVPOReady(deployment *v1.Deployment) bool {
	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas && deployment.Status.AvailableReplicas == *deployment.Spec.Replicas
}

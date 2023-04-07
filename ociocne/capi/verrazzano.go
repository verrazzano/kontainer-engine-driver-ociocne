// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"errors"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	verrazzanoReadyTimeout    = 5 * time.Minute
	verrazzanoPollingInterval = 10 * time.Second

	verrazzanoInstallNamespace = "verrazzano-install"
	verrazzanoPlatformOperator = "verrazzano-platform-operator"
)

func InstallAndRegisterVerrazzano(ctx context.Context, ki kubernetes.Interface, di, adminDi dynamic.Interface, v *variables.Variables) error {
	if !v.InstallVerrazzano || v.VerrazzanoResource == "" {
		return nil
	}
	if err := waitForVerrazzanoPlatformOperator(ctx, ki); err != nil {
		return err
	}

	// Create the Verrazzano Resource
	if err := createOrUpdateObject(ctx, di, object{
		gvr:  gvr.Verrazzano,
		text: v.VerrazzanoResource,
	}, v); err != nil {
		return err
	}

	// Create the Verrazzano Managed Cluster Resource
	return createOrUpdateObject(ctx, di, object{
		gvr:  gvr.VerrazzanoManagedCluster,
		text: templates.VMC,
	}, v)

}

func waitForVerrazzanoPlatformOperator(ctx context.Context, ki kubernetes.Interface) error {
	endTime := time.Now().Add(verrazzanoReadyTimeout)
	for {
		time.Sleep(verrazzanoPollingInterval)
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

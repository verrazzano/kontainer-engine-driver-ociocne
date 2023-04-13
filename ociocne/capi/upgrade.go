// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// UpdateCluster upgrades the CAPI cluster by going through the following stages:
// 1. update the CAPI credentials using the cloud credential. This keeps the cloud credential up-to-date
// 2. update the control plane, and then wait for the control plane to be ready
// 3. update the worker nodes, and then wait for the worker nodes to be ready
// 4. update the remaining cluster resources, and then wait for the cluster to be ready
func UpdateCluster(ctx context.Context, ki kubernetes.Interface, di dynamic.Interface, v *variables.Variables) error {
	// update the CAPI credentials if necessary
	if err := createOrUpdateCAPISecret(ctx, v, ki); err != nil {
		return fmt.Errorf("failed to create CAPI credentials: %v", err)
	}

	// update the control plane nodes
	if err := createOrUpdateObjects(ctx, di, object.ControlPlane, v); err != nil {
		return fmt.Errorf("error updating control plane: %v", err)
	}
	if err := WaitForCAPIClusterReady(ctx, di, v); err != nil {
		return err
	}

	// update the worker nodes
	if err := createOrUpdateObjects(ctx, di, object.Workers, v); err != nil {
		return fmt.Errorf("error updating workers: %v", err)
	}
	if err := WaitForCAPIClusterReady(ctx, di, v); err != nil {
		return err
	}

	// update the remaining capi resources
	if err := createOrUpdateObjects(ctx, di, object.UpdateObjects(v), v); err != nil {
		return fmt.Errorf("error updating cluster resources: %v", err)
	}
	return WaitForCAPIClusterReady(ctx, di, v)
}

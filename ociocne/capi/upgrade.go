// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/client-go/dynamic"
)

func UpgradeClusterVersion(ctx context.Context, di dynamic.Interface, v *variables.Variables) error {
	if err := createOrUpdateObjects(ctx, di, object.ControlPlane, v); err != nil {
		return fmt.Errorf("error updating control plane kubernetes version: %v", err)
	}
	if err := WaitForCAPIClusterReady(ctx, di, v); err != nil {
		return err
	}
	if err := createOrUpdateObjects(ctx, di, object.Workers, v); err != nil {
		return fmt.Errorf("error updating worker kubernetes version: %v", err)
	}
	return WaitForCAPIClusterReady(ctx, di, v)
}

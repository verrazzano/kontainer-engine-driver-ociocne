// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	"k8s.io/client-go/dynamic"
)

func UpgradeClusterVersion(ctx context.Context, di dynamic.Interface, v *variables.Variables, version string) error {
	return nil
}

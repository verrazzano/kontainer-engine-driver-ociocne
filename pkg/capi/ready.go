// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"context"
	"errors"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"time"
)

func IsCAPIClusterReady(ctx context.Context, client dynamic.Interface, state *variables.Variables) error {
	cluster, err := client.Resource(gvr.Cluster).Namespace(state.Namespace).Get(ctx, state.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if isClusterReady(cluster) {
		machinesReady, err := areMachinesReady(ctx, client, state)
		if err != nil {
			return errors.New("Waiting for nodes to be ready")
		}
		if machinesReady {
			return nil
		}
	}
	return errors.New("Waiting for cluster to be ready")
}

// WaitForCAPIClusterReady waits for the CAPI cluster resource to reach "Ready" status, and its Machines
func (c *CAPIClient) WaitForCAPIClusterReady(ctx context.Context, client dynamic.Interface, state *variables.Variables) error {
	endTime := time.Now().Add(c.capiTimeout)
	for {
		time.Sleep(c.capiPollingInterval)
		cluster, err := client.Resource(gvr.Cluster).Namespace(state.Namespace).Get(ctx, state.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if isClusterReady(cluster) {
			machinesReady, err := areMachinesReady(ctx, client, state)
			if err != nil {
				return err
			}
			if machinesReady {
				return nil
			}
		}
		if time.Now().After(endTime) {
			return fmt.Errorf("timed out waiting for cluster %s to create", state.Name)
		}
	}
}

func isClusterReady(cluster *unstructured.Unstructured) bool {
	controlPlaneReady, err := object.NestedField(cluster.Object, "status", "controlPlaneReady")
	if err != nil {
		return false
	}
	controlPlaneReadyBool, ok := controlPlaneReady.(bool)
	if !ok || !controlPlaneReadyBool {
		return false
	}

	infrastructureReady, err := object.NestedField(cluster.Object, "status", "infrastructureReady")
	if err != nil {
		return false
	}
	infrastructureReadyBool, ok := infrastructureReady.(bool)
	if !ok || !infrastructureReadyBool {
		return false
	}

	phase, err := object.NestedField(cluster.Object, "status", "phase")
	if err != nil {
		return false
	}
	phaseString, ok := phase.(string)
	if !ok || phaseString != clusterPhaseProvisioned {
		return false
	}
	return true
}

func areMachinesReady(ctx context.Context, client dynamic.Interface, state *variables.Variables) (bool, error) {
	machineList, err := client.Resource(gvr.Machine).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				"cluster.x-k8s.io/cluster-name": state.Name,
			},
		}),
	})
	if err != nil {
		return false, err
	}

	if machineList == nil || len(machineList.Items) < 1 {
		return false, nil
	}

	for _, machine := range machineList.Items {
		phase, err := object.NestedField(machine.Object, "status", "phase")
		if err != nil {
			return false, nil
		}
		phaseString, ok := phase.(string)
		if !ok {
			return false, nil
		}
		if phaseString != machinePhaseRunning {
			return false, nil
		}
	}
	return true, nil
}

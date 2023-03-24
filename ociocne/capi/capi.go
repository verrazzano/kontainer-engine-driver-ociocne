// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/gvr"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/templates"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/ociocne/variables"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"text/template"
	"time"
)

const (
	clusterCreationTimeout         = 1 * time.Hour
	clusterCreationPollingInterval = 30 * time.Second
)

type object struct {
	gvr  schema.GroupVersionResource
	text string
}

var objects = []object{
	{gvr.ConfigMap, templates.CCMConfigMap},
	{gvr.ConfigMap, templates.CSIConfigMap},
	{gvr.ConfigMap, templates.CalicoConfigMap},
	{gvr.ClusterResourceSet, templates.CalicoResourceSet},
	{gvr.ClusterResourceSet, templates.CCMResourceSet},
	{gvr.ClusterResourceSet, templates.CSIResourceSet},
	{gvr.Cluster, templates.Cluster},
	{gvr.OCICluster, templates.OCICluster},
	{gvr.KubeadmConfigTemplate, templates.OCNEConfigTemplate},
	{gvr.KubeadmControlPlane, templates.OCNEControlPlane},
	{gvr.MachineDeployment, templates.MachineDeployment},
	{gvr.OCIMachineTemplate, templates.OCIMachineTemplate},
	{gvr.OCIMachineTemplate, templates.OCIControlPlaneMachineTemplate},
}

//CreateOrUpdateAllObjects creates or updates all cluster objects
func CreateOrUpdateAllObjects(ctx context.Context, client dynamic.Interface, v *variables.Variables) error {
	for _, o := range objects {
		if err := createOrUpdateObject(ctx, client, o, v); err != nil {
			return err
		}
	}
	return nil
}

//CreateOrUpdateNodeGroup creates or updates the worker node group replica count
func CreateOrUpdateNodeGroup(ctx context.Context, client dynamic.Interface, v *variables.Variables) error {
	return createOrUpdateObject(ctx, client, object{
		gvr.MachineDeployment,
		templates.MachineDeployment,
	}, v)
}

func createOrUpdateObject(ctx context.Context, client dynamic.Interface, o object, v *variables.Variables) error {
	toCreateObject, err := loadTextTemplate(o, *v)
	if err != nil {
		return err
	}

	// Check if the object already exists. Create it if it does not exist and return
	existingObject, err := client.Resource(o.gvr).Namespace(toCreateObject.GetNamespace()).Get(ctx, toCreateObject.GetName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return createIfNotExists(ctx, client, o.gvr, toCreateObject)
		}
	}

	// If the object exists, merge toCreateObject with existingObject, and do an update
	mergedObject := mergeUnstructured(existingObject, toCreateObject)
	if err != nil {
		return err
	}
	_, err = client.Resource(o.gvr).Namespace(mergedObject.GetNamespace()).Update(context.TODO(), mergedObject, metav1.UpdateOptions{})
	return err
}

//DeleteCluster deletes the cluster
func DeleteCluster(ctx context.Context, client dynamic.Interface, v *variables.Variables) error {
	deleteFromTmpl := func(o object) error {
		u, err := loadTextTemplate(o, *v)
		if err != nil {
			return err
		}
		return deleteBytes(ctx, client, o.gvr, u)
	}
	return deleteFromTmpl(object{
		gvr:  gvr.Cluster,
		text: templates.Cluster,
	})
}

//WaitForCAPIClusterReady waits for the CAPI cluster resource to reach "Ready" status
func WaitForCAPIClusterReady(ctx context.Context, client dynamic.Interface, state *variables.Variables) error {
	endTime := time.Now().Add(clusterCreationTimeout)
	for {
		time.Sleep(clusterCreationPollingInterval)
		cluster, err := client.Resource(gvr.Cluster).Namespace(state.Namespace).Get(ctx, state.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if isClusterReady(cluster) {
			return nil
		}
		if time.Now().After(endTime) {
			return fmt.Errorf("timed out waiting for cluster %s to create", state.Name)
		}
	}
}

func isClusterReady(cluster *unstructured.Unstructured) bool {
	clusterStatus := cluster.Object["status"].(map[string]interface{})
	if clusterStatus == nil {
		return false
	}
	controlPlaneReady, ok := clusterStatus["controlPlaneReady"]
	if !ok {
		return false
	}
	controlPlaneReadyBool, ok := controlPlaneReady.(bool)
	if !ok || !controlPlaneReadyBool {
		return false
	}

	infrastructureReady, ok := clusterStatus["infrastructureReady"]
	if !ok {
		return false
	}
	infrastructureReadyBool, ok := infrastructureReady.(bool)
	if !ok || !infrastructureReadyBool {
		return false
	}

	phase, ok := clusterStatus["phase"]
	if !ok {
		return false
	}
	phaseString, ok := phase.(string)
	if !ok || phaseString != "Provisioned" {
		return false
	}
	return true
}

func loadTextTemplate(o object, variables variables.Variables) (*unstructured.Unstructured, error) {
	t, err := template.New(o.gvr.Resource).Parse(o.text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, variables); err != nil {
		return nil, err
	}
	templatedBytes := buf.Bytes()
	u, err := toUnstructured(templatedBytes)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func createIfNotExists(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, u *unstructured.Unstructured) error {
	_, err := client.Resource(gvr).Namespace(u.GetNamespace()).Create(ctx, u, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func deleteBytes(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, u *unstructured.Unstructured) error {
	err := client.Resource(gvr).Namespace(u.GetNamespace()).Delete(ctx, u.GetName(), metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func toUnstructured(o []byte) (*unstructured.Unstructured, error) {
	j, err := apiyaml.ToJSON(o)
	if err != nil {
		return nil, err
	}
	obj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, j)
	if err != nil {
		return nil, err
	}
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("invalid unstructured object")
	}
	return u, nil
}

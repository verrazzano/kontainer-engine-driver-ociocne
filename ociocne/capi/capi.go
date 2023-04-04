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
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strings"
	"text/template"
	"time"
)

const (
	clusterCreationTimeout         = 1 * time.Hour
	clusterCreationPollingInterval = 30 * time.Second
)

// createOrUpdateCAPISecret creates the CAPI secret if it does not already exist
// if the secret exists, update it in place with the new credentials
func createOrUpdateCAPISecret(ctx context.Context, v *variables.Variables, client kubernetes.Interface) error {
	data := map[string][]byte{
		"tenancy":              []byte(v.Tenancy),
		"user":                 []byte(v.User),
		"fingerprint":          []byte(v.Fingerprint),
		"region":               []byte(v.Region),
		"passphrase":           []byte(v.PrivateKeyPassphrase),
		"key":                  []byte(strings.TrimSpace(strings.ReplaceAll(v.PrivateKey, "\\n", "\n"))),
		"useInstancePrincipal": []byte("false"),
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

// CreateOrUpdateAllObjects creates or updates all cluster objects
func CreateOrUpdateAllObjects(ctx context.Context, kubernetesInterface kubernetes.Interface, dynamicInterface dynamic.Interface, v *variables.Variables) error {
	if err := createOrUpdateCAPISecret(ctx, v, kubernetesInterface); err != nil {
		return fmt.Errorf("failed to create CAPI credentials: %v", err)
	}
	for _, o := range createObjects(v) {
		if err := createOrUpdateObject(ctx, dynamicInterface, o, v); err != nil {
			return fmt.Errorf("failed to create object %s/%s/%s: %v", o.gvr.Group, o.gvr.Version, o.gvr.Resource, err)
		}
	}
	return nil
}

// CreateOrUpdateNodeGroup creates or updates the worker node group replica count
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

// DeleteCluster deletes the cluster
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

// WaitForCAPIClusterReady waits for the CAPI cluster resource to reach "Ready" status
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

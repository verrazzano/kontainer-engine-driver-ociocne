// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8s

import (
	"encoding/base64"
	"errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const (
	injectedKubeConfig = "INJECTED_KUBECONFIG"
)

// MustSetKubeconfigFromEnv sets the current kubeconfig from the environment. If the kubeconfig cannot be set, panic.
func MustSetKubeconfigFromEnv() {
	val := os.Getenv(injectedKubeConfig)

	if len(val) < 1 {
		panic(errors.New("injected KubeConfig not found"))
	}

	kc, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		panic(err)
	}

	InjectedKubeConfig = kc
}

var InjectedKubeConfig []byte

// NewInterfaceForKubeconfig creates a kubernetes.Interface given a kubeconfig string
func NewInterfaceForKubeconfig(kubeconfig string) (kubernetes.Interface, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// NewInterface creates a new kubernetes.Interface using the injected kubeconfig
func NewInterface() (kubernetes.Interface, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(InjectedKubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// NewDynamic creates a new dynamic.Interface using the injected kubeconfig
func NewDynamic() (dynamic.Interface, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(InjectedKubeConfig)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

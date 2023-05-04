// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package version

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	configMapName      = "ocne-metadata"
	configMapNamespace = "verrazzano-install"
)

type Defaults struct {
	Calico            string `json:"calico"`
	CoreDNS           string `json:"coredns"`
	ETCD              string `json:"etcd"`
	TigeraOperator    string `json:"tigera-operator" yaml:"tigera-operator"`
	KubernetesVersion string `json:"-"`
}

func GetDefaultsForVersion(ctx context.Context, ki kubernetes.Interface, kubernetesVersion string) (*Defaults, error) {
	versions, err := getVersionMapping(ctx, ki)
	if err != nil {
		return nil, err
	}
	if versions == nil || versions[kubernetesVersion] == nil {
		return nil, fmt.Errorf("no defaults available for version %s", kubernetesVersion)
	}

	return versions[kubernetesVersion], nil
}

func LoadDefaults(ctx context.Context, ki kubernetes.Interface) (*Defaults, error) {
	versions, err := getVersionMapping(ctx, ki)
	if err != nil {
		return nil, err
	}
	if versions == nil {
		return nil, errors.New("no defaults available")
	}

	var kubernetesVersion string
	for k := range versions {
		if len(kubernetesVersion) < 1 || k > kubernetesVersion {
			kubernetesVersion = k
		}
	}

	defaults := versions[kubernetesVersion]
	defaults.KubernetesVersion = kubernetesVersion
	return defaults, nil
}

func getVersionMapping(ctx context.Context, ki kubernetes.Interface) (map[string]*Defaults, error) {
	cm, err := ki.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		// no defaults known
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if cm.Data == nil {
		return nil, nil
	}

	mapping := cm.Data["mapping"]
	if len(mapping) < 1 {
		return nil, nil
	}

	versions := map[string]*Defaults{}
	if err := yaml.Unmarshal([]byte(mapping), &versions); err != nil {
		return nil, err
	}

	if len(versions) < 1 {
		return nil, nil
	}
	return versions, nil
}

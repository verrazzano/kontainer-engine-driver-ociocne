// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package version

import (
	"context"
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
)

const (
	ocneConfigMapName          = "ocne-metadata"
	ocneConfigMapNamespace     = "verrazzano-capi"
	verrazzanoInstallNamespace = "verrazzano-install"
	verrazzanoConfigMapName    = "verrazzano-meta"
)

type Defaults struct {
	Release         string `json:"Release"`
	ContainerImages struct {
		Calico         string `json:"calico"`
		CoreDNS        string `json:"coredns"`
		ETCD           string `json:"etcd"`
		TigeraOperator string `json:"tigera-operator" yaml:"tigera-operator"`
	} `json:"container-images" yaml:"container-images"`
	KubernetesVersion string `json:"-"`
	VerrazzanoVersion string `json:"-"`
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

	verrazzanoVersion, err := getVerrazzanoVersion(ctx, ki)
	if err != nil {
		return nil, err
	}

	defaults := versions[kubernetesVersion]
	defaults.KubernetesVersion = kubernetesVersion
	defaults.VerrazzanoVersion = verrazzanoVersion
	return defaults, nil
}

func getVersionMapping(ctx context.Context, ki kubernetes.Interface) (map[string]*Defaults, error) {
	cm, err := ki.CoreV1().ConfigMaps(ocneConfigMapNamespace).Get(ctx, ocneConfigMapName, metav1.GetOptions{})
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

func getVerrazzanoVersion(ctx context.Context, ki kubernetes.Interface) (string, error) {
	cm, err := ki.CoreV1().ConfigMaps(verrazzanoInstallNamespace).Get(ctx, verrazzanoConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	if cm.Data == nil {
		return "", nil
	}
	verrazzanoVersions := cm.Data["verrazzano-versions"]
	if len(verrazzanoVersions) < 1 {
		return "", nil
	}

	versionMapping := map[string]string{}
	if err := json.Unmarshal([]byte(verrazzanoVersions), &versionMapping); err != nil {
		return "", err
	}

	var versions []string
	for k := range versionMapping {
		versions = append(versions, k)
	}

	sort.Strings(versions)
	return versions[0], nil
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package version

type Properties struct {
	ETCDImageTag    string
	CoreDNSImageTag string
	TigeraTag       string
}

// Mapping has the well-known properties for supported OCNE Kubernetes versions
var Mapping = map[string]Properties{
	"v1.24.8": {
		ETCDImageTag:    "3.5.3",
		CoreDNSImageTag: "1.8.6",
		TigeraTag:       "v1.29.0",
	},
	"v1.25.7": {
		ETCDImageTag:    "3.5.6",
		CoreDNSImageTag: "v1.9.3",
		TigeraTag:       "v1.29.0",
	},
}

// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package version

type Properties struct {
	ETCDImageTag    string
	CoreDNSImageTag string
	CalicoTag       string
}

var Mapping = map[string]Properties{
	"v1.24.8": {
		ETCDImageTag:    "3.5.3",
		CoreDNSImageTag: "1.8.6",
		CalicoTag:       "v3.25.0",
	},
	"v1.25.7": {
		ETCDImageTag:    "3.5.6",
		CoreDNSImageTag: "1.9.3",
		CalicoTag:       "v3.25.0",
	},
}

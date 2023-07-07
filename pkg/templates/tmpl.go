// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package templates

import _ "embed"

//go:embed cluster.goyaml
var Cluster string

//go:embed ocicluster.goyaml
var OCICluster string

//go:embed ociclusteridentity.goyaml
var ClusterIdentity string

//go:embed ocnecontrolplane.goyaml
var OCNEControlPlane string

//go:embed ocicontrolplanemachinetemplate.goyaml
var OCIControlPlaneMachineTemplate string

//go:embed ocneconfig.goyaml
var OCNEConfigTemplate string

//go:embed machinedeployment.goyaml
var MachineDeployment string

//go:embed ocimachinetemplate.goyaml
var OCIMachineTemplate string

//go:embed ccmresourceset.goyaml
var CCMResourceSet string

//go:embed csiresourceset.goyaml
var CSIResourceSet string

//go:embed ccmsecret.goyaml
var CCMConfigMap string

//go:embed csisecret.goyaml
var CSIConfigMap string

//go:embed ccm-module.goyaml
var CCMModule string

//go:embed calico-module.goyaml
var CalicoModule string

//go:embed vmc.goyaml
var VMC string

//go:embed provisioner.goyaml
var ProvisionerCM string

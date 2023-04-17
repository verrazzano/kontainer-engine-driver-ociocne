#!/bin/bash
# Copyright (c) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

set -eu

if [[ -z $DRIVER_URL ]]; then
  echo "Missing DRIVER_URL env variable"
  exit 1
fi

CHECKSUM=$(shasum -a 256 dist/kontainer-engine-driver-ociocne-linux |awk '{print $1}')

echo "apiVersion: management.cattle.io/v3
kind: KontainerDriver
metadata:
  name: ociocneengine
spec:
  active: true
  builtIn: false
  uiUrl: ''
  url: ${DRIVER_URL}
  checksum: ${CHECKSUM}" > dist/kontainerdriver.yaml

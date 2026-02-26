#!/bin/bash

echo Updating Viam app module...
viam module update --module=viam-app-meta.json
echo Done

module_id=$(jq -r '.module_id' viam-app-meta.json)
org_namespace=${module_id%%:*}
echo "Organization namespace is: $org_namespace"

date=$(date +%Y%m%d%H%M%S)
version="0.0.0-${date}"
echo "Version is: $version"

viam module upload --module=viam-app-meta.json --version=${version} --platform=any --public-namespace=${org_namespace} dist
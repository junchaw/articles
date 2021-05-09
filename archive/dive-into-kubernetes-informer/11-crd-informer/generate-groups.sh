#!/bin/bash

scriptPath=${SCRIPT_PATH}

if [ -z ${scriptPath} ]; then
  echo "Tell me, where is the file? (SCRIPT_PATH)"
  echo "Maybe 'export SCRIPT_PATH=~/projects/code-generator/generate-groups.sh'?"
  exit 1
fi

${scriptPath} \
  all \
  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client \
  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/api \
  stable.wbsnail.com:v1 \
  --go-header-file ./boilerplate.go.txt

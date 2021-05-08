#!/bin/bash

~/projects/code-generator/generate-groups.sh \
  all \
  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client \
  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/apis \
  stable.wbsnail.com:v1 \
  --go-header-file ./boilerplate.go.txt

#!/bin/bash

set -e
set -x

unset GOPATH

#deepcopy-gen \
#  --input-dirs ./apis/stable.wbsnail.com/v1 \
#  -O zz_generated.deepcopy \
#  --go-header-file ./boilerplate.go.txt

client-gen \
  --clientset-name versioned \
  --input-base '' \
  --input github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/apis/stable.wbsnail.com/v1 \
  --output-package github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/clientset \
  --go-header-file ./boilerplate.go.txt

#lister-gen \
#  --input-dirs github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/apis/stable.wbsnail.com/v1 \
#  --output-package ./client/listers \
#  --go-header-file ./boilerplate.go.txt

#informer-gen \
#  --input-dirs github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/apis/stable.wbsnail.com/v1 \
#  --versioned-clientset-package github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/clientset/versioned \
#  --listers-package github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/listers \
#  --output-package github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/informers \
#  --go-header-file ./boilerplate.go.txt
#
#
#~/projects/code-generator/generate-groups.sh \
#  all \
#  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client \
#  github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/apis \
#  stable.wbsnail.com:v1 \
#  --go-header-file ./boilerplate.go.txt

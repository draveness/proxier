SHELL=/bin/bash -o pipefail

FIRST_GOPATH:=$(firstword $(subst :, ,$(shell go env GOPATH)))
GO_PKG=github.com/draveness/proxier
K8S_GEN_BINARIES:=deepcopy-gen informer-gen lister-gen client-gen

TYPES_V1BETA1_TARGET:=pkg/apis/maegus/v1beta1/proxier_types.go

K8S_GEN_DEPS:=.header
K8S_GEN_DEPS+=$(TYPES_V1BETA1_TARGET)
K8S_GEN_DEPS+=$(foreach bin,$(K8S_GEN_BINARIES),$(FIRST_GOPATH)/bin/$(bin))
K8S_GEN_DEPS+=$(OPENAPI_GEN_BINARY)

OPERATOR_E2E_IMAGE_TAG:=$(shell git rev-parse --short HEAD)
OPERATOR_E2E_IMAGE_NAME:=draveness/proxier-e2e:$(OPERATOR_E2E_IMAGE_TAG)

.PHONY: test
test:
	go test -count=1 ./pkg/...

e2e:
	./hack/docker-image-exists.sh || \
	(operator-sdk build $(OPERATOR_E2E_IMAGE_NAME) && docker push $(OPERATOR_E2E_IMAGE_NAME))
	go test -v ./test/e2e/ --kubeconfig "$(HOME)/.kube/k8s-playground-kubeconfig.yaml" --operator-image $(OPERATOR_E2E_IMAGE_NAME)

start:
	operator-sdk up local --namespace=default

release:
	./hack/make-release.sh

LISTER_TARGET := pkg/client/listers/maegus/v1beta1/proxier.go
$(LISTER_TARGET): $(K8S_GEN_DEPS)
	$(LISTER_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--input-dirs     "$(GO_PKG)/pkg/apis/maegus/v1beta1" \
	--output-package "$(GO_PKG)/pkg/client/listers"

CLIENT_TARGET := pkg/client/versioned/clientset.go
$(CLIENT_TARGET): $(K8S_GEN_DEPS)
	$(CLIENT_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--input-base     "" \
	--clientset-name "versioned" \
	--input	         "$(GO_PKG)/pkg/apis/maegus/v1beta1" \
	--output-package "$(GO_PKG)/pkg/client"

INFORMER_TARGET := pkg/client/informers/externalversions/maegus/v1beta1/proxier.go
$(INFORMER_TARGET): $(K8S_GEN_DEPS) $(LISTER_TARGET) $(CLIENT_TARGET)
	$(INFORMER_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--versioned-clientset-package "$(GO_PKG)/pkg/client/versioned" \
	--listers-package "$(GO_PKG)/pkg/client/listers" \
	--input-dirs      "$(GO_PKG)/pkg/apis/maegus/v1beta1" \
	--output-package  "$(GO_PKG)/pkg/client/informers"

.PHONY: k8s-gen
k8s-gen: \
  $(CLIENT_TARGET) \
  $(LISTER_TARGET) \
  $(INFORMER_TARGET)


define _K8S_GEN_VAR_TARGET_
$(shell echo $(1) | tr '[:lower:]' '[:upper:]' | tr '-' '_')_BINARY:=$(FIRST_GOPATH)/bin/$(1)

$(FIRST_GOPATH)/bin/$(1):
	go get -u -d k8s.io/code-generator/cmd/$(1)
	cd $(FIRST_GOPATH)/src/k8s.io/code-generator; git checkout $(K8S_GEN_VERSION)
	go install k8s.io/code-generator/cmd/$(1)

endef

$(foreach binary,$(K8S_GEN_BINARIES),$(eval $(call _K8S_GEN_VAR_TARGET_,$(binary))))

include build/makelib/common.mk
include build/makelib/plugin.mk

# Container image
TAG ?= latest
IMG ?= quay.io/validator-labs/validator-plugin-vsphere:$(TAG)

# Helm vars
CHART_NAME=validator-plugin-vsphere

.PHONY: dev
dev: ## Run a controller via devspace
	devspace dev -n validator-plugin-vsphere-system

# Static Analysis / CI

chartCrds = chart/validator-plugin-vsphere/crds/validation.spectrocloud.labs_vspherevalidators.yaml

reviewable-ext:
	rm $(chartCrds)
	cp config/crd/bases/validation.spectrocloud.labs_vspherevalidators.yaml $(chartCrds)

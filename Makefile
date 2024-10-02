# set env var
.EXPORT_ALL_VARIABLES:
SHELL:=/bin/bash
# Variables
########################################################
# Unit test via mock                           		  #
########################################################
.PHONY: unit-test
test:
	go test -v -count=1 -cover -run ^TestMock ./...
########################################################
# Integration test                        			  #
########################################################
.PHONY: integration-test-tg
test-tg:
	go test -v  -timeout 10m  -run ^TestTerragrunt$  ./pkg/terragrunt
########################################################
# Lint                                                #
########################################################
.PHONY: go-lint
go-lint:
	golangci-lint run
.PHONY: default clean build fmt lint vet cyclo ineffassign shellcheck errcheck goconst gosec abcgo style run test cover license before_commit help godoc install_docgo install_addlicense

SOURCES:=$(shell find . -name '*.go')
BINARY:=insights-content-service
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean
	rm -f rest-api-tests

build: ${BINARY}

${BINARY}: ${SOURCES}
	./build.sh

checker/checker: checker/main.go
	cd checker
	go build
	cd ..

fmt: ## Run go fmt -w for all sources
	@echo "Running go formatting"
	./gofmt.sh

lint: ## Run golint
	@echo "Running go lint"
	./golint.sh

vet: ## Run go vet. Report likely mistakes in source code
	@echo "Running go vet"
	./govet.sh

cyclo: ## Run gocyclo
	@echo "Running gocyclo"
	./gocyclo.sh

ineffassign: ## Run ineffassign checker
	@echo "Running ineffassign checker"
	./ineffassign.sh

shellcheck: ## Run shellcheck
	shellcheck $(shell find . -name "*.sh")

errcheck: ## Run errcheck
	@echo "Running errcheck"
	./goerrcheck.sh

goconst: ## Run goconst checker
	@echo "Running goconst checker"
	./goconst.sh

gosec: ## Run gosec checker
	@echo "Running gosec checker"
	./gosec.sh

abcgo: ## Run ABC metrics checker
	@echo "Run ABC metrics checker"
	./abcgo.sh

style: fmt vet lint cyclo shellcheck errcheck goconst gosec ineffassign abcgo ## Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts

run: clean build ## Build the project and executes the binary
	./insights-content-service

test: clean build ## Run the unit tests
	@go test -coverprofile coverage.out $(shell go list ./... | grep -v tests)

cover: test
	@go tool cover -html=coverage.out

integration_tests: ## Run all integration tests
	@echo "Running all integration tests"
	@./test.sh

license: install_addlicense
	addlicense -c "Red Hat, Inc" -l "apache" -v ./

before_commit: style test license
	./check_coverage.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

docs/packages/%.html: %.go
	mkdir -p $(dir $@)
	docgo -outdir $(dir $@) $^
	addlicense -c "Red Hat, Inc" -l "apache" -v $@

godoc: install_docgo install_addlicense ${DOCFILES}

install_docgo:
	[[ `command -v docgo` ]] || GO111MODULE=off go get -u github.com/dhconnelly/docgo

install_addlicense:
	[[ `command -v addlicense` ]] || GO111MODULE=off go get -u github.com/google/addlicense

SOURCE_DIRS      = cmd pkg
SOURCES          := $(shell find . -name '*.go' -not -path "*/vendor/*")
.DEFAULT_GOAL    := apb

apb: $(SOURCES) ## Build the samplebroker
	go build -i -ldflags="-s -w"

install:
	go install -ldflags="-s -w"

uninstall: clean
	@rm -f ${GOPATH}/bin/apb

lint: ## Run golint
	@golint -set_exit_status $(addsuffix /... , $(SOURCE_DIRS))

fmtcheck: ## Check go formatting
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

test: ## Run unit tests
	@go test -cover ./cmd/... ./pkg/...

vet: ## Run go vet
	@go tool vet ./cmd ./pkg

check: fmtcheck vet lint apb test ## Pre-flight checks before creating PR

clean: ## Clean up your working environment
	@rm -f apb

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: clean lint build fmtcheck test vet help install uninstall

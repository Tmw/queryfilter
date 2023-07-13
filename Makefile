GOTESTSUM_PATH ?= $(shell which gotestsum)
COVERAGE_FILENAME ?= ".coverage"

.PHONY: test
test:
	$(if $(GOTESTSUM_PATH), @gotestsum --, @go test) -v -race -count=1 ./...

.PHONY: lint
lint:
	@golangci-lint run -c .golangci.yml

.PHONY: coverage
coverage:
	@go test -coverprofile $(COVERAGE_FILENAME).out ./...
	@go tool cover -html=coverage.out
	@echo "generated coverage: file://$$(pwd)/$(coverage_filename).html"

.PHONY: watch
watch:
ifeq ($(GOTESTSUM_PATH), )
	@echo "watch requires gotestsum to be installed"
	@exit 1
endif
	@gotestsum --watch

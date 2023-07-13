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


# examples
.PHONY: example-sqlite
example-sqlite:
	@pushd examples/sqlite/ \
		&& go run main.go \
		&& popd

.PHONY: example-mysql
example-mysql:
# start docker-compose and wait for mysql to be ready
	@cd examples/mysql/ \
		&& docker-compose up -d db --wait

# run the example
	@cd examples/mysql/ \
		&& go run main.go \

# stop docker-compose
	@cd examples/mysql/ \
		&& docker-compose down

.PHONY: example-postgresql
example-postgres:
# start docker-compose and wait for postgresql to be ready
	@cd examples/postgresql/ \
		&& docker-compose up -d db --wait

# run the example
	@cd examples/postgresql/ \
		&& go run main.go \

# stop docker-compose
	@cd examples/postgresql/ \
		&& docker-compose down

GOTESTSUM_PATH ?= $(shell which gotestsum)
COVERAGE_FILENAME ?= "coverage.out"
HTML_FILENAME := $(subst .out,.html,$(COVERAGE_FILENAME))

.PHONY: test
test:
	$(if $(GOTESTSUM_PATH), @gotestsum --, @go test) -v -race -count=1 ./...

.PHONY: lint
lint:
	@golangci-lint run -c .golangci.yml

coverage:
	@go test -coverprofile $(COVERAGE_FILENAME) ./...

.PHONY: coverage-html
coverage-html: coverage
	@go tool cover -html=$(COVERAGE_FILENAME)
	@echo "generated coverage: file://$$(pwd)/$(HTML_FILENAME)"

.PHONY: coverage-term
coverage-term: coverage
	gocovsh

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

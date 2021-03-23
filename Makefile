ifeq ($(CI),)
    # Not in CI
    VERSION := dev-$(USER)
    VERSION_HASH := $(shell git rev-parse HEAD | cut -b1-8)
else
    # In CI
    ifneq ($(RELEASE_VERSION),)
        VERSION := $(RELEASE_VERSION)
    else
        # No tag, so make one
        VERSION := $(shell git describe --tags 2>/dev/null)
    endif
    VERSION_HASH := $(shell echo "$(GITHUB_SHA)" | cut -b1-8)
endif

GO_FLAGS := -ldflags "-X main.Version=$(VERSION) -X main.VersionHash=$(VERSION_HASH)"

.PHONY: deps
deps:
	@go get -t -d ./...

.PHONY: build
build:
	@mkdir -p build
	@go build $(GO_FLAGS) -o ./build/vegatools ./

.PHONY: lint
lint:
	@go install golang.org/x/lint/golint
	@outputfile="$$(mktemp)" ; \
	go list ./... | xargs golint 2>&1 | \
		sed -e "s#^$$GOPATH/src/##" | tee "$$outputfile" ; \
	lines="$$(wc -l <"$$outputfile")" ; \
	rm -f "$$outputfile" ; \
	exit "$$lines"

.PHONY: test
test: ## Run tests
	@go test ./...

.PHONY: coverage
coverage:
	@go test -covermode=count -coverprofile="coverage.txt" ./...
	@go tool cover -func="coverage.txt"

.PHONY: vet
vet: deps
	@go vet ./...

.PHONY: clean
clean:
	@rm -rf ./build
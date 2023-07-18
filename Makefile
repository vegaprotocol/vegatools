
.PHONY: deps
deps:
	@go get -t -d ./...

.PHONY: build
build:
	@mkdir -p build
	@go build -o ./build/vegatools ./

.PHONY: lint
lint:
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


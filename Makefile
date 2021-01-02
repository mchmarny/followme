APP_NAME         ?=followme
APP_VERSION      ?=v0.3.11

.PHONY: all
all: help

.PHONY: tidy 
tidy: ## Updates go modules and vendors deps
	go mod tidy
	go mod vendor

.PHONY: test 
test: tidy ## Tests the entire project 
	go test -count=1 -race -covermode=atomic -coverprofile=cover.out ./...

.PHONY: static 
static: ## Makes static content into binary data
	go get github.com/go-bindata/go-bindata/...
	go-bindata -o internal/app/static.go -pkg app -fs -prefix "web/static" web/...
	go mod tidy
	
.PHONY: build 
build: tidy ## Builds app locally (/bin)
	CGO_ENABLED=0 go build \
	-ldflags "-X main.Version=$(APP_VERSION)" \
	-mod vendor -o bin/$(APP_NAME) ./cmd/

.PHONY: app 
app: ## Runs compiled app
	bin/$(APP_NAME) app

.PHONY: worker 
worker: ## Runs compiled worker
	bin/$(APP_NAME) worker

.PHONY: spell 
spell: ## Checks spelling across the entire project 
	go get github.com/client9/misspell/cmd/misspell
	go mod tidy
	misspell -locale US cmd/**/* internal/**/* pkg/**/* web/**/* README.md

.PHONY: lint 
lint: ## Lints the entire project
	# brew install golangci-lint
	golangci-lint run --timeout=3m
		
.PHONY: tag 
tag: ## Creates release tag 
	git tag $(APP_VERSION)
	git push origin $(APP_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



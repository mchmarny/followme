APP_NAME         ?=followme
APP_VERSION      ?=v0.1.1

.PHONY: all
all: help

.PHONY: tidy 
tidy: ## Updates go modules and vendors deps
	go mod tidy
	go mod vendor

.PHONY: test 
test: tidy ## Tests the entire project 
	TWITTER_KEY=$(TW_CONSUMER_KEY) \
	TWITTER_SECRET=$(TW_CONSUMER_SECRET) \
	RELEASE=$(RELEASE_VERSION) \
	go test -count=1 -race -covermode=atomic -coverprofile=cover.out ./...

.PHONY: build 
build: tidy ## Builds app locally (/bin)
	go build -a -ldflags "-w -extldflags -static" -mod vendor -o bin/$(APP_NAME) ./cmd/

.PHONY: app 
app: ## Runs compiled app
	bin/$(APP_NAME) app -k $(TW_CONSUMER_KEY) -s $(TW_CONSUMER_SECRET) -p 8080

.PHONY: worker 
worker: ## Runs compiled worker
	bin/$(APP_NAME) worker -k $(TW_CONSUMER_KEY) -s $(TW_CONSUMER_SECRET)

.PHONY: call 
call: ## Runs pre-compiled app 
	curl -i -X POST -H "content-type: application/json" -H "token: $(TEST_API_TOKEN)" \
		 http://127.0.0.1:8080/api/v1.0/update

.PHONY: spell 
spell: ## Checks spelling across the entire project 
	go get github.com/client9/misspell/cmd/misspell
	misspell -locale US -error cmd/**/* build/**/* pkg/**/* tools/**/* web/**/* *.md

.PHONY: lint 
lint: ## Lints the entire project
	# brew install golangci-lint
	golangci-lint run --timeout=3m
		
.PHONY: tag 
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



# Project settings
PROJECT_NAME=Recho
PKG=./...
MAIN=main.go
BUILD_DIR=.

# Git info
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Docker setting
TEST_TAG=dev
DOCKER_FILE_PATH=.

run: ## Run the app
	@clear
	@go run $(MAIN)

compile: ## Compile (without installing)
	@go build -o $(BUILD_DIR)/$(PROJECT_NAME) $(MAIN)

build: clean compile ## Build the project 
	./$(BUILD_DIR)/$(PROJECT_NAME)

test: ## Run tests
	go test $(PKG)

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)/$(PROJECT_NAME)

push: ## Push current branch to origin
	git push origin $(BRANCH)

pull: ## Pull current branch from origin
	git pull origin $(BRANCH)

image: ## build an image from docker file
	docker build -t $(PROJECT_NAME):$(TEST_TAG) $(DOCKER_FILE_PATH) 

drun: ## make dokcer run the image
	docker run --name $(PROJECT_NAME)_$(TEST_TAG) --env-file .env -p 8080:8000  $(PROJECT_NAME):$(TEST_TAG)

dstop: ## delete and stop continer
	docker stop $(PROJECT_NAME)_$(TEST_TAG)
	docker rm $(PROJECT_NAME)_$(TEST_TAG)

swag: ## make swagger api's docs
	swag init -g internal/web/echo.go
help: ## Show this help
	@echo "Usage: make [target]"
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
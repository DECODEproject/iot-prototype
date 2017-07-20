
build: build-node build-search ## builds all of the elm projects

.PHONY: build

build-search: ## builds the elm source code for the search page
	elm make src/Search.elm --output=search.js

.PHONY: build-search

build-node: ## builds the elm source code for the node application
	elm make src/Node.elm --output=node.js

.PHONY: build-node

format: ## formats the elm source code
	elm-format --yes src

.PHONY: format

# 'help' parses the Makefile and displays the help text
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help

build: ## builds the elm source code
	elm make src/Main.elm --output=main.js

.PHONY: build

format: ## formats the elm source code
	elm-format --yes src

.PHONY: format

# 'help' parses the Makefile and displays the help text
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help

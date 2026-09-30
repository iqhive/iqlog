GO ?= go

.PHONY: hero hero-svg
hero: ## Play the README hero in an 80×20 or larger terminal
	env GOWORK=off $(GO) -C examples/hero run .

hero-svg: ## Regenerate the self-contained README animation
	env GOWORK=off $(GO) -C examples/hero run . -svg ../../docs/hero.svg

.PHONY: hero-check
hero-check: ## Check real fixtures and generated hero freshness
	env GOWORK=off $(GO) -C examples/hero test ./...
